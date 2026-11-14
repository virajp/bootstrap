# U2 — Mac data: categories, texts and the git tool

- **Read first:** `texts.md` of this folder, `docs/design/tools.md` (the key
  table, the questions), `docs/design/mac.md`,
  then `macos/tools/git/tools.toml` and the four templates of
  `macos/tools/git/files/.config/git/`.
- **Lazy-load:** each `macos/tools/<tool>/tools.toml` as you edit it,
  `macos/examples/macos.toml`, `docs/setup/tools.md` ("git").

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D1: "Ruling of the user (2026-11-12): repo/categories.toml has 10 categories
  (core, format-lint, editor-ai, languages, package-managers, security, secrets,
  containers, cloud, forge) and macos/categories.toml has 14 (system, shell,
  cli, dev, editors, ai, browsers, passwords, productivity, office, notes,
  media, chat, disk), in that order; the file order is the order of the TUI.
  Each of the 36 repo tools and each of the 109 Mac tools has one category, as
  texts.md gives it: the accepted mapping of plan 1a, plus the addition of
  2026-11-12 (cli += bat, coreutils, curl, doggo, eza, fzf, gping, httping,
  ipcalc, jq, prettyping, ripgrep, zoxide; dev += lazygit, ruby, uv, worktrunk;
  passwords += age; fonts stays in system)."
- D2: "Ruling of the user (2026-11-12): the texts of texts.md, which the planner
  wrote and the user accepted. The units copy them as they are and write no text
  of their own. Each repo tool gets summary and description; each Mac tool gets
  summary, and its description stays as it is; each pack of both sides gets
  description (the reserved key of plan 1a). The first line of the header
  comment of a repo tools.toml stays as it is."
- D4: "Ruling of the user (2026-11-12): email_github and email_gitlab of
  macos/tools/git/tools.toml get group = "email"; email_github gets note = "Give
  at least one email. The SSH key that signs your commits is the same for GitHub
  and GitLab."; email_gitlab gets no note. The templates drop required on the
  two emails: identity.tmpl writes email = email_github when it is not empty,
  else email_gitlab; github.config.tmpl writes no email line when email_github
  is empty; gitlab.config.tmpl writes no email line when email_gitlab is empty;
  allowed_signers.tmpl writes no line for an empty email. name and signing_key
  keep required. A missing key and "" are the same empty value in a template."
- D7: "Ruling of the user (2026-11-12, the split of plan 1): this plan lands
  after the user tags bootstrap@2.1.0, the release of plan 1a. The first step of
  U1 and of U2 reads cli/VERSION; if it is below 2.1.0, the unit stops and
  changes nothing, and reports "cli/VERSION is <v>; plan 1b needs 2.1.0 or
  later". A consumer with an older engine refuses the new keys, so the data
  needs the released engine."
- D10: "Assumed by the planner from D1-D6 and the rule of the simplest
  consistent CLI rule: (a) U3 owns macos/fixtures/ as well as cli/: it gives
  each fixture tool the keys, adds the categories file of each fixture clone,
  and then writes cli/internal/source/testdata/toml-yq.golden again, which
  hashes the fixtures (cli/internal/source/toml_test.go:17-49); U4 owns the test
  tasks only. (b) Each categories.toml starts with a plain comment, in the form
  of the comment of packs.toml (the sync never copies the file, so it has no
  Synced header); texts.md gives the text. (c) The git cases of D4 run in
  test:macos-tools, which syncs the real macos/tools/ with a copy of
  macos/examples/macos.toml: one case with email_gitlab = "", one with
  email_github = "", one with both empty (code 2). (d) summary = "" and
  category = "" count as missing: code 2 with the message of D5."
