#!/usr/bin/env bash
# Usage: macos.sh <engine-dir> <clone-copy> <config-dir>
# Runs the read steps of setup_sync: the values file, the manifests, the
# selection and the merge of the values.
set -u
engine=$1 dir=$2 config=$3
# shellcheck source=/dev/null
source "$engine/setup-sync"
_SS_DIR=$dir
_SS_TOOLS_DIR=$dir/macos/tools
_SS_PACKS=$dir/macos/packs.toml
_SS_CONFIG=$config
export MACOS_SETUP_CONFIG_PATH=$config
_SS_WORK=$(mktemp -d)
cd "$config" || exit 1
rc=0
{ _ss_read_values && _ss_manifests && _ss_select && _ss_tool_values; } || rc=$?
echo "rc=$rc"
if ((rc == 0)); then
  echo "tools:$(printf ' %s' $_SS_TOOLS)"
  echo "init:$(printf ' %s' $_SS_INIT)" | sed 's/ *$//'
  echo "keep:"
  printf '%s\n' "$_SS_KEEP"
  echo "values:"
  _se_jq -c . "$_SS_ALL_VALUES"
fi
rm -rf "$_SS_WORK"
