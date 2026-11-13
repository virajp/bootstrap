# U7 — Docs

- **Read first:** `docs/design/cli.md`, `docs/design/packs.md`,
  `docs/design/bootstrap-config.md`, `docs/sync/index.md`, `docs/sync/vwf.md`.
- **Lazy-load:** `CLAUDE.md` ("Change a shared file"), `docs/setup/index.md`,
  `docs/setup/tools.md`, `docs/sync/packs/`, `docs/sync/migrate.md`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D1: "Ruling of the user (2026-11-12/13, the artifact "Bootstrap Selection
  Model" (https://claude.ai/artifact/JM1D6D53Lqk3SBbSGrRUGu, source
  docs/scratchpad/selection-model/index.html)): bootstrap <verb> <kind>
  <name>... [flags] acts on the repository of the current folder; bootstrap mac
  <verb> <kind> <name>... acts on the Mac with the same grammar. The kinds are
  tool and pack. add takes tool or pack; remove and replace take tool only
  (remove tool <name>..., replace tool <old> <new>). sync, check, list,
  questions, config and detect take no kind word; questions keeps its tool
  names. Each command keeps the flags of today (--jobs, --verbose, --config,
  --cache-dir, --url, --clone; on the Mac --no-bootstrap and --repo-path)."
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
- D6: "Ruling of the user (2026-11-12/13): init takes no tool or pack names.
  With no terminal, or with the new flag --no-tui, it runs the old-form
  migrations (D5), adds a table for each member of base that has no table (add
  pack base), asks nothing, and syncs; --answer, --dry-run and --yes work on
  this path as today. With a terminal, init opens the TUI of plan 2c; until plan
  2c lands, init with a terminal takes the --no-tui path too. The same for
  bootstrap mac init."
- D7: "Ruling of the user (2026-11-13): detect with no flag only lists the
  detected tools (also in a terminal) and exits 1 when a detected tool is not
  selected, 0 when each detected tool is selected. --add adds each detected tool
  that is not selected, after the consent [y/N]; --yes answers that consent with
  yes. --no is gone (code 2 as an unknown flag). --answer and --dry-run stay.
  The same for bootstrap mac detect."
- D8: "Ruling of the user (2026-11-12): bootstrap add node (a name with no kind
  word), remove node and replace a b stop with code 2 and the hint of the new
  form, for example "use: bootstrap add tool node". The same with mac."
- D13: "Ruling of the user (2026-11-13): the release level is MAJOR
  (bootstrap@3.0.0): the grammar of the CLI and the form of the values file
  break the forms of today. The user tags the release by hand right after the
  landing: until the tag writes cli/VERSION, a consumer with the engine of today
  stops with code 2 on the new packs.toml. After the tag, each consumer and the
  Mac run init once (the migration of D5)."
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

## Edits

1. Run `vwf:docs-sync` over the branch delta. Apply its findings and each
   `DOCS FALSIFIED:` line of U1 to U5.
2. **`docs/design/cli.md`** — the grammar, the flags, `init --no-tui`,
   `detect`, `Plan`/`Apply` and the restore, the exit codes.
3. **`docs/design/packs.md`** — the new form, the shortcut rule, and the sync
   order of D19 (c) at `:62`.
4. **`docs/design/bootstrap-config.md`** — `format = 2`, no `[packs.*]`, a
   table for each tool, the migration, and the tables of the required members
   of base that `sync` writes (D18).
5. **`docs/sync/packs/`** — `base.md` for the new repo base; remove
   `security.md` and `container-build.md` (plain `rm`); one line in "Files that
   bootstrap no longer ships" of `docs/sync/index.md` if the list names docs.
6. **`docs/sync/index.md`, `docs/sync/vwf.md`, `docs/setup/index.md`,
   `docs/setup/tools.md`** — a section for the new Mac tool `pre-commit` (D20
   c);
   the new commands (`bootstrap detect` replaces
   `detect --no`), the packs of the Mac, the removed tool `devtools`.
7. **`docs/sync/migrate.md`** — M38 (D15 f): the new grammar, `format = 2`, run
   `bootstrap init` (or `bootstrap mac init`) once after the release
   `bootstrap@3.0.0`, the removed tool `devtools`, the packs that changed.
8. **`README.md`, `CLAUDE.md`** — where a passage names the old grammar or
   `[packs.*]`.

Write the docs in ASD-STE100, in the style of the existing text.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- The grep of D15 (i) finds nothing.

## Guardrails

- Do not touch `cli/`, the test tasks, `repo/` or `macos/`.
- Do not edit a synced copy (see `CLAUDE.md`).

## Commit

`docs: describe the selection model and the new packs`
