# U1 — `remove tools` deletes Template outputs and dependency-only tools

- **Wave:** 1
- **Depends on:** —
- **Owns:** `docs/sync/vwf.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `docs/sync/vwf.md`, top to bottom; `CLAUDE.md` section
  "Layout rule" (the file kinds and the header rule).
- **Lazy-load:** `docs/sync/base.md`, `docs/sync/github.md:25`,
  `docs/sync/vscode.md:25` (the `Template` rows);
  `tools/base/.config/mise/tasks/_scripts/repo-sync:620-640` (the header
  marker, first three lines); `docs/sync/index.md:138,207` (why `base` is
  necessary).

## Ruling

From index.md, assumed decisions:

> R1 — `remove tools` also deletes the `Template` outputs that the tool's
> `docs/sync/<tool>.md` lists, but only a file that carries the
> `Synced from bootstrap` header in its first three lines. A `Template` output
> without the header stays.

> R2 — The stop on `remove tools [base]` is ratified, because
> `docs/sync/index.md:138,207` make `base` necessary. No change.

> R3 — Before the change, the agent finds the tools that the `depends` of the
> remaining tools no longer need and that the values file does not list. It
> deletes their `Synced` files and their `Template` outputs with the header,
> the same way. The values file does not change for them, because they were
> never listed.

## Edits

1. **`docs/sync/vwf.md`, Remove tools section (:95-107)**:
   - Change the deletion step: delete each file that the removed tool's
     `docs/sync/<tool>.md` lists as `Synced`, and each file it lists as
     `Template` whose first three lines contain `Synced from bootstrap`. The
     output path of a `Template` row has no `.tmpl` suffix.
   - Keep `Seed` and `Seed, template` as repository-owned (no change), and keep
     the `Generated` sentence (no change).
   - Replace the sentence that keeps "any other kind, for example `Template`"
     with: a `Template` output without the header stays; any other kind
     stays.
2. **`docs/sync/vwf.md`, Remove tools section**: add one step before the
   deletion step. The agent computes the set of tools that the selection needs
   before the change (the listed tools plus their transitive `depends`, as the
   Add section computes it at :83-86, including the tools that a mise
   `depends` adds) and the same set after the change. Each tool in the first
   set and not in the second, and not listed in the values file `tools`, is a
   dependency-only tool that leaves. The agent deletes its files by the rule
   of edit 1. The values file changes only for the named tools.
3. **`docs/sync/vwf.md`**: the stop on `remove tools [base]` (:63) stays as it
   is (R2). Do not change it.
4. Keep the section order, the header comment and the style of the file.
   Refold the changed paragraphs to the dprint line width.

## Verification

- Wave-gate line 1 passes (dprint checks the Markdown).
- `grep -n 'Template' docs/sync/vwf.md` shows the new deletion rule in the
  Remove tools section, and no sentence that says a `Template` output with the
  header stays.
- The Remove tools section names the dependency-only tools and says that the
  values file does not change for them.

## Guardrails

- Do not touch any other file. Report each falsified passage in another file
  as `DOCS FALSIFIED:`.
- Do not change the Add section, except to refer to its closure rule.
- Delete with `rm`, never `git rm`.

## Commit

`docs(vwf): remove deletes Template outputs and dependency-only tools`
