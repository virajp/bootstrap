# U3 — The Packs tab

- **Read first:** `screens.txt` (screen 2, and the rulings 18-20, 45-48,
  59-60), then `cli/internal/tui/`.
- **Lazy-load:** `cli/internal/core/`, `cli/internal/source/packs.go`.

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

## Edits

1. **The left pane** — the packs, with the marks `[x]`, `[~]`, `[ ]`, `[+]`,
   `[-]` and `[=]` for `base` (ruling 47).
2. **The right pane** — read-only: the members, with the summary and a dim
   status ("installed", "will add", "required"). ← → move the cursor into it;
   space does nothing there (ruling 59).
3. **The bottom pane** — on a pack: the name, the description, the status and
   Members (D7 a); on a member: the description, the tools that it needs and
   the status.
4. **Select and deselect** — select is one `AddPack` item; deselect removes
   each member, except a required member and a member that another selected
   tool needs; a tool outside the pack that needs a member opens the cascade
   modal of U2. A pack is one change in the counter (ruling 60).

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

`feat: add the Packs tab of the TUI`
