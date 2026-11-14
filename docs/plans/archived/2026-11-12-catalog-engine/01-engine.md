# U1 — The engine: the new keys and their rules

- **Read first:** `docs/design/cli.md` ("The command line", "The exit codes"),
  `docs/design/tools.md` (the key table and the questions), then
  `cli/internal/source/manifest.go`, `cli/internal/source/check.go`,
  `cli/internal/source/packs.go`, `cli/internal/questions/questions.go`,
  `cli/internal/cli/select.go` and `cli/internal/cli/init.go`, top to bottom.
- **Lazy-load:** `cli/internal/cli/list.go`, `cli/internal/cli/cli.go`,
  `cli/internal/repo/context.go`, `cli/internal/macos/macos.go`,
  `cli/internal/selection/`, `cli/internal/source/notools_test.go`,
  `cli/internal/cli/git.go` (`interactive`, `readLine`, `consent`).

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D1: "summary (a string of 47 characters or fewer: the right pane of the TUI),
  category (a string: the id of a category of the side), selectable (a boolean,
  default true) and description (a string, now also read on the repo side: the
  bottom pane of the TUI). In this plan, summary and category are optional. A
  wrong type, or a summary of more than 47 characters, stops the sync with code
  2."
- D2: "One categories file for each side: repo/categories.toml and
  macos/categories.toml. Each entry is [category.<id>] with the key title (a
  string). The order of the entries in the file is the order in the TUI. A tool
  names one id: category = "<id>". The sync stops with code 2 if a tool names an
  id that is not in the file of its side, or if a tool names a category and the
  file is missing. A missing file with no category in a tool is not an error.
  Another key in a [category.<id>] table, a [category.<id>] with no title, a
  title that is not a string, or a top-level key or table other than [category]
  stops the sync with code 2."
- D3: "description is a reserved key in [pack.<name>] of repo/packs.toml and
  macos/packs.toml. A string value at the key description is the text of the
  pack. An inline table is a member. A string at another key still stops the
  sync with code 2 with the message of today. A description that is not a string
  stops the sync with code 2. A pack has no summary."
- D4: "hosts is an optional list of the mise OS names: macos, linux and windows.
  The engine maps the Go name darwin to macos (runtime.GOOS). No key means each
  host. It is a key of the repo side only: a Mac tool with hosts stops the sync
  with code 2. Only the TUI (plan 2), init, add and the new tool of replace
  examine it: on another host they refuse the tool with code 2 and a message
  that names the tool and its hosts. hosts = [] stops the sync with code 2
  ("hosts must name at least one host"). On a runtime.GOOS other than darwin,
  linux and windows, the host name is runtime.GOOS as it is, so init and add
  refuse each tool that has hosts. sync and check do not examine it, so a CI job
  on Linux still syncs a repository that has a macOS-only tool. A value that is
  not one of the three names stops the sync with code 2."
- D5: "init and add refuse a named tool with selectable = false, with code 2,
  and replace refuses it as its new tool only (the old tool can be one that a
  values file names directly). The message names the tools that need it, for
  example: "_security_base is not selectable. Select gitleaks, grype or
  osv-scanner." (the tools of the side whose depends name it, sorted). If no
  tool of the side depends on it, the message is only "<tool> is not
  selectable." bootstrap questions --json and bootstrap list do not change: they
  show selected tools only, and a values file that names it directly still shows
  it. sync accepts a values file that names it directly. check prints a warning
  only and keeps its exit code. The key works the same way on both sides."
