# U1 — Repo data: categories, texts, selectable and hosts

- **Read first:** `texts.md` of this folder, `docs/design/tools.md` (the key
  table, the categories file), and
  `repo/packs.toml`.
- **Lazy-load:** each `repo/tools/<tool>/tools.toml` as you edit it.

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
- D3: "Ruling of the user (2026-11-12): selectable = false on
  repo/tools/_security_base only (gitleaks, grype and osv-scanner depend on it).
  hosts = ["macos"] on repo/tools/swift only. No Mac tool gets either key."
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

## Precondition

Read `cli/VERSION` first. If its version is below 2.1.0, stop and change
nothing. Report "cli/VERSION is <v>; plan 1b needs 2.1.0 or later" (D7).

## Edits

1. **`repo/categories.toml`** — create it with the comment and the TOML of
   `texts.md` ("The categories files"), in that order.
2. **Each `repo/tools/<tool>/tools.toml`** (36 files) — add `summary`,
   `category` and `description` with the values of `texts.md` ("Repo tools"),
   as top-level keys, after the header comment and next to the other
   top-level keys (`depends`, `parent`). Keep the header comment as it is.
3. **`repo/tools/_security_base/tools.toml`** — add `selectable = false`.
4. **`repo/tools/swift/tools.toml`** — add `hosts = ["macos"]`.
5. **`repo/packs.toml`** — add `description = "<text>"` as the first key of
   each `[pack.<name>]` table, with the texts of `texts.md` ("The pack
   descriptions of `repo/packs.toml`"). Keep the member order. Replace the
   comment at the top with the text of `texts.md` ("The comments of the packs
   files").

Use a TOML string with double quotes. Escape a double quote in a text, if
there is one.

## Verification

- Build the engine (`cd cli && go build -o <tmp>/bootstrap ./cmd/bootstrap`)
  and run `<tmp>/bootstrap check --clone <worktree>` in the main checkout:
  exit 0.
- Do not run `test:repo` as a check of this unit: its categories cases and
  its scratch tools fail until U4 (D12 a).
- `mise x -- mise run code:format` passes.

## Guardrails

- Change no text of `texts.md`. Write no text of your own.
- Do not touch `macos/`, `cli/`, the test tasks or a doc.
- Do not change `cli/internal/source/testdata/toml-yq.golden`: it is U3's
  (the hashes of each `tools.toml` change; U3 writes the golden again).

## Commit

`feat: add the catalog data of the repo tools and packs`
