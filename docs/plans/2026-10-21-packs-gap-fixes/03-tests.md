# U3 — Test the dev rule and the zz-repo rule

- **Wave:** 1
- **Depends on:** —
- **Owns:** `.config/mise/tasks/test/repo-sync`
- **Model:** opus
- **Kind:** edit
- **Read first:** the owned file, top to bottom (`run_case` at line 46).

## Ruling

Quoted from `index.md`:

- D3: "A tool folder whose name sorts after `zz-repo` stops the sync with
  code 2."
- D4: "Each `env` of the values file and each `default_envs` without `dev`
  stops the sync with code 2."

## Edits

1. Change the case "env ci on a tool whose mise tool allows dev only"
   (line 142) to `env = ["dev", "ci"]`, so it still tests the `envs` of the
   mise tool and not the new `dev` rule.
2. Add three cases, each with code 2:
   - `env = ["ci"]` on a `[tools.*]` entry (no `dev`);
   - a `tools.toml` in the copy with `default_envs = ["ci"]`;
   - a tool folder in the copy named `zzz-late`, which sorts after `zz-repo`.
3. Keep the message checks of `run_case` for each new case: match the words
   that U1's messages use (`dev`, `zz-repo`).

## Verification

- `mise x -- mise run test:repo-sync` passes after wave 1, with U1 and U2.
- `bash -n` and `mise x -- shellcheck -x -s bash` pass.
- `mise x -- mise run code:precommit` passes.

## Guardrails

- Touch only the owned file. The task needs no network.

## Commit

`test: cover the dev rule and the zz-repo rule of repo-sync`