- D6: "Rule 1: group is a string (the name of the group, for example group =
  "email"). note is a string. optional = true is a boolean, and a question with
  a group cannot also have it (code 2). Each key is valid on both sides. Rule 2:
  an empty string "" is a valid answer, and it means no answer. init writes ""
  for a question that the user leaves empty, so init does not ask it again on
  the next run. Rule 3: sync and check: an optional question can have no key or
  "". A group needs at least one answer that is not empty. If all the answers of
  a group are empty or missing, the sync stops with code 2: "the group
  <tool>.<group> has no answer in <file> - <hint>", with the file and the hint
  of the message of today for a single question ("run bootstrap init" on the
  repo side, the hint of oldform.MacHint on the Mac side), for example "the
  group git.email has no answer in macos.toml - run bootstrap mac init". A
  question with no group and no optional still needs a key, as today; "" is
  accepted as its answer. Rule 4 (changed by D14): init, add, replace and detect
  ask in a terminal each question of the selected tools that has no key in the
  files and no --answer or --skip-answer flag. They also ask an optional
  question. The field starts with the proposal (the output of detect, else the
  example), and the user can edit or delete the text. Enter accepts the text. An
  empty text is accepted for an optional question or a group question, and init
  writes "". For another question, the field stays open and shows "an answer is
  necessary". init also asks again each question of a group whose answers in the
  files are all empty or missing, because the sync stops for that group. After
  the last question of a group, if all its answers are empty, init shows "Give
  at least one answer: email_github, email_gitlab" and asks again every question
  of the group, also the questions that a flag set, each field started with its
  proposal. init shows the note on a line under the question. Rule 5 (changed by
  D14): --skip-answer <tool>.<key> (repeatable) and --answer <tool>.<key>= both
  write "" for an optional or a group question, so init does not ask it now or
  on a later run, except in the second ask of an all-empty group (rule 4). Both
  forms are accepted only for an optional or a group question; for another
  question they give code 2. With --yes or no terminal, a question with no key
  in the files and no flag stops init with code 2, as today. The hint names
  --answer, and also --skip-answer for an optional or a group question. A group
  needs at least one answer that is not empty, from the files or from --answer.
  Rule 6: bootstrap questions --json gives each question the keys group, note
  and optional too. The text output of bootstrap questions shows an answer "" as
  answered, with the value (empty); it shows an optional question with no
  answer, and a group question with no answer in a group that has an answer,
  with the info glyph and "no answer (optional)" or "no answer (group <group>
  has an answer)", and it does not count them as "without an answer". Rule 7
  (the git templates of the Mac side) is in plan 1b."
- D7: "The engine reads the new keys generically. No string literal of the
  engine names a tool, also not _security_base, git or swift.
  TestEngineNamesNoTool stays green with no new entry in its allow list. The
  texts of D5 and D6 are examples; the engine builds them from the data."
- D13: "Assumed by the planner from D4-D6 and the rule of the simplest
  consistent CLI rule: (a) bootstrap detect does not propose a tool whose hosts
  leave out the host or that has selectable = false, so detect --yes never stops
  for them. This includes the Mac detect, which proposes tools from the
  installed packages (scanMac, cli/internal/cli/packages.go:212, since 601225f).
  (b) D6 rule 5: with --yes or no terminal, a group passes when one of its
  answers is not empty, from the files or from --answer, so init --yes again on
  a values file that init wrote for the group (each question of the group has a
  key, one of them not empty) passes with no flag. (c) In bootstrap questions, a
  question of a group whose answers are all empty or missing shows the error
  glyph and "no answer", and counts as without an answer. (d) optional = false
  is valid, and it is accepted with group; only optional = true with group is
  code 2. (e) The warning of check (both sides, also bootstrap mac check):
  "<tool> is not selectable - remove [tools.<tool>] from <file>; a tool that
  needs it selects it". (f) The allowed table of cli.go is at :181-193."
- D12: "Assumed by the planner from the rulings D2-D6 and the standing rule of
  the user to take the simplest consistent rule for a CLI case: the repo
  question keys are validated in cli/internal/values/merge.go:224 (change that
  check, no second check); replace examines hosts and selectable on its new tool
  only; questions --json and list do not change for selectable = false; the
  exact group message with the file and the hint of today; init asks again an
  all-empty group; --skip-answer writes "" for an optional or group question
  (D14); the text output of bootstrap questions per D6 rule 6; a category with
  no title and an unknown top-level key are code 2; hosts = [] is code 2 and an
  unknown GOOS is used as it is; a not-selectable tool that no tool needs gets
  the short message."
