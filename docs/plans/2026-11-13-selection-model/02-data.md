# U2 — Data: the new packs, the descriptions and the examples

- **Read first:** `texts.md` of this folder, `docs/design/packs.md`, and the
  branch `packs-shortcut` (`git show --stat packs-shortcut`).
- **Lazy-load:** `macos/examples/macos.toml`, `repo/examples/*/bootstrap.toml`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D4: "Ruling of the user (2026-11-13): each member of a pack is a default. A
  member is <tool> = {}, or in [pack.base] only <tool> = { required = true }; a
  nested pack is "pack:<name>" = {}. The key default is gone: a default key, a
  required key outside base, or another key in a member table is code 2.
  description = "<text>" is the reserved key of a pack (plan 1a). A tool can be
  in no pack. The engine reads the new form only (no second reader for the form
  of today). The member order is the order of the file."
- D5: "Ruling of the user (2026-11-13): a values file of the new form has the
  top-level key format = 2 (repo and Mac). init writes it in a new file and in a
  migrated file. A file without it is an old form: sync and check stop with code
  2 and the hint of today ("run bootstrap init" / the Mac hint), through the
  package oldform. init migrates it: it expands each [packs.<name>] table, and
  the implicit pack base, with a frozen copy of repo/packs.toml and
  macos/packs.toml as they are on main before this plan (the rules of today:
  only the default = true members, base always added, nested packs expanded),
  writes a [tools.<name>] table for each such tool that has no table, removes
  the [packs.*] tables, and writes format = 2. A [tools.<name>] table of a tool
  that no longer exists (for example devtools) is removed by the migration,
  which lists it. The frozen copy is a data file in the engine (embedded with
  go:embed), not a Go literal. After the migration, the new packs.toml applies
  to add pack only."
