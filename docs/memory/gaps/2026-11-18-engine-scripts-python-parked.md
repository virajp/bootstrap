GAP 2026-11-18 ★3 | backlog:B30, repo:bootstrap, cli/internal/scripts,
repo/tools/*/scripts | the tool scripts that the engine runs stay bash | parked:
a later plan, after the mise tasks are Python

**Source** The change-plan interview for B30 (the mise tasks in Python). The
user said: "Let's focus on `mise tasks` only for now".

What the B30 plans decide: every mise task script of `repo/tools/*/tasks/`,
`macos/tools/*/tasks/` and `.config/mise/tasks/` becomes a Python uv script.

What is open: the tool scripts of `repo/tools/*/scripts/` (`detect`, `derive`,
`check`, `migrate-*`, about 30 files) stay bash. The engine runs each one as
`mise x -- bash <path>` with one JSON object on stdin
(`cli/internal/scripts/scripts.go:161`). The `Detect` commands of the questions
run with `bash -c` (`cli/internal/questions/questions.go:168`). A move to Python
changes the Go runner and needs an engine release.