- D14: "Rulings of the user (2026-11-12): (1) init and add ask in a terminal
  only the questions that have no answer in the files and no flag. If --answer
  gives 3 of 4 answers, init asks the fourth. (2) The TUI of plan 2 gives every
  answer with a flag, also the empty ones, so the CLI never asks under the TUI.
  (3) Without a flag, init also asks an optional question. --skip-answer
  <tool>.<key> or --answer <tool>.<key>= leaves it empty. (4) The y/N consent
  before init writes stays, and --yes skips it. (5) The field of the prompt
  starts with the detected answer, and the user can edit or delete it. (6) The
  prompt uses Bubble Tea v2 (charm.land/bubbletea/v2) and the textinput
  component of Bubbles v2 (charm.land/bubbles/v2), the latest releases at the
  run (v2.1.1 and v2.2.1 on 2026-11-12). U1 adds them with go get; go.sum comes
  from go get and go mod tidy only. Plan 2 uses the same field model in the TUI
  form. Assumed by the planner from the simplest consistent CLI rule:
  --skip-answer is accepted wherever --answer is (init, add, replace, detect),
  and also for a group question; the prompt runs only when a.interactive() is
  true, with the stdin and the stdout of the app; the y/N consent keeps
  readLine. (7) Ruling of the user after rulings review round 3: --skip-answer
  <q> writes "", the same as --answer <q>=, so the key is in the file and a
  later init --yes does not stop for it. (8) Ruling of the user after rulings
  review round 3: if each answer of a group is empty after the flags and the
  prompt, init in a terminal asks again every question of the group, also the
  questions that a flag set, each field started with its proposal; with --yes or
  no terminal, init stops with code 2 and names the group and its keys. (9)
  Rulings of the user after rulings review round 4: (a) if the group is empty
  again at the second ask, init asks the group again until one answer is not
  empty; ctrl+c stops init with code 2 and writes nothing. (b) --skip-answer <q>
  and --answer <q>=<value> on the same question stop init with code 2: "--answer
  and --skip-answer both name <q>"; the same flag two times on one question: the
  last one wins, as --answer does today (init.go:437-444). (c) The code-2
  message of an all-empty group with --yes or no terminal: "the group
  <tool>.<group> has no answer", with the hint "give at least one with --answer
  <tool>.<key>=<value>: <key>, <key>" (the keys of the group in key order), for
  example "the group git.email has no answer" and "give at least one with
  --answer <tool>.<key>=<value>: email_github, email_gitlab"."
- D15: "Ruling of the user (2026-11-13), which replaces each --skip-answer rule
  of D6, D12 and D14: the engine has no --skip-answer flag. --answer
  <tool>.<key>= (an empty value) is the one form of an empty answer: it writes
  "", and it is accepted only for an optional or a group question (code 2 for
  another question). Where D6, D12 or D14 names --skip-answer, read --answer
  <tool>.<key>=. The rule of D14 (9 b) for two flags on one question falls away:
  for --answer two times on one question, the last one wins, as today. The hint
  of a missing answer with --yes or no terminal names --answer, and for an
  optional or a group question also --answer <tool>.<key>= . The TUI of plan 2
  gives an empty answer with --answer <tool>.<key>= when it calls the CLI, and
  through the change list in the process."

## Edits

1. **`cli/internal/source/manifest.go`** — add the fields `Summary`,
   `Category`, `Selectable` (true when the key is missing) and `Hosts` to
   `Manifest`, and read them in `LoadManifest` on both sides. Keep
   `Description` as it is.
