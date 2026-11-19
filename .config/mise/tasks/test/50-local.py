#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.14"
# dependencies = []
# ///
# MISE description="Run every test: the Go module of cli/ and the end-to-end tests of e2e/"
# MISE depends=["test:go"]
# MISE tools={uv="latest", python="3.14", gotestsum="latest", gomplate="latest"}
# USAGE flag "--all" help="run each test again, also a test whose inputs did not change (go test -count=1)"

# Runs `go test` with gotestsum in the two Go modules of this repository, at
# the same time: cli/ (the unit tests of the engine) and the root (the
# end-to-end tests of e2e/). `go test` keeps the result of a package whose
# inputs did not change, and does not run it again; --all runs each package.
# Each run writes the JSON log of `go test -json` to
# ${XDG_STATE_HOME:-$HOME/.local/state}/bootstrap/test-runs/, with the time of
# each test and the TIMING lines of the end-to-end cases, and prints the
# slowest tests.

import os
import subprocess
import sys
import time

sys.dont_write_bytecode = True
sys.path.insert(0, f"{os.environ['MISE_PROJECT_ROOT']}/.config/mise/tasks/_scripts")
import helpers  # noqa: E402

# The tests never read the real config file of the user: the engine reads
# this path, which does not exist, so each setting has its default.
os.environ["BOOTSTRAP_CONFIG"] = "/nonexistent/bootstrap/config.toml"

root = os.environ["MISE_PROJECT_ROOT"]
state = os.environ.get("XDG_STATE_HOME") or f"{os.environ['HOME']}/.local/state"
runs = f"{state}/bootstrap/test-runs"
os.makedirs(runs, exist_ok=True)
stamp = time.strftime("%Y%m%dT%H%M%SZ", time.gmtime())
log = f"{runs}/{stamp}.json"
part = f"{runs}/.{stamp}"

# mise gives the tool folders of a task in a different order on each run, and
# `go test` keeps a result only for the same PATH. The tool folders of mise
# go in the order of their names, at the place of the first one.
dirs = os.environ.get("PATH", "").split(":")
tools = sorted(d for d in dirs if "/mise/installs/" in d)
stable = []
done_tools = False
for folder in dirs:
  if "/mise/installs/" in folder:
    if done_tools:
      continue
    stable.extend(tools)
    done_tools = True
  else:
    stable.append(folder)
os.environ["PATH"] = ":".join(stable)

go_flags = ["-count=1"] if os.environ.get("usage_all", "false") == "true" else []

helpers.print_header("Running the Go tests ...")
start = time.monotonic()
cli = subprocess.Popen(
  ["gotestsum", "--format", "pkgname", "--jsonfile", f"{part}.cli.json", "--", *go_flags, "./..."],
  cwd=f"{root}/cli",
)
rc_e2e = subprocess.run(
  ["gotestsum", "--format", "pkgname", "--jsonfile", f"{part}.e2e.json", "--", *go_flags, "./..."],
  cwd=root,
  check=False,
).returncode
rc_cli = cli.wait()
with open(log, "wb") as out:
  for name in (f"{part}.cli.json", f"{part}.e2e.json"):
    with open(name, "rb") as f:
      out.write(f.read())
for name in (f"{part}.cli.json", f"{part}.e2e.json"):
  try:
    os.remove(name)
  except FileNotFoundError:
    pass

helpers.print_subheader("Listing the slowest tests ...")
result = subprocess.run(
  ["gotestsum", "tool", "slowest", "--jsonfile", log, "--threshold", "5s", "--num", "10"],
  check=False,
)
if result.returncode != 0:
  sys.exit(result.returncode)
helpers.print_info(f"The log of the run: {log}")
helpers.print_info(f"The run took {int(time.monotonic() - start)} s")

if rc_cli or rc_e2e:
  helpers.print_error("Tests failed")
  sys.exit(1)
helpers.line_sep1()
helpers.print_success("All tests passed")
