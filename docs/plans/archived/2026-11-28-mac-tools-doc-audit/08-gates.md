# U8 — Gates

- **Read first:** `index.json` of this folder (`gates`).
- **Lazy-load:** none.

## Ruling

> D12: No. The plan lands no runnable code, only docs, so the wave review is
> the only check.

## Edits

1. No generator runs: the plan changes no generated file. Run the full wave
   gate over the finished tree and report its result.

## Verification

- `mise x -- mise run code:lint --all` passes.
- `git status --porcelain` is empty.

## Guardrails

- Do not edit a file. Do not bump a version.

## Commit

`docs: pass the gates of the Mac tools doc audit` — only when a gate fix
changed a file; otherwise no commit.