2. **`cli/internal/source/check.go`** — validate the new keys:
   - On both sides: `summary` is a string of 47 characters or fewer (count
     runes, not bytes); `category` is a string; `selectable` is a boolean;
     `description` is a string. Each wrong value is one problem with code 2,
     in the form of the existing messages (`<tools.toml shown>: <key> must be
     ...`).
   - `macosRules`: accept `summary`, `category` and `selectable` in its
     switch, and add them to the key list of the `default` message. `hosts` on
     the Mac side gives the problem "hosts is for repo tools only", as
     `default_envs` does today.
   - The repo side: validate `hosts` as a list of `macos`, `linux` and
     `windows`.
   - The keys of a question, on both sides: allow `group` (string), `note`
     (string) and `optional` (boolean) next to `question`, `detect` and
     `example`. A question with both `group` and `optional` is a problem.
     Change the list in the message of `check.go:265` (Mac side) and of
     `cli/internal/values/merge.go:224` (repo side, the existing check) to
     name the six keys. Do not add a second check of the question keys.
   - `hosts = []` is a problem: "hosts must name at least one host".
3. **The categories file** — in `cli/internal/source/`, read
   `<side root>/categories.toml` (`repo/categories.toml`,
   `macos/categories.toml`) with the existing TOML reader. Keep the ids in the
   order of the file (the TUI of plan 2 reads that order; use the text order,
   as `scanOrder` does for the packs). Give the title of each id. Problems,
   code 2: a file that does not parse, a top-level key or table other than
   `[category]`, a `[category.<id>]` with no `title` or with a key other
   than `title`, a title that is not a string, a tool whose `category` is not
   an id of the file, a tool with a `category` when the file is missing. Run
   this check where the other checks of the index run, so `sync`, `check` and
   each command that opens the clone see it.
4. **`cli/internal/source/packs.go`** — in `ReadPacks`, the key
   `description` of a `[pack.<name>]` table with a string value is the text of
   the pack: store it (for example `Packs.Descriptions[name]`), and do not add
   it to `Members` or `Bad`. A non-string `description` is a problem with
   code 2.
   Make sure that `scanOrder` does not read the `description` line as a member
   (a member row has `= {`, so it does not match; examine it).
5. **`cli/internal/questions/questions.go`** — add `Group`, `Note` and
   `Optional` to `Question`, and read them in `Of`. Change `Missing` per D6
   rule 3: skip an `optional` question; for a group, give one entry for the
   group only when no question of the group has an answer that is not empty.
   Give the caller what it needs to print the group message. Add a helper
   that tells whether an answer is empty (`""`).
6. **`cli/internal/repo/context.go:35` and `cli/internal/macos/macos.go:447`**
   — print the group message of D6 rule 3 for a group, with the file and the
   hint of the message of today, and the message of today for a single
   question.
7. **`cli/internal/cli/init.go`** — `ask`: rules 2, 4 and 5 of D6, and D14.
   - `--answer <q>=` with an empty value, on a question that is neither
     `optional` nor in a group, is an error with code 2 (D15).
   - A question with an `--answer` takes that value; `--answer <q>=` takes
     `""`.
   - With `--yes` or no terminal, a question with no key in the files and no
     flag is the error of today (code 2). Its hint names `--answer`, and also
     `--answer <tool>.<key>=` for an optional or a group question (D15). A group
     with no
     non-empty answer (from the files or `--answer`) is an error with code 2
     that names the group and its keys.
   - In a terminal, replace the `readLine` loop of the questions
     (`init.go:470-488`) with the field model of item 12. Ask each question
     with no key and no flag, an `optional` one too. The field starts with the
     proposal. Enter accepts the text; an empty text gives `""` for an
     `optional` or a group question, and keeps the field open with "an answer
     is necessary" for another question. Show the `note` on a line under the
     question. After the last question of a group, if each answer of the group
     is empty, print "Give at least one answer: <key>, <key>" (the keys of the
     group, in key order) and ask again every question of the group, also the
     questions that a flag set, each field started with its proposal (D14 (8)).
     Also ask
     each question of a group whose answers in the files are all empty or
     missing (today `ask` skips each question whose key exists,
     `init.go:452`). An interrupt (ctrl+c) stops `init` with code 2 and
     writes nothing.
   - D14 (9): ask an all-empty group again until one answer is not empty.
     For `--answer` two times on one question, the last one wins (D15). The
     code-2 message of an all-empty group with `--yes`
     or no terminal is "the group <tool>.<group> has no answer", with the hint
     "give at least one with --answer <tool>.<key>=<value>: <key>, <key>".
   - `consent` (`git.go:142`) keeps `readLine`.
   - `questions`: add `group`, `note` and `optional` to the JSON rows (`group`
     and `note` as strings, empty when missing; `optional` as a boolean).
     Change the text output (`init.go:600-625`) per D6 rule 6.
