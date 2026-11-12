# Plan — The catalog engine: the new keys of a tool, a pack and a question (2026-11-12)

## Goal

After this plan lands, the engine accepts and validates the new catalog keys
and enforces their CLI rules: `summary`, `category`, `selectable`, `hosts` and
`description` of a tool, the categories file of each side, the `description` of
a pack, and the question keys `group`, `note` and `optional`. `init` takes its
answers
from flags first (`--answer`, the new `--skip-answer`), and asks only the
questions with no answer, in a field that starts with the detected answer
(D14). The data of today still syncs with no change.

This is plan 1a of three chained plans for the TUI `bootstrap tui` and
`bootstrap mac tui`. The user finalized the screens in the artifact "Bootstrap
TUI Screens" (https://claude.ai/artifact/QGqhhDrKesC3N2Kz7x1nr5) on 2026-11-12.
Plan 1b (the catalog data) requires this plan and lands after the release
`bootstrap@2.1.0` (D8). Plan 2 (the TUI) requires plan 1b. The plan reverses no
standing decision. It changes the prompt of `init` of today: an empty line no
longer takes the proposal, because the field starts with it (D14).

## Facts the survey established

- Dispatch: `Usage` is at `cli/internal/cli/cli.go:31`, `Run` at `:150`, the
  `allowed` table of flags at `:181-193`, `parse` at `:268`. The exit codes are
  in `cli/internal/fail/fail.go:19-27` (`fail.Usage` is code 2, `fail.Missing`
  is code 3). `a.interactive()` is at `cli/internal/cli/git.go:121`.
- The manifest: `Source.LoadManifest` is at
  `cli/internal/source/manifest.go:117`.
  The struct `Manifest` (`manifest.go:49-71`) already has `Description`, which
  `LoadManifest` reads on both sides (`manifest.go:124-126`). The repo side
  reads the other keys through a flat line form (`m.repoLines`).
- The checks: `Source.Open()` is at `cli/internal/source/check.go:41`.
  `repoRules` (`check.go:118`) has no rule for an unknown key.
  `macosRules` (`check.go:208-309`) refuses an unknown key in its `default`
  branch (`check.go:304`), with a message that lists the keys. It validates the
  keys of a question at `check.go:253-278`: only `question`, `detect` and
  `example` are allowed now. The repo side validates the keys of a question
  in `cli/internal/values/merge.go:224`, with the same message, and
  `.config/mise/tasks/test/repo:2359` and `:2411` assert it.
- The packs: `Source.ReadPacks` is at `cli/internal/source/packs.go:101`. A
  value that is not a table goes to `p.Bad` at `packs.go:124-127`, and
  `selection.CheckPacks` (`cli/internal/selection/selection.go:117-123`) then
  stops with code 2. `scanOrder` and `reMemberRow` (`packs.go:~160-170`) read
  the order of the members from the text of the file.
- The questions: `cli/internal/questions/questions.go`: `Of` (`:30`), `Answer`
  (`:61`, a key that exists is an answer), `Missing` (`:74`), `Proposal`
  (`:88`). `sync` stops with code 2 for a missing answer in
  `cli/internal/repo/context.go:35-40` and `cli/internal/macos/macos.go:447`,
  with the message "the question <tool>.<key> has no answer in <file> - run
  bootstrap init". `app.ask` (`cli/internal/cli/init.go:428-491`) asks in a
  terminal and loops until the line is not empty (`init.go:470-488`), with
  `a.readLine` (`cli/internal/cli/git.go:128`, a plain `bufio` line). The
  `allowed` table gives `answer` to `init`, `add`, `replace` and `detect`.
  `go.mod` has no TUI library now. `app.questions`
  (`init.go:523-631`) writes the JSON rows at `init.go:563-583` (keys `tool`,
  `key`, `question`, `proposal`, `source`, `answer`).
