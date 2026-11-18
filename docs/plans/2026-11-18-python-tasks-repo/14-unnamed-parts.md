# U14 — The task name of a .py file in UnnamedParts

- **Read first:** `cli/internal/oldform/tasks.go` (`UnnamedParts`,
  `dependsOf`, about lines 175-245) and `cli/internal/oldform/tasks_test.go`
  (`TestUnnamedParts`, about line 275).
- **Lazy-load:** `docs/design/cli.md` (the warnings of `bootstrap check`),
  `docs/plans/archived/2026-11-14-task-ports/index.json` (ruling D15, the
  origin of the warning).

## Ruling

- **D15:** "UnnamedParts builds the task name of a file from its file name
  and looks for it among the # MISE depends lines of the folder. mise removes
  .py from a task name, so for go.py the engine builds test:go.py, and the
  depends line test:go does not match: a false warning. The lookup removes
  .py from the file name before the compare; the warning text keeps the real
  file name. The fix ships with the next engine release; B30 cuts no tag
  (user: \"Yes, engine fix\")."
- **D14:** a Go test of the engine runs from `cli/`.

## Edits

1. **`cli/internal/oldform/tasks.go`** — in `UnnamedParts`, the compare
   `named[task+":"+n]` uses the name without a `.py` suffix
   (`strings.TrimSuffix(n, ".py")`). The path in the output keeps the file
   name with `.py`. Update the comment of the function: a `.py` file has the
   task name of its file name without `.py`, as mise gives it.
2. **`cli/internal/oldform/tasks_test.go`** — `TestUnnamedParts` gets two
   cases: a `test/go.py` that a `test/50-local.py` names in
   `# MISE depends=["test:go"]` gives no warning; a `test/x.py` that no file
   names gives the warning with the path `.config/mise/tasks/test/x.py`.

## Verification

- `cd cli && mise x -- go test ./internal/oldform/...` passes.
- `cd cli && mise x -- go vet ./...` passes.

## Guardrails

- Change only the compare and its comment. No other engine behavior changes.
- Do not bump `cli/VERSION`: the release workflow writes it.
- The engine code names no tool (`TestEngineNamesNoTool`): the test data may
  name `go`, the code may not.
- Write each comment in ASD-STE100.

## Commit

`fix: remove .py from a file name before the depends compare of UnnamedParts`
