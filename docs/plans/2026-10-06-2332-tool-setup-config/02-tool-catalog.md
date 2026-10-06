# U2 — Tool catalog

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/tool/catalog.ts`, `cli/test/tool/catalog.test.ts`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/tool/catalog.test.ts` — the catalog has exactly the
  13 core tools and the non-core tools `github` and `gitlab` named in
  [Tool](../../blueprint/entities/tool/index.md) Catalog (1.0); each tool's
  paths equal the "Target paths (1.0)" table (with `tasks/` expanded to
  `.config/mise/tasks/`); no path is in two tools; `.config/bootstrap.yaml` is
  in no tool. Fails because the module does not exist.
- **Read first:** `docs/blueprint/entities/tool/index.md`,
  `docs/blueprint/entities/tool/schema.yaml`.
- **Lazy-load:** none.

## Ruling

The catalog is the contract in `docs/blueprint/entities/tool/index.md` — names,
kind (`core` / `non-core`), purpose and target paths, exactly as the doc pins
them. It is data, read-only at run time.

## Edits

1. **`cli/src/tool/catalog.ts`** — the 15 tools as typed, read-only data: name,
   kind, purpose, target paths (full repository-relative paths). Export the
   lookups later units need: all tools, core tools, non-core tool names, a tool
   by name, and the path → tool map. No I/O.
2. **`cli/test/tool/catalog.test.ts`** — the test-first assertions above, plus:
   every path is relative, has no `..` segment and no wildcard (setup-config
   validity applies the same rule to `files`).

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Do not touch `cli/package.json` or any file outside the owned list.
- The path list comes from the blueprint doc, not from this repo's tree.

## Commit

`feat: add the tool catalog`
