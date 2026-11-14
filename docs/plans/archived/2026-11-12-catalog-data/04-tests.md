# U4 — Tests: the keys in each scratch tool and fixture

- **Read first:** `.config/mise/tasks/test/repo` (the helpers at the top and
  each place that writes a `tools.toml`), `.config/mise/tasks/test/mac`, and
  `macos/fixtures/`.
- **Lazy-load:** `CLAUDE.md` ("Tests", "Task scripts"),
  `.config/mise/tasks/test/macos-tools`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

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

## Edits

1. **`.config/mise/tasks/test/repo`** — give each scratch `tools.toml` a
   `summary` and a `category` that is a real id of the copied
   `repo/categories.toml`
   (for example `core`). Do not write or replace that file (D11 a). Prefer one
   helper that writes them. Add cases:
   a repo tool with no `summary` (code 2, "summary is missing"), with no
   `category` (code 2, "category is missing"), and with `category = ""`
   (code 2).
2. **`.config/mise/tasks/test/mac`** — give each scratch tool that the task
   writes a `summary` and a `category` that is an id of the categories file of
   its fixture clone, which U3 writes (D11 a). Do not write or replace that
   file. U3 changed the fixtures of `macos/fixtures/`. Add
   cases: a Mac tool with no `summary`, with no `category`, and with
   `summary = ""` (code 2 each).
3. **The categories cases of plan 1a** — keep each one. A case that tests the
   categories file itself deletes or overwrites the copy of the file inside
   its own clone only (D12 b).
4. **`.config/mise/tasks/test/macos-tools`** — the git cases of D4 (D10 c),
   with the real `macos/tools/` and a copy of `macos/examples/macos.toml`:
   `email_gitlab = ""` (exit 0; `identity` has the GitHub email,
   `gitlab.config` has no `email` line, `allowed_signers` has one signer line);
   `email_github = ""` (exit 0; `identity` has the GitLab email,
   `github.config` has no `email` line, `allowed_signers` has one signer line);
   both
   emails `""` (code 2, the group message of plan 1a).
5. Change each existing case whose expected output changes only because of
   the new keys.

Write the shell for bash 3.2 and bash 5.x (see `CLAUDE.md`, "Task scripts").

## Verification

- `mise x -- mise run test` passes: `test:repo`, `test:mac`,
  `test:macos-tools` and `test:go`.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Do not touch `cli/`, `repo/`, `macos/` or a doc.

## Commit

`test: give each test tool the catalog keys`
