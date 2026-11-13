# U1 — Engine: the packs form, the tables of each tool, the format key and the migration

- **Read first:** `docs/design/packs.md`, `docs/design/bootstrap-config.md`,
  then `cli/internal/source/packs.go`, `cli/internal/selection/selection.go`,
  `cli/internal/selection/repo.go`, `cli/internal/selection/macos.go`,
  `cli/internal/oldform/oldform.go`, top to bottom.
- **Lazy-load:** `cli/internal/cli/list.go`, `cli/internal/cli/init.go`,
  `cli/internal/cli/select.go`, `cli/internal/source/notools_test.go`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D2: "Ruling of the user (2026-11-13): the values file keeps no record of a
  pack. add pack P [--exclude m]... takes each member of P (each member is a
  default), expands each nested pack "pack:Q" the same way, removes each
  --exclude member, and writes a [tools.<name>] table for each of these tools
  that is not selected yet and for their new dependencies. --exclude takes a
  member key as packs.toml writes it, also a nested pack ("pack:security").
  --exclude needs exactly one pack name: add pack a b --exclude x is code 2; add
  pack a b with no --exclude is valid. A name that is not a member of P, or
  --exclude on a required member, is code 2. There is no remove pack and no
  --include. add pack P again adds only the members that are not selected."
- D3: "Ruling of the user (2026-11-13): each selected tool has a [tools.<name>]
  table, also each required member of base and each dependency, so a user can
  write overrides in it. A tool is selected when it has a table, or a selected
  tool needs it (depends), or it is a required member of base. If a user deletes
  the table of a required tool or of a needed dependency by hand, the sync still
  installs the tool, with no overrides. add tool t writes [tools.t] and the
  tables of its new dependencies (as today). remove tool t removes [tools.t] and
  each table of a dependency that no other selected tool needs; it stops with
  code 2 for a required tool, a tool that is not selected, and a tool that
  another selected tool needs. No tool is selected by a pack only."
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
- D12: "Assumed by the planner from D7 of plan 1a: no string literal of the
  engine names a tool; TestEngineNamesNoTool stays green with no new entry in
  its allow list. The frozen copy of the packs files is an embedded data file,
  so its tool names are not Go literals. The pack name base stays the constant
  selection.BasePack."
- D14: "Assumed by the planner from D1-D10 and the rule of the simplest
  consistent CLI rule: (a) list prints the selected tools as a tree with no pack
  roots: each root is a tool with a table that no selected tool needs, required
  tools are marked, and the last line counts the tools. (b) The values file
  keeps format = 2 at the top level next to bootstrap_commit; check accepts it;
  a value other than 2 is code 2. (c) The restore of Apply covers the values
  file only; the files that the sync wrote stay as the sync left them, as ruling
  38 of the TUI design says. (d) add pack on a pack with no member that is not
  selected writes nothing and says "<pack> is already selected". (e) questions
  --json and the --answer rules of plan 1a stay as they are."
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
- D16: "Rulings of the user (2026-11-13): (a) a dependency with selectable =
  false (for example _security_base) gets no table; the sync installs it through
  depends; the check warning of plan 1a stays. (b) The migration gives the env
  of a [packs.<name>] table to each new table of its tools; a tool that already
  has a table keeps its own env; when two old packs give a tool different env
  values, the new table gets the union of both. (c) init adds the tables of the
  members of base only to a new values file and in the migration (a file with no
  format = 2); a later init on a file with format = 2 does not add a removed
  member again. (d) In a git submodule, init, add pack and the migration skip a
  member that is root_only (for example vscode), as the sync does today. (e) add
  pack P --exclude m, when another member of P that the command adds needs m
  (depends), is code 2 with a message that names those members; the user
  excludes them too, or keeps m. (f) The change list has one more item,
  AddPackage (Mac only): a key of [mise.bootstrap.packages] or [mise.tools] of
  an own package that mac detect --add writes; Plan and Apply write it with the
  tools in one write and one sync."
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
- D19: "Assumed by the planner from D1-D18: (a) Each line number of this folder
  is from main before plans 2026-11-12-catalog-engine and
  2026-11-12-catalog-data landed; those plans change cli/ and macos/examples/,
  so a unit finds the code by the name of its function or key, and uses the line
  only as a hint. (b) --no-tui, and init in a terminal before plan 2c, leave out
  the selection screen only: they select no tool other than the tables of D6.
  The questions with no answer follow D14 of plan 2026-11-12-catalog-engine with
  no change: a prompt in a terminal, code 2 and the hint with no terminal or
  --yes; the consent [y/N] stays. (c) The repo sync order: the required members
  of base in the order of packs.toml, then the tools of the [tools.*] tables in
  name order, each dependency before its tool. repo/packs.toml stays excluded
  from dprint for that order. U1 rewrites the comment at dprint.jsonc:6-7 to say
  so; U7 rewrites docs/design/packs.md:62 to say so."
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

