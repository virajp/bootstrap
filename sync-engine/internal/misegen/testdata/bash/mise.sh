#!/usr/bin/env bash
# Usage: mise.sh <engine-dir> <case-dir>
# Runs the functions of sync-engine/mise for one case, in the order of the
# front-ends, and prints the result. misegen_test.go prints the same text from
# the Go generator. The case folder holds:
#   side             repo or macos
#   mode             repo: full; macos: full, direct, own or render
#   clone/           the clone (repo/ or macos/ in it)
#   root/            repo: the project root, with .config/bootstrap.toml
#   config/          macos: the folder of macos.toml (full mode)
#   values.toml      macos direct, own, render: the values file
#   tools            macos direct: the tools, one line
#   local            repo: present for the --check form of the overrides
#   aliases          repo: the dev aliases of the tool mise
# The steps before the generator (the read of the values and the selection)
# print only their exit code, so the test sees a broken case.
set -u
engine=$1 c=$2
side=$(< "$c/side")
mode=$(< "$c/mode")
dir=$c/clone

# err <file>: the log lines of one step, without the JSON lines of gomplate.
err() { grep -v '^{"time":' "$1" | sed 's/^/| /'; }
step() {
  local name=$1 rc=0 f="$c/step.err"
  shift
  "$@" 2> "$f" || rc=$?
  printf '%s: rc=%s\n' "$name" "$rc"
  err "$f"
  return "$rc"
}
written() { printf '%s' "$_SE_MISE_WRITTEN" | tr ' ' '\n' | sed '/^$/d' | sed 's/^/  /'; }
files=()
finish() {
  echo "--- files"
  for f in ${files[@]+"${files[@]}"}; do
    echo "== $f $(_se_mode "$c/$f")"
    cat -- "$c/$f"
  done
  exit 0
}

if [[ $side == repo ]]; then
  # shellcheck source=/dev/null
  source "$engine/repo-sync"
  _SE_PREFIX=mise
  root=$c/root
  export MISE_PROJECT_ROOT=$root
  git init -q "$root"
  cd "$root" || exit 1
  values=$root/.config/bootstrap.toml
  rc=0
  { _rs_envs "$values" && _rs_mise_index "$dir" && _rs_resolve "$dir" "$values"; } 2> "$c/pre.err" || rc=$?
  echo "pre: rc=$rc"
  ((rc == 0)) || finish
  step envs _se_mise_envs repo "$_RS_ENVS" "$_RS_SEL" "$_RS_DEFAULT_ENVS" "$_RS_DEPS" "$_RS_MISE_ENTRIES" "$_RS_MISE_FOUND" "$_RS_TOOLS_DIR" _rs_has_file || finish
  echo "selected:$_SE_MISE_SELECTED"
  echo "tool_envs:$_SE_MISE_TOOL_ENVS"
  echo "mise_envs:$_SE_MISE_ENVS"
  only=''
  [[ -f $c/local ]] && only=local
  step overrides _se_mise_overrides repo "$values" "$dir/$_RS_TOOLS_DIR" "$_RS_TOOLS_DIR" "$_RS_TOOLS" "$_RS_ENVS" $only || finish
  [[ -z $only ]] || finish
  _rs_tool_values "$dir" "$values" 2> "$c/pre.err" || {
    echo "values: failed"
    finish
  }
  aliases=''
  [[ -f $c/aliases ]] && aliases=$(< "$c/aliases")
  for tool in $_RS_TOOLS; do
    a=''
    [[ $tool == mise ]] && a=$aliases
    step "render $tool" _se_mise_render repo "$dir/$_RS_TOOLS_DIR" "$_RS_TOOLS_DIR" "$tool" "$values" "$dir/$_RS_TOOLS_DIR/$tool/.config/mise/conf.d" "$a" "$_RS_VALUES_FILE"
    written
    for w in $_SE_MISE_WRITTEN; do
      files+=("clone/repo/tools/$tool/.config/mise/conf.d/$w")
    done
  done
else
  confd=$c/out/conf.d
  if [[ $mode == full ]]; then
    # shellcheck source=/dev/null
    source "$engine/setup-sync"
    _SE_PREFIX=mise
    _SS_DIR=$dir
    _SS_TOOLS_DIR=$dir/macos/tools
    _SS_PACKS=$dir/macos/packs.toml
    _SS_CONFIG=$c/config
    export MACOS_SETUP_CONFIG_PATH=$c/config
    _SS_WORK=$c/sswork
    mkdir -p "$_SS_WORK"
    cd "$c/config" || exit 1
    rc=0
    { _ss_read_values && _ss_manifests && _ss_select && _ss_tool_values; } 2> "$c/pre.err" || rc=$?
    echo "pre: rc=$rc"
    ((rc == 0)) || finish
    values=$_SS_VALUES tools=$_SS_TOOLS all=$_SS_ALL_VALUES
  else
    # shellcheck source=/dev/null
    source "$engine/common"
    # shellcheck source=/dev/null
    source "$engine/values"
    # shellcheck source=/dev/null
    source "$engine/mise"
    _SE_PREFIX=mise
    cd "$c" || exit 1
    values=$c/values.toml
    [[ -f $values ]] || values=''
    tools=''
    [[ -f $c/tools ]] && tools=$(< "$c/tools")
    all=''
    if [[ $mode == direct ]]; then
      all=$c/all.json
      out=$(_se_values "$tools" '' "$dir/macos/tools" '{}' macos/tools macos.toml 2> "$c/pre.err") || {
        echo "pre: rc=1"
        finish
      }
      printf '%s\n' "$out" | _se_jq -R -n 'reduce (inputs | split("\t")) as $l ({}; .[$l[1]] = ($l[2] | fromjson))' > "$all"
    fi
    echo "pre: rc=0"
  fi
  if [[ $mode == full || $mode == direct ]]; then
    if [[ -n $values ]]; then
      step overrides _se_mise_overrides macos "$values" "$dir/macos/tools" macos/tools "$tools" '' || finish
    fi
  fi
  if [[ $mode != own ]]; then
    for tool in $tools; do
      step "render $tool" _se_mise_render macos "$dir/macos/tools" macos/tools "$tool" "$values" "$confd" '' "$all"
      written
      for w in $_SE_MISE_WRITTEN; do
        files+=("out/conf.d/$w")
      done
    done
  fi
  if [[ $mode == own || ($mode == full && -n $values) ]]; then
    owntools=$tools
    [[ $mode == own ]] && owntools=''
    step own _se_mise_own_config macos "$values" "$confd" "$owntools" "$all"
    written
    for w in $_SE_MISE_WRITTEN; do
      files+=("out/conf.d/$w")
    done
  fi
fi
finish
