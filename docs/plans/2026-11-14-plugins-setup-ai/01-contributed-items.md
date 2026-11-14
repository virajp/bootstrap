# U1 — The engine: contributed merge items

- **Read first:** `cli/internal/values/merge.go` (the `[contributes.<tool>]`
  merge of values, about lines 93-125 and 289-318),
  `cli/internal/format/shared.go` (the merge-mode items and the override
  table, about lines 306-347), the code that reads a merge-mode file's override
  table from the owning tool's merged values (`cli/internal/macos/file.go`,
  about lines 187-193, and its repo-side twin), `cli/internal/source/check.go`
  (the rules), and `repo/tools/git/tools.toml` (`.gitignore` as a merge-mode
  file, about lines 61-64) with `repo/tools/git/files/gitignore.merge.toml`
  (the item format, lines 1-8).
- **Lazy-load:** `docs/design/tools.md` (`[contributes.*]`, merge mode),
  `docs/design/repo.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1, D7, D11:** a tool may declare, in its `tools.toml`,
  `[[contributes_items.<owner>."<target>".item]]` entries in the item format of
  the owner's merge file (`value = "<line>"`, and the other keys that format
  allows). `contributes_items` is a separate top-level key that the values
  merge of `[contributes.<tool>]` does not read. When the owner tool and the
  contributing tool are both
  selected, the engine adds the contributed items to the owner's merge items
  for that target, after the owner's own items, in the order of the
  contributing tools' names, with no duplicates. The owner stays the one owner
  of the target in the manifest. `check` refuses, with code 2: a contributed
  item for a target that the owner does not declare as a merge-mode file, and
  a `contributes_items.<owner>` whose owner tool does not exist in the catalog.
  A
  contributing tool that is selected while the owner is not adds nothing (no
  error).
- The override table of the owner (`"<line>" = {"$mode"="remove"}`) applies to
  contributed items as to the owner's own items.
- **D17:** an override key that names a line that any catalog tool contributes
  to that owner and target counts as known, not unused, also when the
  contributing tool is not selected.
- **D15:** `bootstrap config get` of a `tools.*` key returns the merged value:
  the file, the defaults and the values that other tools contribute through
  `[contributes.<tool>]`. A value that comes only from a contribution has
  `source: "contributed"` (additive to R1a's `file` and `default`).
- **D20:** `source` is `"file"`, `"contributed"` or `"default"` for the value
  as a whole. When a contribution adds entries, the result also has
  `contributed_by`, the list of the contributing tool names.
- **D23:** `config get` reads the cached clone when one exists and never
  fetches. With no cache, it returns the file value alone, with no
  `contributed_by`.
- **D27:** `config get` also accepts `--clone <path>` and `BOOTSTRAP_CLONE` as
  its catalog, with no fetch and no TTL.
- **D28:** the `get` envelope has `"catalog": false` when `get` reads no
  catalog, and `"catalog": true` when it reads one.
- **D24 (replaces D22):** this unit adds nothing to `config set`. R1a's `set`
  already types a list given as JSON.

## Edits

1. **The merge of a merge-mode file** — collect the contributed items of every
   selected tool for that owner and target, and add them as D1 says.
2. **`cli/internal/source/check.go`** — the D7 rules.
3. **Tests** — Go tests: a contributed item reaches the owner's file; two
   contributors in name order; a duplicate line appears once; an override
   removes a contributed line; an unselected owner adds nothing; each D7
   refusal; an override of a line that an unselected tool contributes is not
   reported as unused (D17); `config get` of a contributed `tools.*` value
   returns it with `source: "contributed"` (D15), and a value with file and
   contributed entries has `contributed_by` (D20); with no cached clone,
   `config get` returns the file value alone with `catalog: false` and makes
   no fetch (D23, D28); `config get --clone <path>` reads that catalog (D27).
4. **`config get`** — D15, D20, D23, D27, D28.
5. **The unused-override rule** — D17.

## Verification

- `mise x -- mise run test:go` passes.

## Guardrails

- Touch nothing outside `cli/`.
- Do not change the merge of a file that has no contributed items.
- Delete with `rm`, never `git rm`.

## Commit

`feat: a tool can contribute items to another tool's merge-mode file`
