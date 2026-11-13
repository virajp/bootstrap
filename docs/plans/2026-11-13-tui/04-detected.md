# U4 — The Detected tab

- **Read first:** `screens.txt` (screen 2b, and the rulings 11-16, 53 and 61),
  then `cli/internal/tui/`, `cli/internal/core/` (`Detect`).

## Ruling

Quoted verbatim from the decisions of `index.json`:

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
- D5: "Ruling of the user (2026-11-13), which changes ruling 16 of screens.txt:
  the left pane of the Detected tab lists only the categories that have a
  proposal. If nothing is detected, the right pane shows "No tool detected"."
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

## Edits

1. **The dialog** — each time the user opens the tab, the modal "Detecting
   ..." runs `core.Detect` in a Bubble Tea command; `esc` cancels and goes back
   to the tab of before.
2. **The left pane** — only the categories with a proposal (D5); the ◆ hint on
   the Tools tab.
3. **The right pane** — the proposals with `[?]` (green to add, red to
   remove); space accepts one (`[+]` or `[-]`); `a` accepts all. With no
   proposal: "No tool detected" (D5).
4. **The bottom pane** — the evidence file of the proposal.
5. **The own packages** — on the Mac, a last category "Own packages" (D11 c).

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

`feat: add the Detected tab of the TUI`
