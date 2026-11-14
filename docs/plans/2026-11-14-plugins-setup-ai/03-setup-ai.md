# U3 — A declarative setup:ai

- **Read first:** `repo/tools/claude/tasks/setup-ai` (as R1a U4 leaves it: the
  submodule exit, the no-`claude` warning, `claude_step`, `plugins_here`, the
  hardcoded vwf block at about lines 64-79, the declared plugins at about
  81-101, the updates at about 103-110, `autoremove` at about 112),
  `repo/tools/claude/tools.toml` (with the `plugins.user` and
  `plugins.project` defaults of R1a D86).
- **Lazy-load:** `repo/tools/git/tasks/merge` (the
  `${BOOTSTRAP_BIN:-mise x -- bootstrap} config get|set` pattern and the D85
  exit-2 rule of R1a U4).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D2:** remove the hardcoded vwf block and the pnpx call. The vwf plugin
  comes from the claude tool's values, which `virajp-plugins` contributes.
- **D3:** on each run, setup:ai uninstalls at project scope each plugin of the
  D5 record that is no longer in `plugins.project`, and drops it from the
  record. It never uninstalls a plugin that is not in the record, and never
  touches user scope.
- **D4:** a plain run installs each listed plugin that is missing at its scope
  (adding its marketplace when no scope has it) and does D3; it runs no
  `marketplace update`, no `plugin update` and no `autoremove`. A second run
  with no list change changes nothing. `mise run setup:ai --update` also runs
  `claude plugin marketplace update` and `claude plugin update` for each listed
  plugin at its scope.
- **D5:** the record is the claude tool's key `plugins.installed`, a list of
  `"<plugin>@<marketplace>"` entries, default empty, declared as a default in
  `repo/tools/claude/tools.toml` (the R1a D86 pattern). setup:ai reads and
  writes it through `${BOOTSTRAP_BIN:-mise x -- bootstrap} config get|set
  tools.claude.plugins.installed`. It adds a project-scope plugin when it
  installs it, and removes one when D3 uninstalls it.
- **D13:** record a plugin only after its install succeeds. If an uninstall
  fails, keep the plugin in the record, so the next run tries again; the
  failed step gives exit 1.
- **D18:** drop from the record a recorded plugin that is not installed at
  project scope, with no uninstall call.
- **D19:** record each listed project plugin that is already installed at
  project scope, on first sight.
- **D24 (replaces D14 and D22):** write the list with
  `config set tools.claude.plugins.installed '<JSON list>'`, with no flag.
- **D25:** ignore the stdout of `set`. A failed record write is a failed step
  (exit 1); the next run corrects the record through D18 and D19.
- **D26:** read the installed list once, at the start. Then do, in order: 1.
  the D18 drop, 2. the install pass with D19 and recording, 3. the D3
  uninstall.
- **D28:** if a `config get` envelope has `catalog: false`, install only:
  skip the D18 drop and the D3 uninstall, and give a warning. Read an absent
  `plugins.installed` as an empty list.
- **D29:** the D3 uninstall runs
  `claude plugin uninstall --scope project --prune -y <plugin>`.
- **D21:** after this unit, `setup-ai` and `repo/tools/claude/` name no vwf.
- **D15:** read `plugins.user` and `plugins.project` with `config get`, which
  returns the merged value with the contributed entries (U1).

## Edits

1. **`repo/tools/claude/tasks/setup-ai`** — D2 to D5, D13, D18, D19, D24 to D26,
   D28, D29;
   `#USAGE` for
   `--update`. Keep the submodule exit, the no-`claude` warning and
   `claude_step`'s collect-and-continue failures, and the exit 1 when any step
   failed.
2. **`repo/tools/claude/tools.toml`** — the `plugins.installed` default (D5).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'pnpx\|vwf\|autoremove' repo/tools/claude/tasks/setup-ai` finds
  nothing.

## Guardrails

- Touch nothing outside `repo/tools/claude/`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: a declarative setup:ai that names no plugin`