- The selection: `app.open` is at `cli/internal/cli/select.go:48`, `known` at
  `:158`, `planAdd` at `:183`, `planRemove` at `:210`, `add` at `:344`,
  `addNames` at `:357`, `replace` at `:393`, `detect` at `:529`. Since
  `601225f`, the Mac `detect` also proposes tools from the installed
  packages: `installedPackages` and `scanMac`
  (`cli/internal/cli/packages.go:153`, `:212`). `init` is at `init.go:132`, and
  `finish` at `init.go:231`. `list` is at `cli/internal/cli/list.go:24`.
  `sync` is at `cli.go:529` and `check` at `cli.go:580`.
- No code of the engine examines the host now (no `runtime.GOOS`).
- `TestEngineNamesNoTool` (`cli/internal/source/notools_test.go:103`, allow
  lists
  at `:25-62`) fails when a string literal of the engine names a tool folder.
- Goldens: `cli/internal/selection/testdata` and others. Write them again with
  `UPDATE_GOLDEN=1 go test ./...` in `cli/`.
  `cli/internal/source/testdata/toml-yq.golden`
  holds a hash for each `tools.toml`.
- The tools with the new keys in plan 1b: `_security_base` is the only repo tool
  that gets `selectable = false` (`gitleaks`, `grype` and `osv-scanner` depend
  on it). `swift` is the only tool that gets `hosts = ["macos"]`. The Mac tool
  `git` gets `group = "email"` on `email_github` and `email_gitlab`
  (`macos/tools/git/tools.toml:18-25`).
- The repo already uses the mise OS names: `os = "macos"` in
  `repo/tools/swift/tools.toml:23` and `repo/tools/mise/tools.toml:43`.
- The test tasks: `.config/mise/tasks/test/repo` has the packs cases at
  `:1780-1809` (for example `packs_member_string`, which expects
  "repo/packs.toml: [pack.odd] node must be a table, { default = true } or {}")
  and `questions --json` at `:3001-3002`. `.config/mise/tasks/test/mac` has the
  packs cases near `:330-338`, and its fixtures are in `macos/fixtures/`.
- The release rule: `docs/design/cli.md` ("The release of the engine", "The
  rule"). `cli/VERSION` is `2.0.3`. Nobody edits it by hand.
- The docs that describe the keys: `docs/design/tools.md` (the key table near
  `:39-57`, the questions near `:293-297`), `docs/design/packs.md` (near
  `:56-57`), `docs/design/cli.md` (the commands, near `:75-110`),
  `docs/design/mac.md` (near `:103-118`), `docs/sync/index.md` (near `:325`
  and `:418`), `docs/setup/tools.md` (near `:6-40`), `README.md` (near
  `:24-34`) and `CLAUDE.md` (near `:34-36`).
- The repo has no commit scopes. The commit types are in
  `.config/git-conventional-commits.yaml`: `feat`, `fix`, `perf`, `refactor`,
  `revert`, `test`, `ops`, `docs`, `merge` and `wip`.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file with Write, Edit, `sed` or a
script.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence>
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

A block with no question carries the one line `QUESTION: none`.

## Out of scope

- The data: the texts (`summary`, `description`), the categories files, the
  `category` of each tool, the pack descriptions, `selectable = false` on
  `_security_base`, `hosts` on `swift`, and the `group` and `note` of the Mac
  git questions. They need the released engine (D8), so they are in plan 1b.
- `summary` and `category` as necessary keys: plan 1b, after the data exists.
- The git templates of the Mac side (rule 7 of D6): plan 1b.
- The TUI, the extraction of `finish` and `detect` into functions, the "select
  one or none" rule for `parent`, and the kind of a Mac tool: plan 2.
- A `conflicts` key: the user ruled that `parent` is the only conflict data.

## Parked

- Plan 1b, the catalog data: `repo/categories.toml` and
  `macos/categories.toml` with the accepted lists and mapping, the 128
  summaries, the 36 repo descriptions and the pack descriptions (written in the
  plan folder for the user to read), `selectable = false` on `_security_base`,
  `hosts = ["macos"]` on `swift`, the Mac git questions and templates, and
  `summary` and `category` made necessary.
- Plan 2, the TUI: `bootstrap tui` and `bootstrap mac tui` per the artifact.
  It uses the field model of D14 in its question form, and gives every answer
  to the CLI with `--answer` or `--skip-answer`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-12-catalog-engine

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
