# U3 — Gates

- **Read first:** `cli/internal/source/toml_test.go` (lines 16-20 and 90-113:
  how the golden file is written).

## Ruling

> S1: cli/internal/source/testdata/toml-yq.golden is generated from every TOML
> file of repo/; this unit regenerates it.

> D11: U3 is the fixed gates unit: it runs the golden generator.

## Edits

1. **`cli/internal/source/testdata/toml-yq.golden`** — regenerate it: in
   `cli/`, run `UPDATE_GOLDEN=1 mise x -- go test ./internal/source`. The bun
   lines go and the android line changes. Keep what the test wrote; do not
   edit the file by hand.
2. Run the full wave gate and the plan gate lines, and report each result.

## Verification

- `mise x -- mise run code:lint --all` is green.
- `mise x -- mise run test` is green.
- `git diff cli/internal/source/testdata/toml-yq.golden` changes only the lines
  of the bun files and of `repo/tools/android/tools.toml` (D7: "Only the lines
  of the TOML files this plan changes or deletes. Any other changed line is
  reported as a QUESTION and not kept.").
- The plan gate runs `mise run test` with
  `CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins`.

## Guardrails

- Bump no version. Edit no doc.
- If a test outside the golden fails, report it as a QUESTION; do not fix
  another unit's file.

## Commit

`test: regenerate the TOML golden file`
