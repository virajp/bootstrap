# U8 — Docs

- **Read first:** `docs/design/cli.md`, `README.md`, `CLAUDE.md`.
- **Lazy-load:** `docs/sync/index.md`, `docs/setup/index.md`,
  `docs/setup/tools.md`.

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
- D4: "Ruling of the user (2026-11-13): the key ? opens a modal help dialog. It
  lists each key of the current screen, one line for each key, with the full
  view of the Bubbles help component. esc closes it. The key line at the bottom
  shows the short view."
- D5: "Ruling of the user (2026-11-13), which changes ruling 16 of screens.txt:
  the left pane of the Detected tab lists only the categories that have a
  proposal. If nothing is detected, the right pane shows "No tool detected"."
- D6: "Ruling of the user (2026-11-13): when the user selects a tool with
  questions, the questions form opens. Cancel in that form also cancels the
  select: the tool stays [ ], and no change goes on the change list. Cancel in
  the form of e keeps the old answers, and the tool stays selected. Cancel in a
  form at Apply stops Apply and keeps the change list (ruling 58 of
  screens.txt)."
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
- D8: "Ruling of the user (2026-11-13): the release level of bootstrap is MINOR
  (bootstrap@3.1.0): the commands tui and mac tui are new, and init opens the
  TUI in a terminal; with no terminal and with --no-tui nothing changes, and the
  forms of the values file and packs.toml stay. The run records the level and
  bumps nothing. The user tests the TUI by hand, then tags bootstrap@3.1.0, and
  the release workflow writes cli/VERSION."
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

1. Run `vwf:docs-sync` over the branch delta. Apply its findings and each
   `DOCS FALSIFIED:` line of U1 to U6.
2. **`docs/design/cli.md`** — a row for `internal/tui` in "The files of
   `cli/`"; the commands `tui` and `mac tui` in "The command line"; `init` in
   a terminal; the exit codes of D7 (f), (g) and (h), D10 (d) and D13 (c); the
   `--no-tui` path of D11 (a) and the order of D14 (b).
3. **`README.md`** and **`CLAUDE.md`** — the commands `tui` and `mac tui` in
   the list of the commands.
4. **`docs/sync/index.md`** and **`docs/setup/index.md`** — `init` opens the
   TUI in a terminal; an agent uses `--no-tui` (the TUI is for a person, the
   CLI is for an agent).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Change only the docs of the owns list.

## Commit

`docs: describe the TUI`
