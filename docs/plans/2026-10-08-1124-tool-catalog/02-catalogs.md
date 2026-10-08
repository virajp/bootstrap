# U2 — Tool category catalog and the rescoped tool catalog

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/tool/catalog.ts`, `cli/src/tool-category/catalog.ts` (new),
  `cli/test/tool/catalog.test.ts`, `cli/test/tool-category/catalog.test.ts`
  (new), `cli/src/tool/render.ts` (edits that follow the new types only),
  `cli/src/setup-config/validity.ts` (the `kind` check at `:183` only)
- **Model:** opus
- **Kind:** code
- **Test first:**
  - `cli/test/tool-category/catalog.test.ts`: the 13 categories in the blueprint
    order, each with `max` and `purpose` equal to the
    [Tool category](../../blueprint/entities/tool-category/index.md) Catalog
    table. Category invariant 6 (every category holds at least one tool) and
    invariant 5 (a `max: one` category holds at most one `default: on` tool and
    at most one unremovable tool) hold over the data.
  - `cli/test/tool/catalog.test.ts`: the 15 tools in the blueprint order, each
    row equal to the [Tool](../../blueprint/entities/tool/index.md) Catalog
    table (name, category, purpose, base, removable, replaceable, default,
    `origin_hosts`, requires) and to the Target paths table (paths and
    `create_only`). Tool invariants 1–4 and 7–12 hold over the data. A duplicate
    path fails the build of the path index.
  - Both fail because the category catalog does not exist and the tool rows have
    the old shape.
- **Read first:** `docs/blueprint/entities/tool/index.md`,
  `docs/blueprint/entities/tool/schema.yaml`,
  `docs/blueprint/entities/tool-category/index.md`,
  `docs/blueprint/entities/tool-category/schema.yaml`, every owned file.
- **Lazy-load:** `cli/src/tool/templates.ts`.

## Ruling

| #  | Decision                                   | Ruling                                                                                                                                          | Rejected                             | Unit |
| -- | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ---- |
| 1  | Drift: code has `kind: core \| non-core`   | Change the code to `category`, `base`, `removable`, `replaceable`, `default`, `requires` and `origin_hosts`, as the Tool catalog says           | Amend the blueprint to keep `core`   | U2   |
| 6  | Catalog integrity invariants               | Tests over the static catalog data enforce tool invariants 1–12 and category invariants 5–6                                                     | Checks at runtime                    | U2   |
| 12 | Order of catalog fields that break callers | U2 makes only the edits that follow the new types in `render.ts` and `validity.ts:183`, so every wave compiles; U4 and U5 then do the behaviour | Leave the build broken between waves | U2   |

## Edits

1. **`cli/src/tool-category/catalog.ts`** (new) — a `ToolCategory` type (`name`,
   `max: "one" | "many"`, `purpose`) and a read-only `categories` array in the
   blueprint display order, plus `categoriesByName`. Data only, no I/O.
2. **`cli/src/tool/catalog.ts`** — replace `kind` with `category` (a category
   name), `base`, `removable`, `replaceable`,
   `default: "on" | "off" |
   "origin"`, `originHosts` (present only when
   `default` is `origin`), `requires` (tool names). `files` becomes a list of
   `{ path, createOnly }`; only `mempalace.yaml` is `createOnly: true`. Rows,
   order, purposes and paths equal the blueprint tables, with the new `_default`
   dispatcher paths under `mise` and the per-tool subtask and
   `conf.d/<tool>/mise.dev.toml` paths. Remove `.config/dprint.json` (dprint)
   and `conf.d/ai/mise.dev.toml` (claude). Remove `coreTools` and
   `nonCoreToolNames`; add `unremovableTools` and `defaultTools(originHost)`
   only if a test in this unit needs them. `toolByPath` fails on a duplicate
   path when the index is built.
3. **`cli/src/tool/render.ts`** — the edits that the new `files` shape needs to
   compile (read `file.path`). No new behaviour.
4. **`cli/src/setup-config/validity.ts`** — remove the `tool.kind === "core"`
   branch at `:183` and its error, because no `kind` exists. Leave every other
   rule for U4.
5. **`cli/test/tool/catalog.test.ts`**,
   **`cli/test/tool-category/catalog.test.ts`** — the test-first cases. Remove
   the `coreTools` assertions.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `grep -rn "coreTools\|nonCoreToolNames\|kind === \"core\"" cli/src cli/test`
  finds nothing.

## Guardrails

- Do not touch `cli/templates/**` (U3) or any other file of
  `cli/src/setup-config/` (U4).
- Template tests can fail on paths that U3 adds. If `core-templates.test.ts` or
  `new-templates.test.ts` fails only because a template file for a new catalog
  path does not exist yet, report it as a `GAP:` and do not edit those tests.
  They are U3's.
- Copy the purpose texts from the blueprint byte for byte.

## Commit

`refactor: model tool categories and the rescoped tool catalog`
