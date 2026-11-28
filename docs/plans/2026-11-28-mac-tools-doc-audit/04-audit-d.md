# U4 — Audit part D: starship to the last tool

- **Read first:** `docs/setup/tools.md` "How a tool works" (the item
  legend), then each `### <tool>` section of part D, top to bottom.
- **Lazy-load:** `docs/design/tools.md` (the schema of `tools.toml`),
  `docs/design/mac.md` (the stage and the entry tasks), `macos/packs.toml`.

## Ruling

> D1: Audit docs/setup/tools.md against the current macos/tools/. The Notes of
> the deleted macos-setup docs/tool-reference.md (lines 71-201 at commit
> a4ec825^) are a checklist, not a source: each claim that is still true must
> be in the doc, and a false claim is dropped.

> D2: All the ### <tool> sections of docs/setup/tools.md (110 on 2026-11-28),
> plus its sections 'How a tool works' and 'The packs'. Each section is
> compared with tools.toml, files/, tasks/ and scripts/ of its tool folder,
> and 'The packs' with macos/packs.toml.

> D3: Each fact goes in the ### section of the tool that owns it. A rule for
> many tools goes in 'How a tool works', as a paragraph before the item legend
> ('Each section below gives these items for one tool:').

> D4: It requires docs/plans/2026-11-27-tool-skills, which requires
> 2026-11-27-tool-selection and 2026-11-27-mac-backup-target. The audit then
> examines the final state of macos/tools/ after the three plans. Each claim of
> the survey is examined again at run time, because
> those plans can change it.

> D5: Audit in parallel, then apply once: U1 to U5 (wave 1) each write one
> findings file in docs/plans/2026-11-28-mac-tools-doc-audit/audit/ and edit
> nothing else; U6 (wave 2) applies the five files to docs/setup/tools.md.

> D6: By the name of the tool folder, in LC_ALL=C sort order: part A is each
> name up to and including dprint; part B is each name from
> duplicate-detective up to and including mas; part C is each name from
> microsoft-excel up to and including ssh; part D is each name from starship
> on. A tool that a later plan adds or removes falls in its range by sort
> order. A tool with a folder and no section, or a section and no folder,
> belongs to the part of its name.

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

> D10: No. U1 to U5 read docs/setup/tools.md, macos/tools/,
> macos/packs.toml, the docs of docs/design/ and, for How a tool works,
> cli/; they write only their own audit file.

> D11: teams. An audit unit that finds a fact that belongs to another part can
> send it to the unit of that part. Talk is a help only: each unit file stands
> alone.

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

1. **`docs/plans/2026-11-28-mac-tools-doc-audit/audit/d.md`** — create it.
   Get the tools of part D with
   `LC_ALL=C command ls macos/tools` and the range of D6. For each tool,
   compare its `### <tool>` section with its folder, item by item:
   - the first prose sentence, against the `description` of `tools.toml`
     and what the files and tasks do;
   - **Packs** against `macos/packs.toml` (`(required)` included);
   - **Stage** against `stage` of `tools.toml`;
   - **Depends** against `depends`;
   - **Files** against `[files]`: each target, its mode, and the `format`
     and `dir_perm` keys;
   - **Tasks** against `[tasks]` (`library` for a file of `_scripts/`);
   - **Scripts** against `scripts/`;
   - **Contributes to**, **Defaults**, **Mise sections** and **Question**
     against their tables;
   - the "why" of the tool: a behavior that a file or a task has and that a
     reader cannot see from the item list, for example an order of drop-in
     files, a reason in a comment, a step that runs only on a condition. Write
     it as one or two sentences (D8).

   An item that the legend does not name follows D16: compare it with the
   files like the other items, and for a kind that occurs in one section only,
   keep it or move its text into the prose.

   Also find each folder of the range with no section, each section of the
   range with no folder, and each section out of sort order.

   Write the file in this form:

   ```markdown
   # Audit part D — <first tool> to <last tool>

   ## Findings

   ### `<tool>`

   - F<n> <kind: wrong | missing | extra | order> <item or "prose"> —
     doc: "<the current text, or none>" — evidence: <path:line[, path:line]>
     — new: "<the full text that U6 writes>"

   ## Stale comments

   - <path:line> — <what the comment says> — <the fact, with its path:line>

   ## No change

   - `<tool>`, `<tool>`, … (each section of the part that matches)
   ```

   Number the findings F1, F2, … across the file. Each tool of the part is
   under "Findings" or in "No change", and in no other place. A tool with
   no finding goes in "No change", also when it has a stale comment. Write
   "none" under a heading with no item.

   An item kind that the legend does not name today has no legend place
   yet: U5 gives it one (D16). Do not write an `order` finding for such an
   item; U6 puts it in its place by the legend of U5.

## Verification

- `mise x -- mise run code:format --all` passes.
- Each tool of part D (`LC_ALL=C command ls macos/tools`, the range of D6)
  is named exactly once in the audit file, under "Findings" or in "No change".
- `git status --porcelain` shows no change outside
  `docs/plans/2026-11-28-mac-tools-doc-audit/audit/`, and this unit changed
  only its own audit file.

## Guardrails

- Do not edit `docs/setup/tools.md`, any file of `macos/tools/`, or the
  audit file of another unit. Do not read `cli/`: the evidence of a tool
  section is in `macos/tools/` and `macos/packs.toml` (D7).
- Do not audit a tool outside part D. A fact for another part goes to that
  unit by a message, or into the finding of the tool of this part that owns it.
- Write no secret value. A secret is named by its variable name only.
- Read a template (`*.tmpl`) as a template: a key that only renders under a
  condition is a conditional fact, and the new text says the condition.
- Delete with `rm`, never `git rm`.

## Commit

`docs: audit the Mac tools starship to the last tool against macos/tools`
