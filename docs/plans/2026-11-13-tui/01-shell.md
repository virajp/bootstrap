# U1 — The shell: the package tui, the layout, the commands and the help

- **Read first:** `screens.txt` of this folder (the layout at the top, the
  rulings list, screens 1 and 9), `docs/design/cli.md` ("The files of `cli/`",
  "The command line", "`init`", "The exit codes"), then
  `cli/internal/cli/cli.go` (`Usage`, `Run`, the `allowed` map, the switch).
- **Lazy-load:** `cli/internal/cli/init.go` (`init`), `cli/internal/cli/git.go`
  (`interactive`), `cli/internal/core/`, the prompt of plan
  2026-11-12-catalog-engine D14 (find it by `charm.land/bubbletea/v2` in
  `cli/`), `cli/internal/source/notools_test.go`.

## Ruling

Quoted verbatim from the decisions of `index.json`:

- D1: "Ruling of the user (2026-11-13): the TUI uses Lip Gloss v2
  (charm.land/lipgloss/v2) for the styles, the colors and the layout of the
  panes, with Bubble Tea v2 and Bubbles v2 of plan 2026-11-12-catalog-engine. U1
  adds it with go get at the latest v2 release at the run; go.sum comes from go
  get and go mod tidy only. The colors are those of screens.txt; the TUI uses
  lipgloss.LightDark with tea.RequestBackgroundColor when screens.txt gives a
  light and a dark value."
- D2: "Ruling of the user (2026-11-13), D23 of plan 2026-11-13-selection-model:
  the TUI in cli/internal/tui calls cli/internal/core (the change list, Plan,
  Apply, Detect) and the catalog reads of cli/internal/source and
  cli/internal/selection. It does not import cli/internal/cli, and core does not
  import it. cli/internal/cli parses the arguments and starts the TUI. The TUI
  builds a change list one key at a time; Review calls core.Plan; Apply calls
  core.Apply. The TUI gives each answer to core directly, so the prompts of the
  CLI never run under the TUI (D14 (2) of plan 2026-11-12-catalog-engine). No
  string literal of cli/internal/tui names a tool: each name, summary,
  description, category and question text comes from the data, and
  TestEngineNamesNoTool stays green with no allow entry for cli/internal/tui."
- D3: "Ruling of the user (2026-11-13): plan 2c adds no automated test of the
  TUI screens; the user tests the TUI by hand after the landing. Backlog item
  B29 (P1) holds the tests: golden View strings at 100x25, no teatest. The tests
  of core come from plan 2026-11-13-selection-model, and test:go, test:repo and
  test:mac stay the gates. A unit can add a Go unit test of a pure function (for
  example a truncation or the pack mark) when it helps; it adds no golden file
  of a screen."
- D4: "Ruling of the user (2026-11-13): the key ? opens a modal help dialog. It
  lists each key of the current screen, one line for each key, with the full
  view of the Bubbles help component. esc closes it. The key line at the bottom
  shows the short view."
- D7: "Assumed by the planner from D1-D6 and screens.txt: (a) the bottom pane
  of a pack also shows the line Members, as screen 2 of screens.txt does. (b)
  The repo and Mac categories and their order come from the category data of
  plan 2026-11-12-catalog-data, not from the script of screens.txt. (c) The TUI
  has no equivalent of --exclude: a pack is a shortcut, and the user deselects
  one tool on the Tools tab. (d) The change to the Mac tool git (ruling 44) is
  in plan 2026-11-12-catalog-data (its D4); plan 2c does not change it. (e) tui
  and mac tui take the flags of init and mac init without --no-tui, --yes,
  --answer and --dry-run; they take no names. (f) tui on a repository with no
  values file stops with code 2 and the message and the hint of the other
  commands today ("<file> is not there", "run bootstrap init first",
  select.go:60); mac tui the same with "bootstrap mac init". (g) With no
  terminal (a.interactive() is false), tui and mac tui stop with code 2 and the
  hint "use: bootstrap init --no-tui" ("bootstrap mac init --no-tui" on the
  Mac); init and mac init with no terminal keep the --no-tui path of plan
  2026-11-13-selection-model. (h) A terminal smaller than 100x25 at the start:
  code 2 and the size message of ruling 56 of screens.txt, for tui, mac tui,
  init and mac init. (i) screens.txt in this folder is a copy of the design
  artifact Bootstrap TUI Screens (version 47); each unit reads its screens
  there, and the rulings of this index.json win over it. (j) Each line number in
  this folder is from main before plans 1a, 1b and 2a; a unit finds the code by
  the name of its function."