8. **`cli/internal/cli/select.go` and `init.go`** — D4 and D5 in `init`,
   `add` and the new tool of `replace` (not its old tool): before a change is
   planned, refuse each named tool with `selectable = false` (code 2, the
   message of D5, built from the `depends` of the other tools of the side;
   the short message when no tool depends on it) and each named tool whose
   `hosts` does not
   contain the current host (code 2, for example "swift needs the host macos -
   this host is linux"). The check is on the names that the user gives, not on
   the dependencies that `planAdd` adds. Put the host name in one function that
   maps `runtime.GOOS` (`darwin` → `macos`, another value as it is), so the
   tests can replace it.
9. **`detect`** (`select.go:529`, and the Mac scan `scanMac` at
   `packages.go:212`) — do not propose a tool whose `hosts`
   leave out the host or that has `selectable = false` (D13 a).
10. **`check`** (`cli.go:580`) — for each tool with `selectable = false` that
    the values file names directly, print a warning (the existing warning print
    function), and keep the exit code. `list` and `questions --json` do not
    change for `selectable = false` (D5).
11. **`cli/internal/cli/cli.go`** — no new flag (D15). Change the line of
    `--answer` in `Usage` (`cli.go:31`) to say that `--answer <tool>.<key>=`
    writes an empty answer for an optional or a group question.
12. **The field model** — a new file in `cli/internal/cli/` (or a small new
    package in `cli/internal/`) with a Bubble Tea v2 model around the
    `textinput` of Bubbles v2: the question, the note under it, the field
    started with `SetValue(proposal)`, and the line "an answer is necessary".
    It gives the text, or an interrupt. Plan 2 uses it in the TUI form, so keep
    it free of the questions of `init`. Add the modules with
    `go get charm.land/bubbletea/v2@latest charm.land/bubbles/v2@latest` in
    `cli/`, then `go mod tidy`. Add `charm.land/lipgloss/v2` only if the model
    sets a style.
13. **Go tests** — a unit test for each rule above, next to the existing tests
    of each package. Write the golden files again with `UPDATE_GOLDEN=1` where
    a golden changes, and examine each changed golden.

## Verification

- `mise x -- mise run test:go` passes, `TestEngineNamesNoTool` included. A
  unit test drives the field model with key messages: the start text is the
  proposal, a delete and a new text change the answer, and an empty text on a
  necessary question keeps the field open.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn '_security_base\|"swift"\|"git"' cli/internal --include='*.go' |
  grep -v _test`
  finds no new literal.
- Build the binary and run `bootstrap check` in the main checkout: exit 0 with
  the data of today.

## Guardrails

- Do not touch the test tasks, `repo/`, `macos/` or a doc.
- Do not edit `cli/VERSION`.
- Add no Go module other than the modules of D14 (item 12). Do not edit
  `go.mod` or `go.sum` by hand: use `go get` and `go mod tidy`.
- Use Context7 (`resolve-library-id`, then `query-docs`) for the API of Bubble
  Tea v2 and Bubbles v2 before the code. The import paths start with
  `charm.land/`, not `github.com/charmbracelet/`.
- Do not change a message that an existing case of `test:repo` or `test:mac`
  asserts, except the key list of the Mac unknown-key message and the key list
  of the question message (U2 changes those cases).

## Commit

`feat: add the catalog keys of a tool, a pack and a question to the engine`
