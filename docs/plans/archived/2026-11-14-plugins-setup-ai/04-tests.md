# U4 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (its case helpers, its
  scratch repos and how a case runs the engine), and the work of U1 to U3.

## Ruling

D1 and D7 are quoted in `01-contributed-items.md`; D1, D2 and D6 in
`02-virajp-plugins-tool.md`; D2 to D5 in `03-setup-ai.md`. Every case that runs
a task exports `BOOTSTRAP_BIN=<the engine that test/repo builds>` and
`BOOTSTRAP_CLONE=<the worktree>` (D27). No case uses
the network or a real `claude`.

## Edits

1. **`.config/mise/tasks/test/repo`** — add cases:
   1. A scratch repo with `virajp-plugins`: the synced `.gitignore` has the
      vwf lines once; without the tool, it has none; an override
      `{"$mode"="remove"}` of one vwf line removes it (D1, D7).
   2. `check` refuses a contributed item for a target the owner does not
      declare as merge-mode (D7).
   3. `setup:ai` with a stub `claude` on `PATH` that logs its calls and
      answers `plugin list --json` and `plugin marketplace list --json` from a
      state file:
      1. a first run installs `vwf@virajp-plugins` at user scope and a listed
         project plugin at project scope, and records the project plugin in
         `plugins.installed`;
      2. a second run makes no install, update or uninstall call (D4);
      3. after the project plugin leaves `plugins.project`, a run uninstalls
         it at project scope and drops it from the record (D3); a project
         plugin that is not in the record is never uninstalled;
      4. `--update` makes the update calls (D4).
      5. A stub install that fails: the plugin is not recorded; a stub
         uninstall that fails: the plugin stays recorded; both exit 1 (D13).
      6. a recorded plugin that the stub does not list as installed is
         dropped from the record with no uninstall call (D18);
      7. a listed project plugin that the stub already lists as installed is
         recorded with no install call (D19);
      8. a recorded, listed plugin that the stub does not list as installed
         is installed again and stays recorded (D26);
      9. a `config set` that fails (a stub `BOOTSTRAP_BIN` that exits 5 on
         `set`) gives exit 1 (D25);
      10. with a stub `BOOTSTRAP_BIN` whose `get` answers `catalog: false`,
          a run installs, makes no uninstall call, drops nothing from the
          record, and warns (D28);
      11. the D3 uninstall call carries `--scope project --prune -y` (D29).
   4. With `virajp-plugins` selected, `bootstrap config get
      tools.claude.plugins.user --format json` returns
      `vwf@virajp-plugins` with `source` `"contributed"` (D15); with a
      `plugins.user` entry in the file too, the result has `contributed_by`
      `["virajp-plugins"]` (D20).
2. Update any case that asserts the vwf lines in the git tool's `.gitignore`
   or the pnpx call.

## Verification

- `mise x -- mise run test:repo` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`.
- Delete with `rm`, never `git rm`.

## Commit

`test: contributed merge items, virajp-plugins and a declarative setup:ai`
