# U4 — Gates: pass the full wave gate and the full-selection dry-run

- **Wave:** 3
- **Depends on:** U3
- **Owns:** none. A fix that a gate needs goes back to the unit that owns the
  file, through the orchestrator.
- **Model:** opus
- **Kind:** edit
- **Read first:** index.md sections "Wave gate" and "Gates the orchestrator
  keeps".
- **Lazy-load:** `tools/base/.config/mise/tasks/_scripts/repo-sync:8-110`
  (exit codes).

## Ruling

From index.md, assumed decisions:

> D2 — A new `java` tool. `tools/java/tools.toml` defines only
> `[mise.tools.java]`, `version = "lts"`, `envs = ["dev", "ci"]`.

> D3 — A new `android` tool defines `[mise.tools.android-sdk]`,
> `version = "23"`, `envs = ["dev", "ci"]`, mise `depends = ["java"]`.

> D4 — `tools/flutter/tools.toml` drops `[mise.tools.android-sdk]` and
> `[mise.tools.java]` and sets `depends = ["android"]`.

This plan names no generator. No version file exists to bump.

## Edits

1. Run wave-gate line 1: `mise run code`.
2. Run wave-gate line 2 (index.md, Wave gate), from the root of the worktree.
3. Run the full-selection dry-run, from the root of the worktree:

   ```text
   bash -c 't=$(mktemp -d) && sed "s/^tools *=.*/tools = [\"base\", \"flutter\", \"kotlin\", \"android\", \"java\", \"swift\"]/; s/^environments *=.*/environments = [\"dev\", \"ci\"]/" .config/bootstrap.toml > "$t/values.toml" && REPO_SYNC_CLONE="$PWD" REPO_SYNC_VALUES="$t/values.toml" mise run setup:configs:sync --dry-run'
   ```

## Verification

- Line 1 exits 0.
- Line 2 exits 0.
- The full-selection dry-run exits 0, and its output:
  - lists `.config/mise/conf.d/java/mise.dev.toml`,
    `.config/mise/conf.d/java/mise.ci.toml`,
    `.config/mise/conf.d/android/mise.dev.toml` and
    `.config/mise/conf.d/android/mise.ci.toml`;
  - shows no `android-sdk` and no `java` in the rendered
    `.config/mise/conf.d/flutter/mise.*.toml`.
- `git status --short` in the worktree shows no file that the dry-runs wrote.

## Guardrails

- Change no file. If a line fails, return `UNRESOLVED:` with the failing
  command, its exit code, the last lines of its output, and the unit whose
  file causes it.
- Do not run `mise run setup:configs` without `--dry-run`, and do not run the
  after-landing steps.

## Commit

None when nothing changes. If the orchestrator routes a fix through this
unit: `fix: pass the wave gate of the android, swift and java tools`.
