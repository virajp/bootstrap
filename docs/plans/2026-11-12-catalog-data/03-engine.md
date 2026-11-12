# U3 — Engine: summary and category become necessary

- **Read first:** `docs/design/cli.md` ("The exit codes"),
  `docs/design/tools.md` (the key table), then
  `cli/internal/source/check.go` (search `summary` and `category`: the rules
  of plan 1a).
- **Lazy-load:** `cli/internal/source/notools_test.go`, the Go test files of
  the facts section, `cli/internal/source/testdata/`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D5: "Ruling of the user (2026-11-12): on both sides, a tools.toml with no
  summary, or with no category, stops the sync with code 2. The message has the
  form of the existing messages of check.go: "<tools.toml shown>: summary is
  missing" and "<tools.toml shown>: category is missing". Each other rule of
  plan 1a stays: 47 characters or fewer, a string, a known category id."
- D8: "Assumed by the planner from D7 of plan 1a: U3 reads the keys generically.
  No string literal of the engine names a tool. TestEngineNamesNoTool stays
  green with no new entry in its allow list."
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

## Edits

1. **`cli/internal/source/check.go`** — on both sides, a `tools.toml` with no
   `summary`, or with `summary = ""`, is a problem with code 2: "<tools.toml
   shown>: summary is
   missing". The same for `category` (also `category = ""`): "<tools.toml
   shown>: category is
   missing". Keep each rule of plan 1a.
2. **The Go tests** — give each `tools.toml` text of a Go test a `summary`
   and a `category`, and a categories file where the test opens a clone. Add
   a unit test for each new problem, on both sides.
3. **`macos/fixtures/`** — give each fixture `tools.toml` a `summary` and a
   `category`, and add a `categories.toml` to each fixture clone that names
   those categories (for example `macos/fixtures/repo/categories.toml`), so
   `test:mac` still opens the fixtures (D10 a).
4. **The goldens** — write them again with `UPDATE_GOLDEN=1 go test ./...` in
   `cli/`, and examine each changed golden. The hashes of
   `cli/internal/source/testdata/toml-yq.golden` change because U1 and U2
   changed each `tools.toml` and the packs files, and step 3 changed the
   fixtures. Write the
   golden after step 3.

## Verification

- `mise x -- mise run test:go` passes, `TestEngineNamesNoTool` included.
- Build the engine and run `<tmp>/bootstrap check --clone <worktree>` in the
  main checkout and `<tmp>/bootstrap mac sync --dry-run --clone <worktree>`:
  each exits 0.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Do not touch the test tasks, `repo/`, `macos/tools/`, `macos/packs.toml`,
  `macos/categories.toml`, `macos/examples/` or a doc.
- Do not edit `cli/VERSION`. Add no Go module.
- No string literal of the engine names a tool.
- `test:repo` and `test:mac` can fail after this unit: U4 gives their scratch
  tools the keys.

## Commit

`feat: make summary and category necessary keys of a tool`
