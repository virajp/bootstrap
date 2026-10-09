# U4 — Resolution of values and tools (pure)

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/init/resolve.ts` (new), `cli/test/init/resolve.test.ts`
  (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/init/resolve.test.ts` — a pure function over the
  parsed flags, the recorded config (or none), the `origin` repo and host, and
  the catalog data. Tool names in the test come from the catalog data or are
  data to check results. One case per rule:
  - Values: a flag wins over the recorded value; the recorded value wins over
    the default; the repo from `origin` is used without `-y`; with no repo path
    at all, exit 2 naming `--repo`; a missing merge model without `-y` exits 2
    naming the flag, and the next command offers `bootstrap tui` or the flag,
    and `-y` only when `-y` would supply the value; with `-y` the merge models
    default to `direct` (develop) and `pr` (main).
  - Scopes: `--add-scope` adds once (also when repeated or already listed);
    `--remove-scope` of an unlisted scope leaves the list and adds the warning
    "scope `<s>` is not recorded"; `--reset-scope` empties the list; add and
    remove of different scopes both apply; a first run with `--add-scope api`
    gives `api` alone.
  - Selection: a first run with `-y` and no `--tool` gives `defaultSelection` of
    the `origin` host; without `-y` only the unremovable tools; a re-run without
    `--tool` keeps the recorded selection; `--tool` gives the full list of
    removable tools, unremovable tools are added, a repeated name is used once;
    a `requires` tool missing after the request exits 2 naming the required tool
    (through `checkRequires`).
  - Replacement (re-run): a different tool in a `max: one` category than
    recorded needs `--replace`, else exit 2; `-y` is not consent; replacing a
    tool whose `replaceable` is false exits 2; `--replace` with nothing to
    replace is ignored; the result lists `{from, to}`.
  - Drop (re-run): a dropped recorded removable tool, not a replacement, exits 2
    "removing files needs -y" without `-y`; proceeds with `-y`; under
    `--dry-run` it is reported and needs no `-y`.
  - It fails because `resolve.ts` does not exist.
- **Read first:** `docs/blueprint/flows/cli/110-setup-repository/index.md`
  (steps 3–6), `docs/blueprint/entities/tool/index.md` (requires paragraph),
  `docs/blueprint/entities/tool-category/index.md` (invariant 3),
  `docs/blueprint/conventions.md` `#config`, `#errors`, `cli/src/init/usage.ts`,
  `cli/src/report/errors.ts`.
- **Lazy-load:** `cli/src/tool/catalog.ts`, `cli/src/tool-category/catalog.ts`,
  `cli/src/setup-config/schema.ts`.

## Ruling

No assumed decision binds this unit beyond the plan constraint of row 7 in
`index.md`: "The literal scan of plan 1 (`cli/test/catalog-driven.test.ts`)
extends to every `.ts` file under `cli/src` except the two catalogs and
`cli/src/git/`." The rules read `removable`, `replaceable`, `max`, `category`,
`requires`, `default` and `origin_hosts` from the catalogs.

## Edits

1. **`cli/src/init/resolve.ts`** — the pure resolution: values, scopes,
   selection, replacements, drops; it returns the resolved values, the
   selection, the replaced pairs, the dropped tools and the warnings, or a usage
   error (exit 2) from `cli/src/report/errors.ts`.
2. **`cli/test/init/resolve.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- No git, no disk, no prompt: a pure function.
- Do not touch `cli/src/init/command.ts` or `usage.ts` (U2, U8).
- Copy message texts from the flow byte for byte.

## Commit

`feat: resolve init values, scopes and the tool selection`
