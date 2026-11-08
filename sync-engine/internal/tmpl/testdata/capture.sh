#!/usr/bin/env bash
# Writes the golden files of the template test from the bash engine and the
# gomplate CLI. Run it from the root of the repository:
#   UPDATE_GOLDEN=1 go test ./internal/tmpl    (it calls this script)
# or by hand: sync-engine/internal/tmpl/testdata/capture.sh <file.tar.gz>
#
# It runs `repo_sync --dry-run` for each example values file (and for a values
# file with each tool), and `setup_sync` for the Mac example and the fixture
# configs. A wrapper of `_se_gomplate` saves, for each render of a file
# (`gomplate -f`): the source path, the JSON context, the flag --missing-key and
# the rendered output. The scratch folder is written as @TMP@ in the context
# and in the output.
set -uo pipefail

root=$(pwd -P)
final=$1
tmp=$(mktemp -d "${TMPDIR:-/tmp}/tmpl-capture.XXXXXX")
tmp=$(cd "$tmp" && pwd -P)
trap 'rm -rf -- "$tmp"' EXIT
out="$tmp/out"
mkdir -p -- "$out"

unset GIT_DIR GIT_INDEX_FILE GIT_WORK_TREE GIT_OBJECT_DIRECTORY \
  GIT_ALTERNATE_OBJECT_DIRECTORIES GIT_COMMON_DIR GIT_PREFIX GIT_EXEC_PATH \
  GIT_CONFIG GIT_CONFIG_PARAMETERS GIT_CONFIG_COUNT \
  "${!GIT_CONFIG_KEY_@}" "${!GIT_CONFIG_VALUE_@}"
git_vars=$(git rev-parse --local-env-vars 2> /dev/null) || git_vars=''
# shellcheck disable=SC2086
unset $git_vars
unset git_vars

_git() {
  GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git -c user.name=test -c user.email=test@example.com \
    -c commit.gpgsign=false -c core.hooksPath=/dev/null -c init.defaultBranch=main "$@"
}