- D10: "Ruling of the user (2026-11-13): this plan lands the local branch
  packs-shortcut (420e3e6): repo/packs.toml (base only), macos/packs.toml (the
  new packs), macos/tools/brave-browser/tools.toml (CHROME_EXECUTABLE) and the
  removal of macos/tools/devtools/. U2 copies each file from the branch (git
  show packs-shortcut:<path>), adds the accepted description of each pack
  (texts.md of this folder) as the first key of its [pack.<name>] table, and
  replaces the comment at the top of each packs file with the comment of
  texts.md. The example values files (macos/examples/macos.toml,
  repo/examples/*/bootstrap.toml) get format = 2, no [packs.*] and no
  [tools.devtools]."
- D15: "Assumed by the planner from D1-D14: (a) U2 adds only the [mise.env]
  CHROME_EXECUTABLE block of the branch to macos/tools/brave-browser/tools.toml
  of the tree, so the summary and category of plan 1b stay; the packs files come
  whole from the branch (plan 1b does not change them, its D13). (b) The
  migration and add pack write a table for each new dependency too (D3); K1 and
  K2 show them. (c) The migration skips a member of the frozen copy that the
  clone does not have (for example devtools), and does not list it. (d) U1 owns
  the migration and the base tables that it writes; U4 owns init: the base
  tables of a new file and of the --no-tui path. U1 does not change finish or
  init.go. (e) U3, U1 and U4 verify with go test of their packages and skip
  TestTOMLMatchesYq (U5 writes cli/internal/source/testdata/toml-yq.golden again
  after its fixture edits, as the last writer of a .toml in the run; shared file
  S3). (f) The migrate item of this plan is M38 (plan 1b adds M37). (g) U2 owns
  four falsified passages and writes this text:
  repo/tools/mise/tasks/setup-configs.tmpl:40-41 "The tools come from the
  [tools.*] tables of .config/bootstrap.toml and from the required members of
  the pack base."; repo/tools/mise/tasks/code-sec:25 "The scanners of the pack
  base must be there"; repo/tools/_security_base/tools.toml:1 "_security_base:
  the code:sec orchestrator and the sec hook of the scanners gitleaks, grype and
  osv-scanner."; and macos/tools/claude-cli/files/.claude/CLAUDE.md:36: the list
  of folders names each folder of macos/tools/ whose [mise.shell_alias] defines
  one of the aliases of that table (U2 reads them), without devtools. (h) K1 and
  K2 unset MISE_PROJECT_ROOT and MACOS_SETUP_REPO_PATH before they run. (i) The
  docs check of U7 is: grep -rnE "detect --no|bootstrap (mac
  )?(add|remove|replace) [a-z]" docs/design docs/sync docs/setup README.md
  CLAUDE.md | grep -vE "(add|remove|replace) (tool|pack) |migrate.md" finds
  nothing."
- D17: "Assumed by the planner from D1-D16: (a) U1 also owns dprint.jsonc and
  adds an exclude for cli/internal/oldform/packs-v2/, as for repo/packs.toml and
  macos/packs.toml, so the frozen copies stay byte for byte. (b) env in the
  migration: when one source of a tool has no env (the implicit base, or a pack
  table with no env), the new table has no env (all the default environments);
  else the union of the env lists (D16 b). (c) Apply restores the values file
  only (D14 c); the other files of an old-form migration stay, and a second init
  accepts that state (each migration step is idempotent). (d) detect: --yes and
  --answer need --add, else code 2 with the hint "add --add"; --dry-run works
  with or without --add. (e) Plain mac detect exits 1 when a detected tool or an
  own installed package is not selected, as mac detect --no does today. (f)
  --exclude takes a member of P or a tool that a nested pack of P adds. (g) add
  pack skips, with an info line, a member whose hosts leave out this host or
  that has selectable = false, as D16 (d) does for root_only. (h) A file with
  format = 2 and a [packs.*] table: init migrates the [packs.*] tables with the
  frozen copy and does not add the tables of base again (D16 c); sync and check
  stop with code 2 until then. (i) init with names: code 2 and the hint "use:
  bootstrap add tool <names>" (D8). (j) list shows a required member of base
  with no table as a root, marked required. (k) U5 owns
  cli/internal/source/testdata/toml-yq.golden and nothing else of cli/. (l) D15
  (g) names four passages. (m) The review unit covers D15 and D16 too."
- D18: "Ruling of the user (2026-11-13), with the details the planner assumed
  from D1-D17: (a) The values file holds tools only and never a pack, so a later
  change of a pack other than base reaches no values file. The migration writes
  tables for the tools of the frozen copy only (the frozen base and each old
  [packs.*] table) and their dependencies (D15 b). It writes no table for a
  non-required member of the new base (for example visual-studio-code, python
  and uv on the Mac); such a member comes only into a new file (D16 c). (b)
  sync, on both sides, writes an empty [tools.<name>] table into the values file
  for each required member of base that has no table, so the user can add
  overrides; this also applies when the user deleted the table by hand, and it
  replaces that part of D3. It skips a member that hosts, selectable = false or
  root_only in a submodule leaves out (D16 a, d; D17 g). It uses the
  comment-keeping write of the migration, before it syncs the files, and logs
  one line for each table. A dependency that has no table gets no table from
  sync: D3 stays for it. (c) sync --dry-run and check write nothing; they list
  each table that sync will write. check exits as it does for a file that
  differs. (d) The example values files hold a table for each required member of
  base, so a sync of an example writes no table. (e) U5 adds the cases:
  required = true in a pack other than base is code 2 (repo and Mac); sync
  writes the
  table of a required member with no table and logs it; sync --dry-run and check
  do not write it; a non-required member of the new base gets no table in a
  migration."
- D20: "Ruling of the user (2026-11-13) on the env of the deleted Mac tool
  devtools (rulings review round 3, gap 4): (a) DOTNET_CLI_TELEMETRY_OPTOUT is
  dropped; a later .NET tool brings it. (b) RUNTIME_ENV is dropped; each
  repository sets it, and no bootstrap tool sets it. (c) PRE_COMMIT_HOME goes to
  a new Mac tool macos/tools/pre-commit/, a member of [pack.base] of
  macos/packs.toml with { required = true }, after vorssaint. Its tools.toml,
  with the bootstrap header comment of the other Mac tools: depends = [ "python"
  ], the summary "pre-commit through mise, with a shared cache", category =
  "dev", [mise.tools.pre-commit] version = "latest", and [mise.env]
  PRE_COMMIT_HOME = "$HOME/.cache/pre-commit". The version, the depends and the
  env value are those of repo/tools/pre-commit/tools.toml, because
  test:macos-tools needs one value on both sides. macos/examples/macos.toml gets
  a [tools.pre-commit] table (D18 d). docs/setup/tools.md gets a section for the
  tool (U7)."
- D21: "Assumed by the planner from D1-D20: (a) The check of U2 with git diff
  packs-shortcut also accepts the member line pre-commit = { required = true }
  of D20 (c). (b) The comment at the top of macos/packs.toml in texts.md names
  bootstrap mac add pack (D1); the orchestrator corrected the text. (c) U2 also
  owns repo/tools/pnpm/scripts/migrate-aube and
  repo/tools/aube/scripts/migrate-pnpm, and changes "bootstrap replace" in their
  top comments to "bootstrap replace tool" (D8). (d) D18 (b) and D9 agree: Plan
  adds each missing table of a required member of base to the new values bytes,
  so a command writes the values file one time; the sync that Apply runs then
  finds no missing table. Only a plain sync writes the tables itself, in its one
  write of the values file."
- D22: "Ruling of the user (2026-11-13) on rulings review round 4, gap 2: the
  required Mac tool pre-commit depends on python, so python becomes a required
  member of the Mac base: python = { required = true } in [pack.base] of
  macos/packs.toml (U2). The Mac base description of texts.md names Python with
  the tools that a Mac cannot remove, and only VS Code and uv as tools that the
  user can remove; the orchestrator corrected the text. The git diff
  packs-shortcut check of U2 also accepts this changed member line."

## Edits

1. **The branch files** — write each file of the branch into the worktree:
   `git show packs-shortcut:repo/packs.toml > repo/packs.toml`, the same for
   `macos/packs.toml`, and `rm -r macos/tools/devtools`. In
   `macos/tools/brave-browser/tools.toml`, add only the `[mise.env]` block of
   the branch (D15 a). If the branch does not exist, stop and report
   it.
2. **The descriptions** — add `description = "<text>"` as the first key of
   each `[pack.<name>]` table of both packs files, with the texts of `texts.md`.
3. **The comments** — replace the comment at the top of each packs file with
   the comment of `texts.md`.
4. **The examples** — in `macos/examples/macos.toml` and each
   `repo/examples/**/bootstrap.toml`: add `format = 2` next to
   `bootstrap_commit`, remove each `[packs.*]` table and `[tools.devtools]`, add
   a table for each tool that a removed pack table selected (new form) and for
   each required member of the new base (D18 d), and
   change the comment that describes `[packs.<name>]` (macos.toml:5) to say
   that the file holds tools only.

5. **The Mac tool `pre-commit`** — create `macos/tools/pre-commit/tools.toml`
   and add the member to `[pack.base]` of `macos/packs.toml`, as D20 (c) gives.
   Change the member `python` of that table to `{ required = true }` (D22).
   Do not carry `DOTNET_CLI_TELEMETRY_OPTOUT` or `RUNTIME_ENV` (D20 a, b).

6. **The migrate scripts** — D21 (c): in the top comments of
   `repo/tools/pnpm/scripts/migrate-aube` and
   `repo/tools/aube/scripts/migrate-pnpm`,
   `bootstrap replace` becomes `bootstrap replace tool`.

7. **The falsified passages** — the four passages of D15 (g), with the text
   that D15 gives.

## Verification

- `git diff packs-shortcut -- repo/packs.toml macos/packs.toml` shows only the
  descriptions, the comments and the member lines `pre-commit` of D20 (c) and
  `python` of D22
  (D21 a).
- `mise x -- mise run code:format` passes.
- Do not run `test:repo`, `test:mac` or `test:macos-tools` as a check of this
  unit: they need the engine of U1, U3 and U4, and U5 changes them.

## Guardrails

- Change no text of `texts.md`. Write no text of your own. The one exception is
  the tool `pre-commit` of D20 (c): its summary, and a top comment in the form
  of the other Mac tools.
- Do not touch `cli/`, the test tasks or a doc.

## Commit

`feat: reshape the packs as shortcuts and describe each pack`
