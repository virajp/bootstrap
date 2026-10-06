# U4 — Setup config: schema, read, validate, write

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/setup-config/**`, `cli/test/setup-config/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/setup-config/setup-config.test.ts` — one failing
  case per [Setup config](../../blueprint/entities/setup-config/index.md)
  Validity rule (schema violation, newer `format`, newer `version`, a removed or
  core tool name in `values.tools`, a path in both `kept` and `deleted`, a
  non-relative / `..` / wildcard path), each a typed error carrying what, why
  and the fix; a write omits empty `kept` and `deleted`, never lists
  `.config/bootstrap.yaml` in `files`, and a read of that output round-trips.
  Fails because the module does not exist.
- **Read first:** `docs/blueprint/entities/setup-config/index.md`,
  `docs/blueprint/entities/setup-config/schema.yaml`, `cli/src/tool/catalog.ts`.
- **Lazy-load:** Effect Schema and `yaml` docs via Context7;
  `docs/blueprint/conventions.md` `#errors`.

## Ruling

> | 5 | Setup config parsing | the `yaml` package reads and writes
> `.config/bootstrap.yaml`; Effect Schema validates it against `schema.yaml` | a
> hand-written parser | U4 |

## Edits

1. **`cli/src/setup-config/schema.ts`** — an Effect Schema equal to
   `schema.yaml`: `format` (1), `version` (semver), `values` (`repo` with one or
   more owner segments then the name, `commit_scopes` lowercase kebab-case,
   `merge_model.develop` / `.main` in `direct` | `pr`, `tools`), `files`,
   optional `kept` and `deleted`; closed objects (no extra keys).
2. **`cli/src/setup-config/validity.ts`** — the Validity rules beyond the
   schema, each a typed error with what, why and the fix text the entity doc
   gives; the running cli's version and format are inputs; tool names are
   checked against the catalog.
3. **`cli/src/setup-config/io.ts`** — read the file at
   `<repository root>/.config/bootstrap.yaml` through the Effect file-system
   service (absent → a distinct "absent" result, not an error), parse with
   `yaml`, decode and validate; write: encode, omit empty `kept`/`deleted`, sort
   every path list by path, never put `.config/bootstrap.yaml` in `files`.
   Writing is a function that returns the file text; the atomic write and
   rollback are plan 2's.
4. **`cli/test/setup-config/**`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- Do not touch `cli/src/tool/**` (U2, U3).
- No `node:fs` import; use the platform file-system service so tests run on a
  stub layer.

## Commit

`feat: read, validate and write the setup config`
