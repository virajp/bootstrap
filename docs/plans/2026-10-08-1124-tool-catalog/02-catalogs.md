# U2 — Tool category catalog and the rescoped tool catalog

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/tool/catalog.ts`, `cli/src/tool-category/catalog.ts` (new),
  `cli/test/tool/catalog.test.ts`, `cli/test/tool-category/catalog.test.ts`
  (new), `cli/test/catalog-driven.test.ts` (new), `cli/src/tool/render.ts`
  (edits that follow the new types only), `cli/src/setup-config/validity.ts`
  (the `kind` check at `:183` only)
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
  - `cli/test/tool/catalog.test.ts`, the catalog functions (rows 19, 20):
    `defaultSelection("github.com")` holds every `on` tool and `github`, and no
    `fnox`; `defaultSelection("gitlab.com")` holds `gitlab` and no `github`;
    `defaultSelection("example.org")` and `defaultSelection(undefined)` hold no
    `origin` tool. `unremovableTools` equals the tools with `removable: false`.
    `checkRequires` over a selection with `taplo` and without `dprint` returns
    the pair (`taplo`, `dprint`); over a selection with both, or with neither,
    it returns none. The test names tools only as data to check the results.
  - `cli/test/catalog-driven.test.ts` (row 13): reads every `.ts` file under
    `cli/src/tool/`, `cli/src/tool-category/` and `cli/src/setup-config/`,
    except `cli/src/tool/catalog.ts` and `cli/src/tool-category/catalog.ts`, and
    fails when a quoted string literal equals a tool name or a category name of
    the catalogs. It reads the names from the catalogs, not from a list in the
    test.
  - All fail because the category catalog does not exist and the tool rows have
    the old shape.
- **Read first:** `docs/blueprint/entities/tool/index.md`,
  `docs/blueprint/entities/tool/schema.yaml`,
  `docs/blueprint/entities/tool-category/index.md`,
  `docs/blueprint/entities/tool-category/schema.yaml`, every owned file.
- **Lazy-load:** `cli/src/tool/templates.ts`.

## Ruling

| #  | Decision                                                | Ruling                                                                                                                                                                                                                                                                                                                       | Rejected                             | Unit       |
| -- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ---------- |
| 1  | Drift: code has `kind: core \| non-core`                | Change the code to `category`, `base`, `removable`, `replaceable`, `default`, `requires` and `origin_hosts`, as the Tool catalog says                                                                                                                                                                                        | Amend the blueprint to keep `core`   | U2         |
| 6  | Catalog integrity invariants                            | Tests over the static catalog data enforce tool invariants 1–12 and category invariants 5–6                                                                                                                                                                                                                                  | Checks at runtime                    | U2         |
| 12 | Order of catalog fields that break callers              | U2 makes only the edits that follow the new types in `render.ts` and `validity.ts:183`, so every wave compiles; U4 and U5 then do the behaviour                                                                                                                                                                              | Leave the build broken between waves | U2         |
| 13 | Plan constraint: catalog-driven code (user, 2026-10-08) | No tool name or category name appears as a literal in `cli/src/{tool,tool-category,setup-config}` outside `cli/src/tool/catalog.ts` and `cli/src/tool-category/catalog.ts`. Every rule reads catalog fields. A test scans these folders and fails on a literal                                                               | Enforce it in the review only        | U2, U4, U5 |
| 19 | Tool `default` (first-run selection)                    | `defaultSelection(originHost)` selects every `on` tool, plus each `origin` tool whose `origin_hosts` holds the host, and never an `off` tool. Init `-y` and the tui first run use it                                                                                                                                         | Build it in plan 2                   | U2         |
| 20 | Tool `requires`                                         | `checkRequires(selection)` returns each pair (tool, missing required tool). Validity uses it over `values.tools` (exit 3). The flows use it over the selection after the request and name the missing tool (init, add) or the dependent tool (remove, replace) with exit 2. `unremovableTools` is part of the catalog module | A separate check for each command    | U2         |

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
   `nonCoreToolNames`. Add `unremovableTools`, `defaultSelection(originHost)`
   (row 19) and `checkRequires(selection)` (row 20). Each reads only catalog
   fields. `toolByPath` fails on a duplicate path when the index is built.
3. **`cli/src/tool/render.ts`** — the edits that the new `files` shape needs to
   compile (read `file.path`). No new behaviour.
4. **`cli/src/setup-config/validity.ts`** — remove the `tool.kind === "core"`
   branch at `:183` and its error, because no `kind` exists. Leave every other
   rule for U4.
5. **`cli/test/tool/catalog.test.ts`**,
   **`cli/test/tool-category/catalog.test.ts`**,
   **`cli/test/catalog-driven.test.ts`** — the test-first cases. Remove the
   `coreTools` assertions.

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
- Row 13: no tool name or category name as a literal in logic. Rules read the
  catalog fields (`removable`, `max`, `requires`, `default`, `origin_hosts`,
  `category`).

## Commit

`refactor: model tool categories and the rescoped tool catalog`