# The wrapper, sourced after the library.
cat > "$tmp/wrap.sh" << 'WRAP'
# A render of a file (-f) saves: src, missingkey, rc, ctx.json, out.
# A render of an inline template (-i) saves: inline, missingkey, rc, env.json,
# ctx/<name>.<ext> (each -c file), out, err.
# The engine renders from its own copy of the clone, in $TMPDIR/repo-sync.XXXXXX/.
_cap_scrub() { TMPV=$CAP_TMP perl -0777 -pe 's/\Q$ENV{TMPV}\E/\@TMP\@/g; s{[^"\s:]*\/repo-sync\.\w+\/}{\@TMP\@/bootstrap/}g'; }
_se_gomplate() {
  local a=("$@") i c ctx='' src='' outf='' mk='' inl='' rc=0 d rel n name path
  local names=() paths=()
  for ((i = 0; i < ${#a[@]}; i++)); do
    case ${a[i]} in
      -c)
        c=${a[i + 1]}
        case $c in
          .=*) ctx=${c#.=}; ctx=${ctx%%\?*} ;;
          *=*)
            name=${c%%=*}; path=${c#*=}; path=${path%%\?*}
            names+=("$name"); paths+=("$path")
            ;;
        esac
        ;;
      -f) src=${a[i + 1]} ;;
      -i) inl=${a[i + 1]} ;;
      -o) outf=${a[i + 1]} ;;
      --missing-key) mk=${a[i + 1]} ;;
    esac
  done
  if [[ -n $src && -n $ctx && -n $outf ]]; then
    _se_tool gomplate --left-delim '<%' --right-delim '%>' "$@" || rc=$?
    rel=${src#"$CAP_TREE"/}
    # The engine renders the copy of the clone in its own temp folder.
    case $rel in
      /*/repo/tools/*) rel=repo/tools/${rel##*/repo/tools/} ;;
      /*/macos/tools/*) rel=macos/tools/${rel##*/macos/tools/} ;;
    esac
    case $rel in "$CAP_FROM"*) rel=$CAP_TO${rel#"$CAP_FROM"} ;; esac
    d=$(mktemp -d "$CAP_OUT/c.XXXXXX")
    printf '%s' "$rel" > "$d/src"
    printf '%s' "$mk" > "$d/missingkey"
    printf '%s' "$rc" > "$d/rc"
    _cap_scrub < "$ctx" > "$d/ctx.json"
    if ((rc == 0)); then
      _cap_scrub < "$outf" > "$d/out"
    fi
    return "$rc"
  fi
  if [[ -n $inl && -z $src && -d ${CAP_INLINE:-} ]]; then
    d=$(mktemp -d "$CAP_INLINE/i.XXXXXX")
    printf '%s' "$inl" | _cap_scrub > "$d/inline"
    printf '%s' "$mk" > "$d/missingkey"
    mkdir -p -- "$d/ctx"
    for ((n = 0; n < ${#names[@]}; n++)); do
      if [[ -f ${paths[n]} ]]; then
        _cap_scrub < "${paths[n]}" > "$d/ctx/${names[n]}.${paths[n]##*.}"
      fi
    done
    if [[ -n $ctx && -f $ctx ]]; then _cap_scrub < "$ctx" > "$d/ctx/..json"; fi
    perl -MJSON::PP -e 'my %e = map { $_ => $ENV{$_} } grep { /^(REPO_SYNC_|SYNC_ENGINE_MISE_|SS_|SE_VALUES_)/ } keys %ENV; print JSON::PP->new->canonical->encode(\%e)' | _cap_scrub > "$d/env.json"
    _se_tool gomplate --left-delim '<%' --right-delim '%>' "$@" > "$d/out.raw" 2> "$d/err.raw" || rc=$?
    printf '%s' "$rc" > "$d/rc"
    if [[ -n $outf && -f $outf ]]; then
      _cap_scrub < "$outf" > "$d/out"
    else
      _cap_scrub < "$d/out.raw" > "$d/out"
    fi
    cat -- "$d/out.raw"
    cat -- "$d/err.raw" >&2
    rm -f -- "$d/out.raw" "$d/err.raw.tmp"
    _cap_scrub < "$d/err.raw" > "$d/err"
    rm -f -- "$d/err.raw"
    return "$rc"
  fi
  _se_tool gomplate --left-delim '<%' --right-delim '%>' "$@"
}
WRAP

# The bootstrap tree: the working tree's repo/, macos/, sync-engine/ and tools/.
tree="$tmp/bootstrap"
mkdir -p -- "$tree"
cp -R -- "$root/repo" "$tree/repo"
cp -R -- "$root/macos" "$tree/macos"
cp -R -- "$root/sync-engine" "$tree/sync-engine"
cp -R -- "$root/tools" "$tree/tools"
_git -C "$tree" init --quiet
_git -C "$tree" add -A
_git -C "$tree" commit --quiet --no-verify -m "test copy"

export CAP_TREE=$tree CAP_TMP=$tmp CAP_FROM='' CAP_TO=''
# The real mise folders, found before a case changes HOME and XDG_*, so that
# `mise x <tool>` finds the installed tools and downloads nothing.
export MISE_DATA_DIR="${MISE_DATA_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/mise}"
export MISE_CACHE_DIR="${MISE_CACHE_DIR:-${XDG_CACHE_HOME:-$HOME/.cache}/mise}"
export MISE_OFFLINE=1
export REPO_SYNC_NO_AUTOUPDATE=1
unset SYNC_ENGINE_URL SYNC_ENGINE_CLONE REPO_SYNC_CLONE MISE_ENV
n=0

# ---- the repo side ----
repo_case() {
  local values=$1 name=$2 consumer
  [[ -z ${CAP_ONLY:-} || " $CAP_ONLY " == *" repo-$name "* ]] || return 0
  n=$((n + 1))
  consumer="$tmp/consumer-$n"
  mkdir -p -- "$consumer/.config"
  _git -C "$consumer" init --quiet -b main
  _git -C "$consumer" commit --quiet --no-verify --allow-empty -m "init"
  cp -- "$values" "$consumer/.config/bootstrap.toml"
  local mname mtype mfolder
  while IFS=$'\t' read -r mname mtype mfolder; do
    [[ -n $mname ]] || continue
    mfolder=${mfolder:-$mname}
    if [[ $mtype == submodule ]]; then
      _git -C "$consumer" -c protocol.file.allow=always submodule --quiet add -- "$tree" "$mfolder"
    else
      mkdir -p -- "$consumer/$mfolder"
    fi
  done < <(cd / && mise x yq -- yq -p toml -o tsv '.member // {} | to_entries | .[] | [.key, .value.type // "", .value.folder // ""]' "$consumer/.config/bootstrap.toml")
  (
    export CAP_OUT="$out/repo-$name" CAP_INLINE="$out/inline-repo-$name"
    mkdir -p -- "$CAP_OUT" "$CAP_INLINE"
    export REPO_SYNC_VALUES="$consumer/.config/bootstrap.toml" MISE_PROJECT_ROOT=$consumer SYNC_ENGINE_CLONE=$tree
    cd "$consumer" || exit 1
    # shellcheck source=/dev/null
    source "$tree/sync-engine/repo-sync"
    # shellcheck source=/dev/null
    source "$tmp/wrap.sh"
    repo_sync --dry-run > "$tmp/log-repo-$name" 2>&1
    rc=$?
    echo "repo $name: exit $rc"
    ((rc == 0)) || tail -n 4 "$tmp/log-repo-$name" | sed 's/^/    /' >&2
  )
}

while IFS= read -r ex; do
  repo_case "$root/repo/examples/$ex" "$(echo "${ex%/bootstrap.toml}" | tr '/' '_')"
done < <(cd "$root/repo/examples" && find . -name bootstrap.toml | sed 's#^\./##' | LC_ALL=C sort)

# A values file with each tool of repo/tools/.
{
  printf '%s\n' 'repo_name = "all"' 'repo_url = "https://example.com/all"'
  printf '%s\n' 'environments = ["dev", "ci"]'
  (cd "$root/repo/tools" && ls) | sed 's/.*/[tools.&]/'
} > "$tmp/all.toml"
repo_case "$tmp/all.toml" all

# ---- the Mac side ----
mac_case() {
  local cfg=$1 name=$2 macos=$3 from=$4 to=$5 dir home
  [[ -z ${CAP_ONLY:-} || " $CAP_ONLY " == *" mac-$name "* ]] || return 0
  n=$((n + 1))
  dir="$tmp/mac-$n"
  home="$dir/home"
  mkdir -p -- "$home" "$dir/xdg/cache" "$dir/xdg/config" "$dir/xdg/state" "$dir/xdg/data" "$dir/tmp"
  cp -R -- "$cfg" "$dir/config"
  # The seed file of the core fixture needs this file.
  mkdir -p -- "$home/.config" && : > "$home/.config/core-needed"
  (
    export HOME=$home USER=tester
    export XDG_CACHE_HOME="$dir/xdg/cache" XDG_CONFIG_HOME="$dir/xdg/config"
    export XDG_STATE_HOME="$dir/xdg/state" XDG_DATA_HOME="$dir/xdg/data"
    export SYNC_ENGINE_CACHE="$XDG_CACHE_HOME/sync-engine"
    export MACOS_SETUP_CONFIG_PATH="$dir/config"
    export GIT_CONFIG_GLOBAL="$dir/gitconfig" GIT_CONFIG_NOSYSTEM=1
    : > "$GIT_CONFIG_GLOBAL"
    export TMPDIR="$dir/tmp"
    export CAP_OUT="$out/mac-$name" CAP_INLINE="$out/inline-mac-$name" CAP_FROM=$from CAP_TO=$to
    # The macos/ folder of this case.
    rm -rf -- "$tmp/mtree" && mkdir -p -- "$tmp/mtree"
    cp -R -- "$tree/sync-engine" "$tmp/mtree/sync-engine"
    mkdir -p -- "$tmp/mtree/macos"
    cp -R -- "$macos/." "$tmp/mtree/macos"
    rm -rf -- "$tmp/mtree/macos/tools/bad"
    _git -C "$tmp/mtree" init --quiet -b main
    _git -C "$tmp/mtree" add -A
    _git -C "$tmp/mtree" commit --quiet --no-verify -m "tree"
    export CAP_TREE="$tmp/mtree" SYNC_ENGINE_CLONE="$tmp/mtree"
    mkdir -p -- "$CAP_OUT" "$CAP_INLINE"
    cd "$dir" || exit 1
    # shellcheck source=/dev/null
    source "$tmp/mtree/sync-engine/setup-sync"
    # shellcheck source=/dev/null
    source "$tmp/wrap.sh"
    # The first run can stop at the init stage (exit 4), so run it again.
    setup_sync --no-bootstrap > "$tmp/log-mac-$name" 2>&1
    rc=$?
    if ((rc == 4)); then
      setup_sync --no-bootstrap > "$tmp/log-mac-$name" 2>&1
      rc=$?
    fi
    echo "mac $name: exit $rc"
    ((rc == 0)) || tail -n 4 "$tmp/log-mac-$name" | sed 's/^/    /' >&2
  )
}

mkdir -p -- "$tmp/ex-mac"
cp -- "$root/macos/examples/macos.toml" "$tmp/ex-mac/macos.toml"
mac_case "$tmp/ex-mac" example "$tree/macos" '' ''
for cfg in "$root"/macos/fixtures/config/*/; do
  cfg=${cfg%/}
  mac_case "$cfg" "fixture-${cfg##*/}" "$root/macos/fixtures/repo" 'macos/' 'macos/fixtures/repo/'
done

# Remove a capture that has the same inputs as an earlier one (any case). Keep
# at most 8 renders of each template file and 25 of each inline template, to
# keep the archive small.
seen=' '
counts=''
for c in "$out"/*/c.* "$out"/*/i.*; do
  [[ -d $c ]] || continue
  key=$(cd "$c" && find . -type f ! -name out ! -name err | LC_ALL=C sort | while read -r f; do
    echo "$f"
    cat -- "$f"
  done | cksum)
  key=${key// /_}
  case "$seen" in
    *" $key "*)
      rm -rf -- "$c"
      continue
      ;;
  esac
  seen+="$key "
  if [[ -f $c/src ]]; then
    group=src:$(cat -- "$c/src")
    limit=8
  else
    group=inline:$(cksum < "$c/inline")
    limit=25
  fi
  group=${group// /_}
  n=$(printf '%s\n' "$counts" | grep -cxF -- "$group" || true)
  if ((n >= limit)); then
    rm -rf -- "$c"
  else
    counts+="$group"$'\n'
  fi
done
find "$out" -mindepth 1 -maxdepth 1 -type d -empty -delete

# Pack the captures: sorted, with no file times in the gzip header.
(cd "$out" && find . -type f | LC_ALL=C sort | COPYFILE_DISABLE=1 tar -cf - -T -) | gzip -n -9 > "$final.new" && mv -- "$final.new" "$final"