- D9: "Ruling of the user (2026-11-13): nine units in nine waves: U1 the shell,
  U2 the Tools tab, U3 the Packs tab, U4 the Detected tab, U5 the questions form
  and Apply, U6 the Mac side, U7 the review, U8 the docs, U9 the gates.
  Execution is teams. Each of U1-U6 changes cli/, so each waits for the unit
  before it."
- D10: "Assumed by the planner from D1-D9 after rulings review round 1: (a) The
  names: plan 1a is 2026-11-12-catalog-engine, plan 1b is
  2026-11-12-catalog-data, plan 2a is 2026-11-13-selection-model, and plan 2c is
  this plan. (b) The order of the checks of tui and mac tui: the terminal check
  of D7 (g) first, then the size check of D7 (h), then the values file of D7
  (f). (c) mac tui finds macos.toml as each mac command except mac init does:
  --repo-path, MACOS_SETUP_REPO_PATH, then the current folder; mac tui also
  takes --repo-path. (d) tui and mac tui on a values file in an old form (no
  format = 2, or a [packs.*] table) stop with code 2 with the same message and
  hint as sync and check on that file (D17 h of plan 2a). (e) The clean-target
  rule of docs/design/cli.md applies to Apply of the TUI: the review names each
  dirty target, Apply writes nothing, and the TUI keeps the change list. U5
  calls the check of that rule in cli/internal/core (D23 of plan 2a, amended
  2026-11-13), the same function that the CLI calls. (f) The search of the Tools
  tab covers the name, the summary, the description and the category, as screen
  3 of screens.txt does. (g) If cli/internal/core does not export the package
  prefixes when U6 runs, U6 exports them from core; the allow entry of
  TestEngineNamesNoTool for those words moves with the code to the new path,
  with the same words (D23 of plan 2a)."
- D11: "Rulings of the user (2026-11-13) on rulings review round 1: (a) init and
  mac init in a terminal with --yes, --dry-run or --answer take the --no-tui
  path, also without --no-tui: these flags come from an agent or a script, and
  the TUI is for a person. (b) On a values file in an old form, init in a
  terminal opens the TUI with the migration as one change on the change list,
  for example "migrate the values file to format 2"; the counter counts it, and
  the review shows it. Apply writes it with the other changes in one write of
  the values file (core.Plan and core.Apply of plan 2a). A quit that discards
  the change list writes nothing, and the next init migrates again. The user
  cannot remove the migration change from the list. (c) On the Mac, the Detected
  tab lists the own packages that core.Detect returns (the installed packages
  that no tool has) in a last category "Own packages", with the [?] marks; space
  accepts one, and a accepts all, also the own packages, the same as bootstrap
  mac detect --add (AddPackage items, D16 f of plan 2a)."
- D12: "Assumed by the planner from D1-D11 after rulings review round 2: (a) D11
  (b) applies to mac init too: on an old-form macos.toml, mac init in a terminal
  opens the TUI with the migration as one change on the change list. (b) The
  next command after a sync that passes: on the repo side, MISE_ENV=dev mise run
  setup (ruling 39 of screens.txt names setup:all, a task that
  docs/sync/migrate.md retired for setup); on the Mac side, none, because the
  Mac sync installs the tools itself and prints its own warnings. The string is
  a constant of cli/internal/tui; it names no tool. (c) The output of the sync:
  core.Apply takes the stdout and the stderr writers from its caller, as
  repo.RunWith does today; core still prints nothing of its own. Before Apply,
  the TUI releases the terminal (the ReleaseTerminal and RestoreTerminal of the
  Bubble Tea program), so the before script, the sync and mise bootstrap on the
  Mac write to the terminal; after Apply, the TUI shows "Press a key to go
  back", then restores the terminal and the main screen. (d) When the user opens
  the Detected tab again, a proposal that is on the change list shows [+] or [-]
  from the change list, not [?]. A Detect result that arrives after esc is
  dropped."