- D11: "Assumed by the planner from D1, D2 and D10: (a) A scratch tool of
  test:repo names a real id of the copied repo/categories.toml (for example
  core); U4 never writes or replaces that file. A scratch tool of test:mac names
  an id of the categories file of its fixture clone, which U3 writes. (b) The
  first comment lines of repo/packs.toml and macos/packs.toml change to the text
  that texts.md gives ("The comments of the packs files"), because the key
  description is not a member; U1 and U2 copy it. (c) toml-yq.golden hashes each
  .toml below repo/ and macos/ (cli/internal/source/toml_test.go:36), so the
  edits of U1 and U2 change it too; U3 writes it again after U1, U2 and its own
  fixture edits. (d) "allowed_signers has one line" means one signer line; the
  two Synced header lines do not count."
- D12: "Assumed by the planner from D6 and D11: (a) test:repo and test:mac can
  fail between W1 and W3: the cases of plan 1a for the categories file and each
  scratch tool without the new keys meet the data of U1 and U2 and the rule of
  U3. U1 and U2 do not run them; the plan gate runs them after U4. (b) D11 a
  applies to the shared setup of the test tasks. A case that tests the
  categories file itself (no file, an unknown id, a bad table, a valid file) may
  delete or overwrite the copy of the file inside its own clone, and no other.
  U4 never changes repo/categories.toml or a file of macos/fixtures/ in the
  working tree. Each case of plan 1a for the categories file stays, rewritten
  that way."
- D13: "Ruling of the user (2026-11-13), which replaces the pack parts of D2 and
  D11 b: this plan does not change repo/packs.toml or macos/packs.toml. It
  writes no pack description and no new comment there. The user reshaped the
  packs on the local branch packs-shortcut (420e3e6): new names, members with no
  default key, and a base whose members are mostly required. The engine of today
  refuses that form, so the branch lands with plan 2, which also writes the
  descriptions of the new packs. The sections "The comments of the packs files"
  and "The pack descriptions" of texts.md are not used."

## Precondition

Read `cli/VERSION` first. If its version is below 2.1.0, stop and change
nothing. Report "cli/VERSION is <v>; plan 1b needs 2.1.0 or later" (D7).

## Edits

1. **`macos/categories.toml`** — create it with the comment and the TOML of
   `texts.md` ("The categories files"), in that order.
2. **Each `macos/tools/<tool>/tools.toml`** (109 files) — add `summary` and
   `category` with the values of `texts.md` ("Mac tools"), next to
   `description`. Keep `description` as it is.
3. **`macos/packs.toml`** — do not change it (D13).
4. **`macos/tools/git/tools.toml`** — add `group = "email"` to
   `[questions.email_github]` and `[questions.email_gitlab]`, and the `note`
   of D4 to `[questions.email_github]` only.
5. **The git templates** in `macos/tools/git/files/.config/git/` (D4):
   - `identity.tmpl`: replace the `email` line with a value that is
     `email_github` when it is not empty, else `email_gitlab`, with no
     `required`. Treat a missing key and `""` the same (for example
     `index .config "email_github" | default ""`).
   - `github.config.tmpl`: write the `email` line only when `email_github` is
     not empty. `gitlab.config.tmpl`: the same for `email_gitlab`.
   - `allowed_signers.tmpl`: write each line only when its email is not
     empty.
   - Keep `required` on `name`, `signing_key` and `projects_dir`.
6. **`macos/examples/macos.toml`** — keep both email answers. If the file has
   a comment that says each question needs an answer, change it to name the
   group rule.

## Verification

- Build the engine and run `<tmp>/bootstrap mac sync --dry-run --clone
  <worktree>`: exit 0, and the four git files do not change when both emails
  have a value.
- With a copy of `macos/examples/macos.toml` where `email_gitlab = ""`, a
  `mac sync` into a temporary `HOME` writes `identity` with the GitHub email,
  `gitlab.config` with no `email` line, and `allowed_signers` with one signer
  line.
  Do the same with `email_github = ""`.
- `mise x -- mise run test:macos-tools` passes. Do not run `test:mac` as a
  check of this unit (D12 a).
- `mise x -- mise run code:format` passes.

## Guardrails

- Change no text of `texts.md`. Write no text of your own.
- Do not touch `repo/`, `cli/`, `macos/fixtures/`, the test tasks or a doc.

## Commit

`feat: add the catalog data of the mac tools and the one-email git rule`