## Edits

1. **The frozen copy** — save `repo/packs.toml` and `macos/packs.toml` as they
   are at the branch base of this run (`git show HEAD:repo/packs.toml`, before
   U2 changes them) into data files of `cli/internal/oldform/` (for example
   `packs-v2/repo.toml` and `packs-v2/macos.toml`), embedded
   with `go:embed`. Give the old reader of these files its own small function in
   `oldform`: the `default = true` members, nested packs, and the implicit base.
2. **The new form** — `ReadPacks` and `CheckPacks` read the form of D4: each
   member is `{}`, or `{ required = true }` in base only; `description` is the
   reserved key; any other key, and `default`, are code 2 with a message in the
   form of today. `Expand` takes each member. Remove `IsDefaultTrue` and the
   default state where no caller needs it.
3. **The selection** — `selection.Select` no longer reads `[packs.*]` of a
   values file and no longer adds base implicitly: it adds the required members
   of base, then the `[tools.*]` tables, then the depends (D3).
4. **The format key** — a values file of either side has `format = 2` in the
   new form. `oldform.Repo` and `oldform.Mac` give the code-2 message of an old
   form for a file without it, or with `[packs.*]` (with the hint of today).
   Another value of `format` is code 2.
5. **The migration** — `MigrateRepo` and `MigrateMac` (D5): expand each
   `[packs.<name>]` table and the implicit base with the frozen copy; write a
   `[tools.<name>]` table for each such tool with no table; remove the
   `[packs.*]` tables; remove and list each `[tools.<name>]` table of a tool
   that the clone does not have; write `format = 2`. Keep the comments of the
   file (the yq edit of today). Also write a table for each new dependency
   (D15 b), and skip a frozen member that the clone does not have (D15 c).
   `init` and a new file are U4's (D15 d). Apply D16 (a), (b) and (d): no table
   for a `selectable = false` dependency, the `env` of an old pack table, and
   no `root_only` member in a submodule.
6. **The required tables** — D18 (b, c): on both sides, `sync` writes an empty
   `[tools.<name>]` table for each required member of base that has no table
   (with the skips of D18 b), before it syncs the files, with the
   comment-keeping write of Edit 5, and logs one line for each table.
   `sync --dry-run` and `check` write nothing and list the tables; `check`
   exits as for a file that differs. A dependency gets no table from `sync`.
   The migration of Edit 5 writes tables for the frozen copy only (D18 a).
7. **`list`** — D14 (a): roots are tools, no pack roots.
8. **Go tests** — a unit test for each rule, next to the existing tests;
   golden files again with `UPDATE_GOLDEN=1` where they change. The test data
   of the Go tests that write a `packs.toml` move to the new form.

## Verification

- `cd cli && mise x -- go test ./... -skip TestTOMLMatchesYq` passes,
  `TestEngineNamesNoTool` included (D15 e).
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Do not touch `repo/`, `macos/`, the test tasks or a doc. In `dprint.jsonc`,
  add only the exclude of D17 (a) and the comment of D19 (c).
- Do not change the command grammar, `finish`, or `detect`: U3 and U4 do.
- Add no Go module. Do not edit `cli/VERSION`.

## Commit

`feat: read packs as shortcuts and migrate the pack tables of a values file`
