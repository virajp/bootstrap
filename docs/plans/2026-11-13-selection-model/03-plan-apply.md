# U3 — Engine: Plan and Apply

- **Read first:** `docs/design/cli.md` ("The command line", "The exit codes",
  "The output of a run"), then `cli/internal/cli/init.go` and
  `cli/internal/cli/select.go`, top to bottom.
- **Lazy-load:** `cli/internal/cli/git.go`, `cli/internal/cli/cli.go`,
  `cli/internal/oldform/oldform.go`, `cli/internal/repo/repo.go`,
  `cli/internal/macos/macos.go`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D9: "Ruling of the user (2026-11-12, option 1): the core of finish
  (cli/internal/cli/init.go:231-368) moves into two functions that return errors
  as values and print nothing: Plan(side, list) gives the new values bytes, the
  cascade, the questions without an answer and the dry-run file list;
  Apply(side, plan) runs the before script, writes the values file and the
  migration changes, and runs the sync; if the sync fails, Apply puts the values
  file back as it was before Apply (ch.cur and ch.existed hold the old bytes)
  and returns the code of the sync. init, add, remove, replace and detect --add
  build a change list (items AddTool, AddPack with Exclude, RemoveTool, Replace,
  and answers) and call Plan, then Apply, so each command writes one time and
  syncs one time. The consent, the dirty-tree check and the print of the changes
  stay in the CLI. Plan 2c calls the same functions."
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

1. **The change list** — a type for the list (for example in a new package
   `cli/internal/change` or in `cli/internal/cli`): items `AddTool`,
   `AddPack` (with `Exclude`), `RemoveTool`, `Replace`, `AddPackage` (Mac,
   D16 f), and the answers
   (`<tool>.<key>` → value, `""` for an empty answer).
2. **`Plan`** — from `finish` (`init.go:231-310`) and from `planAdd`,
   `planRemove` and the replace checks: it expands `AddPack` (D2), applies the
   items in order to a temporary copy of the values file (D16 a, d, e:
   no table for a `selectable = false` dependency, no `root_only` member in a
   submodule, code 2 for an excluded member that another member needs), finds
   the questions
   with no answer, and runs the dry-run plan. It returns the new values bytes,
   the change rows, the questions with no answer, the dry-run file list, and
   each problem as an error value (no print, no `a.fail`).
3. **`Apply`** — from `finish` (`init.go:350-365`): the before script of a
   replace, the migration changes, the write of the values file, the sync. If
   the sync fails, it writes the bytes of before Apply back (or removes a file
   that did not exist) and returns the code of the sync (D14 c).
4. **`finish`** — becomes the CLI shell: it calls `Plan`, prints the changes,
   asks the questions (`ask`), takes the consent and the dirty-tree check,
   then calls `Apply`. The output of each command of today stays the same.
5. **Go tests** — unit tests of `Plan` and `Apply`, with a sync that fails and
   restores the file.

## Verification

- `cd cli && mise x -- go test ./... -skip TestTOMLMatchesYq` passes (D15 e).
- Do not run `test:repo` or `test:mac` as a check of this unit: U5 changes
  them after U4.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Do not change the grammar or the flags: U4 does.
- Keep the process-global state serial (facts section).
- Do not touch `repo/`, `macos/`, the test tasks or a doc.

## Commit

`refactor: move the plan and the apply of a change into functions`