- D13: "Assumed by the planner from D1-D12 after rulings review round 3: (a) If
  core.Apply of plan 2a takes no stdout and stderr writers when U5 runs, U5 adds
  them (it owns cli/) and the CLI passes its own writers; nothing else of
  core.Apply changes. (b) When core.Detect, core.Plan or the clean-target check
  returns an error value, the TUI shows the error in a modal with a message only
  (as ruling 26 of screens.txt); any key closes it, and the TUI keeps the
  change list. (c) The exit code: a quit by the user after the TUI opened is
  code 0, also after a failed sync; the TUI shows the failure on the screen.
  Each error before the TUI opens keeps the exit code that the CLI gives for it
  today (the preflight, the clone, the parse of the values file, a format other
  than 2, the repository of macos.toml), and the checks of D7 f, g, h and D10 b,
  d give code 2."
- D14: "Assumed by the planner from D1-D13 after rulings review round 4: (a) The
  engine has no flag --skip-answer (D15 of plan 1a); D7 (e) and D11 (a) do not
  name it; an empty --answer <tool>.<key>= is an --answer. (b) The order of the
  checks of init and mac init: the terminal check, then the flags of D11 (a) and
  --no-tui (each gives the --no-tui path, also in a small terminal), then the
  size check of D7 (h), then the migration of D11 (b) in the TUI. (c) On a
  repository with no values file, init opens the TUI with one change on the
  change list, "create the values file with base" (ruling 50 of screens.txt):
  the counter counts it, the review shows it, a quit asks "Discard 1 change?
  [y/N]", and the user cannot remove it, as the migration of D11 (b)."

## Edits

1. **Lip Gloss** — `cd cli && mise x -- go get charm.land/lipgloss/v2@latest`,
   then `mise x -- go mod tidy` (D1). Do not write `go.sum` by hand.
2. **The package `cli/internal/tui`** — a Bubble Tea v2 model with: the header,
   the tab line `[Tools] | Packs | Detected`, the left pane (30 columns), the
   right pane (69 columns), the bottom pane (7 lines) and the key line, at the
   sizes of `screens.txt`; the scroll marks ▲ ▼; the fixed widths and the
   truncation; the scroll of the text on the cursor row; the colors of
   `screens.txt` (D1). The tabs are empty in this unit: U2, U3 and U4 fill
   them. The model holds the change list of `core` and the counter of the
   header.
3. **The size** — a terminal smaller than 100x25 at the start: code 2 and the
   size message (D7 h). A shrink during a run shows the message and keeps the
   change list (ruling 17 of `screens.txt`).
4. **The quit** — `q` and ctrl+c: with an empty change list, quit with code 0;
   else the modal "Discard N changes? [y/N]" (rulings 23 and 24).
5. **The help** — `?` opens the help modal; `esc` closes it (D4). The key line
   shows the short view of the Bubbles `help` component.
6. **The commands** — `tui` and `mac tui` in `Usage`, the `allowed` map and the
   switch, with the flags of D7 (e). With no terminal: D7 (g). With no values
   file: D7 (f). `init` and `mac init` open the TUI when `a.interactive()` is
   true and `--no-tui` is not given; a new values file starts with the change of
   D14 (c) (ruling 50); the
   order of the checks of `init` is D14 (b). With no terminal, and with the
   flags of D11 (a), `init`
   keeps its path. On
   an old-form values file, the migration is one change on the change list
   (D11 b).
7. **The entry** — one function, for example `tui.Run(side, opts)`, that
   `cli/internal/cli` calls; it returns an exit code and never calls `os.Exit`.

## Verification

- `cd cli && mise x -- go vet ./... && mise x -- go test ./...` passes.
- `cd cli && mise x -- go build ./cmd/bootstrap` passes.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Touch only `cli/`. Do not touch `repo/`, `macos/`, the test tasks or a doc.
- Do not import `cli/internal/cli` from `cli/internal/tui` (D2).
- No string literal of `cli/internal/tui` names a tool (D2).
- Add no module (S3 is U1 only), and no golden file of a screen (D3).

## Commit

`feat: add the shell of the TUI and the commands tui and mac tui`
