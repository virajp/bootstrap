# U6 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (the cases that name the values
  file, about 60), `.config/mise/tasks/test/_default`,
  `.config/mise/tasks/test/go`.
- **Lazy-load:** `.config/mise/tasks/test/mac`,
  `.config/mise/tasks/test/macos-tools`.

## Ruling

D4, D5, D10 and D11 (quoted in `02-migration-config-verb.md`), D7 (quoted in
`03-members.md`) and D8 (quoted in `01-folder-schema.md`) are what the cases
test.

### Rulings review, round 1

These rulings override any earlier text of this file that they contradict.

- **D21** — User: "A config get is not a failure so it must emit either the
  default (with status showing default) or the actual value but never an error
  code. Other verbs/commands will continue raising error of missing config".
  config get of a valid key exits 0 and prints {"key": <path>, "value": <v>,
  "source": "file" | "default"} (the same three fields in toml and yaml); a key
  with no schema default prints value null with source default. Only an invalid
  key (code 2) or a read failure (code 1) is an error. Scripts read .value.
- **D34** — Flat, as today: repo/examples/<x>/repo.toml and
  repo/examples/<x>/tools.toml. U6 (test/repo:2780) and U8 (CLAUDE.md:59) follow
  it.

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D47** — init stops with code 2, naming the table, with the fix line "remove
  the table or create the branch". The migration never drops it silently.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D61** — Exit 0 with the default (D21).
- **D62** — They are flags of the setup:configs task, not of the engine. There
  is no init --all: the user runs bootstrap init in each member. D54 reads:
  setup:configs --all skips an old-form member with the warning "member <name>:
  run bootstrap init there" and continues.

### Rulings review, round 4

These rulings override any earlier text of this file that they contradict. D20
no longer has a version check (D84).

- **D84** — D20's version check is removed. A script calls config get; when an
  old engine answers with a usage error (exit 2), the task stops with the
  install hint. No version comparison.

## Edits

1. **`test/repo`.** Every case that writes or reads `.config/bootstrap.toml`,
   `.config/bootstrap.overrides.toml` or `.config/repo-generated.tsv` in a
   scratch consumer uses the folder form. Keep each case's intent.
2. **New cases in `test/repo`:**
   - `init` on an old-form file (with `[tools.git.branch.*]`, `commit_scopes`,
     `worktree_files`, `[member.*]` and an overrides file) produces the folder
     per D15 and D11, and a second `init` changes nothing.
   - `sync` and `check` on an old-form file exit 2 with "Run bootstrap init".
   - `config get|set|unset|show` in each format; `set` of an unknown key exits
     2; `get` of an unset key exits 0 and prints its default with `source:
     "default"` (D61); `set releases.x PATCH` then `MINOR` then
     `PATCH` keeps `MINOR`.
   - `settings` prints the engine settings; `config` alone exits 2.
   - A sibling member that is present gets `[home]` after `mise run
     setup:configs --all` (D62, D63); a
     missing one is skipped with an info line.
   - The git merge task honours `landing` (`pr` opens no direct merge; the
     protected-branch arguments leave out a `direct-commit` branch).
3. **`test/mac` and `test/macos-tools`:** no change unless a case reads a repo
   values file; report any such case.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/`.
- Do not edit `cli/`, `repo/` or the docs: a failure that needs such an edit is
  a `QUESTION`.
- Delete with `rm`, never `git rm`.

## Commit

`test: cases for the config folder, the config verb and members`
