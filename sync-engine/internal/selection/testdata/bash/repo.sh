#!/usr/bin/env bash
# Usage: repo.sh <engine-dir> <clone-copy> <values-file> [submodule]
# Runs the repo-side read steps of repo_sync: the index, the resolve and the
# merge of the values. It prints the result on stdout and the log on stderr.
set -u
engine=$1 dir=$2 values=$3 sub=${4:-}
root=$(mktemp -d)
export MISE_PROJECT_ROOT=$root
if [[ -n $sub ]]; then
  # A git submodule: the superproject has the folder as a submodule.
  git init -q "$root/sub"
  git -C "$root/sub" -c user.name=t -c user.email=t@t commit -q --allow-empty -m init
  git init -q "$root/super"
  git -C "$root/super" -c protocol.file.allow=always submodule -q add "$root/sub" mod
  export MISE_PROJECT_ROOT=$root/super/mod
fi
# shellcheck source=/dev/null
source "$engine/repo-sync"
cd "$(dirname "$values")" || exit 1
rc=0
{ _rs_envs "$values" && _rs_mise_index "$dir" && _rs_resolve "$dir" "$values"; } || rc=$?
if ((rc == 0)); then
  _rs_tool_values "$dir" "$values" || rc=$?
fi
echo "rc=$rc"
if ((rc == 0)); then
  echo "tools:$(printf ' %s' $_RS_TOOLS)"
  echo "sel:"
  printf '%s\n' "$_RS_SEL"
  echo "deps:"
  printf '%s' "$_RS_DEPS"
  echo "values:"
  _se_jq -c . "$_RS_VALUES_FILE"
fi
rm -rf "$root"
