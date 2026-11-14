# U2 — The Tools tab

- **Read first:** `screens.txt` (screens 1, 3, 4, 5 and 6, and the rulings
  1-10, 24-31, 49, 51 and 54), then `cli/internal/tui/` (U1),
  `cli/internal/core/`.
- **Lazy-load:** `cli/internal/selection/`, `cli/internal/source/packs.go`.

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

## Edits

1. **The left pane** — the categories in the order of the data (D7 b), with
   the counter of the selected tools.
2. **The right pane** — the tools of the category: the selector, the name, the
   summary; the dependency tree under each tool, without a selector, repeated
   under each tool that needs it; `selectable = false` only in a tree.
3. **The bottom pane** — the name, the description, the dependencies and the
   status of the tool under the cursor.
4. **The marks** — `[x]`, `[ ]`, `[+]`, `[-]`, `[=]` for a locked row (a
   required tool) and `[·]` for a row that is not available (`hosts`,
   `root_only` in a submodule), grey, with the reason in the bottom pane.
5. **Search** — `/` opens the search of screen 3 over the name, the summary, the
   description and the category (D10 f); `esc` closes it.
6. **Space** — select or unselect; each action is one change on the change
   list, with "Same as:" in the bottom pane (ruling 54).
7. **The alternatives** — the tools with the same `parent` are a group "select
   one or none"; space on another tool of the group opens the replace modal,
   or the long way "Remove A and add B? [y/N]" when there is no migrate script.
8. **The cascade** — an unselect that removes other tools opens the cascade
   modal with `[y/N]`; when a required tool needs the tool, the modal shows
   only a message, and any key closes it.
9. The select of a tool with questions calls a hook that U5 fills; in this
   unit the hook adds the change with no form.

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

`feat: add the Tools tab of the TUI`
