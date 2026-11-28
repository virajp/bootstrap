# U6 — Apply the audit to docs/setup/tools.md

- **Read first:** `docs/setup/tools.md`, top to bottom, then the five audit
  files `docs/plans/2026-11-28-mac-tools-doc-audit/audit/{a,b,c,d,general}.md`.
- **Lazy-load:** each `path:line` that a finding cites, when two findings
  contradict each other or a finding is not clear.

## Ruling

> D1: Audit docs/setup/tools.md against the current macos/tools/. The Notes of
> the deleted macos-setup docs/tool-reference.md (lines 71-201 at commit
> a4ec825^) are a checklist, not a source: each claim that is still true must
> be in the doc, and a false claim is dropped.

> D3: Each fact goes in the ### section of the tool that owns it. A rule for
> many tools goes in 'How a tool works', as a paragraph before the item legend
> ('Each section below gives these items for one tool:').

> D5: Audit in parallel, then apply once: U1 to U5 (wave 1) each write one
> findings file in docs/plans/2026-11-28-mac-tools-doc-audit/audit/ and edit
> nothing else; U6 (wave 2) applies the five files to docs/setup/tools.md. The
> audit files stay in the plan folder as the record of the audit.

> D7: Each finding gives the section, the kind (wrong, missing, extra,
> order), the current text of the doc (or 'none'), the evidence as
> file:line, and the full correct text that U6 writes. The evidence is in
> macos/tools/ or macos/packs.toml; for a finding of How a tool works it can
> also be in docs/design/ or cli/. A finding with no evidence is not
> written. A missing section and a section with no tool folder are findings
> too. The kind order covers a section out of sort order and an item out of
> the order of the legend inside one section. A new text of more than one
> line (a whole section, an item with a sub-list) goes in a fenced markdown
> block, indented under its finding, in place of the quoted new text.

> D8: New text follows the existing doc: the items of the legend in 'How a
> tool works' (Packs, Stage, Depends, Files, Tasks, Scripts, Contributes to,
> Defaults, Mise sections, Question), in that order, and the plain STE prose
> of the doc (ASD-STE100: short sentences, simple present, active voice). A
> 'why' rule is one or two sentences in the prose paragraph under the ###
> heading, after the first sentence. No secret value is written: a secret is
> named by its variable name only.

> D9: No unit changes a file of macos/tools/. An audit unit lists each stale
> comment that it finds in the '## Stale comments' section of its audit file.
> U6 does not apply them.

> D14: The evidence in macos/tools/ decides. When a part audit (U1 to U4) and
> the general audit (U5) give text for the same section, U6 writes one text
> that holds the facts of both, with no repeated sentence. When two findings
> contradict each other, U6 reads the cited lines and writes the text that the
> files support.

> D15: A claim with a part that has no evidence in the files is stated with
> the proven part only, and its line names the dropped part with the reason
> (for example C2, C4, C12). A claim that a bootstrap doc already states is
> stated — already in <doc> <section>, and that doc can be another doc than
> docs/setup/tools.md (for example C14, docs/design/mac.md). A claim that
> the doc states in part (for example C24) is stated with new text for the
> rest only, and its line names the section that already holds the first
> part.

> D16: Keep an item when the files support it. Each item kind that occurs in
> more than one section, and the forms Question `<key>`: and Contributes to
> `<tool>`:, gets a line in the legend of How a tool works: U5 finds these
> kinds with a grep of the whole doc and writes the legend lines. The form
> Question `<key>`: takes the place of Question, and Contributes to
> `<tool>`: the place of Contributes to; each other new kind goes after Mise
> sections and before Question, in the order that U5 gives in the legend.
> For a kind that occurs in one section only, the part unit of that section
> decides: keep it when it names a real part of the tool (for example Manual
> step:), else move its text into the prose paragraph of the section, as a
> finding of kind wrong. An item that the files do not support is a finding
> of kind wrong or extra.

## Edits

1. **`docs/setup/tools.md`** — apply each finding (`F<n>` of `a.md` to `d.md`,
   `G<n>` of `general.md`) and each `stated` checklist claim (`C<n>` with new
   text) of `general.md`:
   - `wrong`: replace the doc text with the `new` text.
   - `missing`: add the `new` text at its place: an item in the order of the
     legend, a "why" rule in the prose under the `###` heading, a general rule
     in "How a tool works" before the item legend, a missing section in sort
     order.
   - `extra`: remove the text, or the whole section of a tool with no folder.
   - `order`: move the section to its place in `LC_ALL=C` sort order of the
     tool names, or move the item to its place in the order of the legend.
   - Put each item kind that the legend did not name before into its place by
     the new legend lines of `general.md` (D16), also with no `order` finding.
   - Do not apply the "Stale comments" lists (D9), or a `stated — already in`
     claim.
   - When the count of the packs or another count in the prose changes, change
     the count too.
   - Keep the style of D8. Wrap prose at 80 columns, as the rest of the file.

## Verification

- `mise x -- mise run code:format --all` passes.
- This command prints nothing: one section for each tool folder, in sort
  order.

  ```bash
  diff <(LC_ALL=C command ls macos/tools) <(grep '^### ' docs/setup/tools.md | sed 's/^### `\(.*\)`$/\1/')
  ```

- For each finding and each `stated` claim with new text, a grep of a key
  word of its `new` text finds it in its section of `docs/setup/tools.md`.
- `git status --porcelain` shows only `docs/setup/tools.md`.

## Guardrails

- Do not edit the audit files, any file of `macos/tools/`, or another doc.
  A falsified passage of another doc is a `DOCS FALSIFIED:` line for U7.
- Do not add a fact that no finding gives. A gap that you see is a
  `QUESTION` line, or a `DECIDED` line when the files settle it.
- Write no secret value. A secret is named by its variable name only.
- Delete with `rm`, never `git rm`.

## Commit

`docs: make docs/setup/tools.md match macos/tools`
