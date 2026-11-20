#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.14"
# dependencies = []
# ///
# MISE description="Build the engine of cli/ and link it into mise as github:virajp/bootstrap@local"
# MISE tools={uv="latest", python="3.14"}

# The local release of the engine: [release.bootstrap.local] of
# .config/bootstrap/repo.toml runs this task. It builds cli/ into
# ${XDG_DATA_HOME:-~/.local/share}/bootstrap/local, and runs
# `mise link --force github:virajp/bootstrap@local <folder>`, so each new build
# replaces the last one. A repository uses the build after
# `mise use github:virajp/bootstrap@local`, or with BOOTSTRAP_BIN set to the
# path of the binary. `version` of the build prints `<cli/VERSION>-local`.

import os
import subprocess
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
import helpers  # noqa: E402

TOOL = "github:virajp/bootstrap@local"


def main():
  root = os.environ["MISE_PROJECT_ROOT"]
  data = os.environ.get("XDG_DATA_HOME") or os.path.expanduser("~/.local/share")
  folder = os.path.join(data, "bootstrap", "local")
  with open(os.path.join(root, "cli", "VERSION")) as f:
    version = f.read().strip() + "-local"

  helpers.print_header("Building the local engine ...")
  os.makedirs(folder, exist_ok=True)
  build = helpers.tool(
    "go",
    "build",
    "-trimpath",
    "-ldflags",
    f"-X main.version={version}",
    "-o",
    os.path.join(folder, "bootstrap"),
    "./cmd/bootstrap",
    cwd=os.path.join(root, "cli"),
  )
  if build.returncode != 0:
    return build.returncode

  helpers.print_subheader(f"Linking {TOOL} ...")
  rc = subprocess.run(["mise", "link", "--force", TOOL, folder], check=False).returncode
  if rc != 0:
    return rc

  helpers.print_success(f"Built {version} at {folder}/bootstrap")
  helpers.print_info(f"Use it with: mise use {TOOL}, or BOOTSTRAP_BIN={folder}/bootstrap")
  return 0


if __name__ == "__main__":
  sys.exit(main())
