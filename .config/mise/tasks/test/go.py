#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.14"
# dependencies = []
# ///
# MISE description="Vet and format-check the Go modules of cli/ and of the root"
# MISE tools={uv="latest", python="3.14"}

# Runs `go vet` and a `gofmt` check in the two Go modules of this repository:
# cli/ (the engine) and the root (the end-to-end tests of e2e/). The task
# `test` runs the tests. The task runs the Go of the `go` tool, which this
# repository selects.

import os
import subprocess
import sys

sys.dont_write_bytecode = True
sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
import helpers  # noqa: E402

root = os.environ["MISE_PROJECT_ROOT"]

helpers.print_header("Examining the Go modules ...")

helpers.print_subheader("Running go vet ...")
for folder in (f"{root}/cli", root):
  result = helpers.tool("go", "vet", "./...", cwd=folder)
  if result.returncode != 0:
    sys.exit(result.returncode)

helpers.print_subheader("Examining the format with gofmt ...")
result = helpers.tool("gofmt", "-l", f"{root}/cli", f"{root}/e2e", stdout=subprocess.PIPE, text=True)
if result.returncode != 0:
  sys.exit(result.returncode)
unformatted = result.stdout.rstrip("\n")
if unformatted:
  helpers.print_error("files that gofmt changes:")
  sys.stderr.write(f"{unformatted}\n")
  sys.exit(1)

helpers.print_success("The Go modules pass go vet and gofmt")
