# U6 — Docs

- **Read first:** `docs/design/tools.md` (the key table, the categories file,
  the questions), `docs/setup/tools.md` ("git"), `docs/sync/migrate.md` (the
  last items).
- **Lazy-load:** `CLAUDE.md` ("Change a shared file"), `docs/sync/index.md`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D3: "Ruling of the user (2026-11-12): selectable = false on
  repo/tools/_security_base only (gitleaks, grype and osv-scanner depend on it).
  hosts = ["macos"] on repo/tools/swift only. No Mac tool gets either key."
- D4: "Ruling of the user (2026-11-12): email_github and email_gitlab of
  macos/tools/git/tools.toml get group = "email"; email_github gets note = "Give
  at least one email. The SSH key that signs your commits is the same for GitHub
  and GitLab."; email_gitlab gets no note. The templates drop required on the
  two emails: identity.tmpl writes email = email_github when it is not empty,
  else email_gitlab; github.config.tmpl writes no email line when email_github
  is empty; gitlab.config.tmpl writes no email line when email_gitlab is empty;
  allowed_signers.tmpl writes no line for an empty email. name and signing_key
  keep required. A missing key and "" are the same empty value in a template."
- D5: "Ruling of the user (2026-11-12): on both sides, a tools.toml with no
  summary, or with no category, stops the sync with code 2. The message has the
  form of the existing messages of check.go: "<tools.toml shown>: summary is
  missing" and "<tools.toml shown>: category is missing". Each other rule of
  plan 1a stays: 47 characters or fewer, a string, a known category id."
- D9: "Assumed by the planner from the standing rule that a contract change
  needs a migrate.md item: U6 adds the item M37 to docs/sync/migrate.md. A
  values file that names [tools._security_base] directly gets the warning of
  check; remove the table, because gitleaks, grype or osv-scanner selects the
  tool. A macos.toml needs no change: an empty email is now valid."
- D13: "Ruling of the user (2026-11-13), which replaces the pack parts of D2 and
  D11 b: this plan does not change repo/packs.toml or macos/packs.toml. It
  writes no pack description and no new comment there. The user reshaped the
  packs on the local branch packs-shortcut (420e3e6): new names, members with no
  default key, and a base whose members are mostly required. The engine of today
  refuses that form, so the branch lands with plan 2, which also writes the
  descriptions of the new packs. The sections "The comments of the packs files"
  and "The pack descriptions" of texts.md are not used."

## Edits

1. Run `vwf:docs-sync` over the branch delta of the run. Apply its findings,
   and each `DOCS FALSIFIED:` line of U1 to U4.
2. **`docs/design/tools.md`** — `summary` and `category` are necessary now
   (D5), with the two messages. Remove the sentence that says they become
   necessary in a later change.
3. **`docs/setup/tools.md`** ("git") — one email is enough: the group
   `email`, the note, and the files that each empty email changes (D4).
4. **`docs/sync/tools/swift.md`** — the tool runs only on macOS
   (`hosts = ["macos"]`). **`docs/sync/tools/_security_base.md`**, if the file
   exists — the tool is not selectable; a scanner selects it (D3).
5. **`docs/sync/migrate.md`** — add the item M37 of D9.
6. **`README.md`, `CLAUDE.md`, `docs/sync/index.md`** — only where a passage
   says that a key is optional or that each question needs an answer.
7. **`CLAUDE.md`** ("Tests") — the items that `test:macos-tools` examines:
   add the git cases of one empty email and of two empty emails (D10 c).

Write the docs in ASD-STE100 Simplified Technical English, in the style of the
existing text.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'M37' docs/sync/migrate.md` finds the item.

## Guardrails

- Do not touch `cli/`, the test tasks, `repo/` or `macos/`.
- Do not edit a synced copy (see `CLAUDE.md`).

## Commit

`docs: describe the catalog data and the one-email git rule`
