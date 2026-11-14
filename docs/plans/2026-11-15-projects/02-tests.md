# U2 — Integration tests

- **Read first:** `.config/mise/tasks/test/repo` (how a case builds a scratch
  repo and runs the engine), and `01-engine.md` of this folder.

## Ruling

- Add cases for D7 to D12: a pnpm and Go monorepo (the K1 shape); a
  single-package repo (the root is one project named `repo_name`); two
  folders with the same basename; a gitignored folder that is skipped; a
  deleted project folder (check code 2); `config unset` of a linked project
  (code 2).

## Verification

- `mise x -- mise run test` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`.

## Commit

`test: integration cases for projects and tool links`
