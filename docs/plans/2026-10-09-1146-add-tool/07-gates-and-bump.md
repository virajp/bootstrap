# U7 — Gates and bump

- **Wave:** 5
- **Depends on:** U6
- **Owns:** — (no version bump; no generated file)
- **Model:** opus
- **Kind:** edit

## Ruling

| Action               | Granted                       |
| -------------------- | ----------------------------- |
| Release cli publicly | none (stays `0.0.1`; no bump) |

## Edits

1. Bump no version. The release intent is `none`.
2. Run the full wave gate:

   ```text
   pnpm install --frozen-lockfile
   mise x -- mise run p:cli:check
   mise x -- mise run p:cli:test
   mise x -- mise run p:cli:build
   mise x -- mise run p:cli:e2e
   ```

## Verification

- Every gate line exits 0. The report is the run's final gate.

## Guardrails

- Do not write the `implementation:` stamps; the executor's Reconcile step
  writes them.
- Do not run the after-landing steps; the orchestrator runs them after the
  landing.

## Commit

None unless a gate line changes a file. If one does, report it as `GAP:`.
