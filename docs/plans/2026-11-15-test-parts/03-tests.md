# U3 — Integration tests

- **Read first:** `.config/mise/tasks/test/repo` (how a case builds a scratch
  repo and runs the engine), and `01-engine.md` and `02-parts.md` of this
  folder.

## Ruling

- Add cases for D4 to D6: a known copy at `test/golden` is removed by
  `init --yes` and not moved to `50-local`; a known copy at
  `test/golden/50-local` is removed; an edited copy stays with no warning (D16);
  `check` gives code 2 before `init`; a repo without the
  `android` tool gets the file moved to `test/golden/50-local`, not removed
  (D17); `mise run test:golden` in a repo with the
  two tools and no Gradle or Xcode project prints on stdout exactly the
  success line of the R1c dispatcher; stderr is not examined (D19).
- **D25:** the case for a known copy at the old path `test/golden` selects
  android before the first `init`: the case writes the values file with a
  `[tools.android]` table (the R1a folder form; `init --yes` runs no detect,
  D26), and the old file at `test/golden`; then `check` gives code 2 and `init
  --yes` removes the file.
  The case for `50-local` runs `init`, then `bootstrap add tool android
  --yes`, then writes the file at `test/golden/50-local`.
- **D14, D15:** a case that runs `mise run` first runs `mise trust` in the
  scratch repo, and sets `MISE_TASK_RUN_AUTO_INSTALL=false`.
- A case builds the known copy from the bytes that give the `replaces` value
  U2 recorded: copy the file with `git show 0562cb44:<path>` from the
  claude-plugins clone at `$CLAUDE_PLUGINS_CLONE`; when that variable is not
  set, skip the case with one line (D21).

## Verification

- `mise x -- mise run test` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`.

## Commit

`test: integration cases for the tool test parts and the old copies`
