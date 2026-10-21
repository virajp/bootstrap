# U5 — Docs: the tools, the packs and the new values file

- **Wave:** 5
- **Depends on:** U4
- **Owns:** `docs/sync/**`, `README.md`, `CLAUDE.md`, `docs/design/*.md`
- **Model:** opus
- **Kind:** edit
- **Read first:** `docs/sync/index.md`, `docs/sync/base.md`,
  `docs/design/packs.md`, `docs/design/bootstrap-config.md`, `CLAUDE.md`.
- **Lazy-load:** each other doc that an edit below names.

## Ruling

Quoted from `index.md`:

- D18: "'one file for each tool, as the rule says. Also we need to have a file
  for pack.' Tool docs in `docs/sync/tools/`, pack docs in `docs/sync/packs/`.
  `index.md`, `migrate.md` and `vwf.md` stay in `docs/sync/`.
  `claude-tool.md` keeps its name."
- D19: "Record in `docs/design/fnox.md`: a run task calls a helper of
  `_scripts/helpers`, for example `run_with_secrets <command>`. With fnox
  (`mise x -- which fnox`) it runs
  `mise x -- fnox exec --profile "${MISE_ENV%%,*}" -- <command>`; with
  doppler, `doppler run -- <command>`; else the command directly. Both `fnox`
  and `doppler` selected stops the sync with code 2. Not built now."
- D1 to D17, as `index.md` states them: the docs describe what U1, U2 and U3
  built.

## Edits

1. **Run `vwf:docs-sync`** over the branch delta of the run, and apply its
   findings plus every `DOCS FALSIFIED:` line of U1, U2 and U3, and the list of
   docs in "Facts the survey established" of `index.md`.
2. **Move the tool docs.** Move each `docs/sync/<tool>.md` (with
   `claude-tool.md`) to `docs/sync/tools/`. Delete `docs/sync/base.md`. Keep
   `index.md`, `migrate.md` and `vwf.md` in `docs/sync/`. Correct every link
   to a moved file in the whole repository, `README.md` and `CLAUDE.md`
   included.
3. **New tool docs** in `docs/sync/tools/`: `mise.md`, `repo-sync.md`,
   `git.md`, `pre-commit.md`, `dprint.md`, `linter.md`, `_security_base.md`,
   `gitleaks.md`, `grype.md`, `osv-scanner.md`. Each has the same sections as
   the other tool docs. "When the tool is relevant" says that the `base` pack
   always adds it.
4. **New pack docs** in `docs/sync/packs/`: `base.md` and `security.md`. Each
   says what the pack adds, its `required` tools, and the questions that the
   adopting AI asks (none for `base`, which is always added).
5. **`docs/sync/index.md`:** step 1 selects tools and packs; step 2 copies the
   sync files from `tools/repo-sync/` and `tools/mise/`; step 3 describes
   `[tools.<folder>]`, `[packs.<name>]`, `env` and `default_envs`, and the
   example uses tables; the exit code table matches U2. `docs/sync/vwf.md`:
   `add tools` and `remove tools` write `[tools.*]` and `[packs.*]` entries,
   and a tool of the `base` pack cannot be removed.
6. **`docs/sync/tools/claude-tool.md`:** not root only, `code:graph` and the
   `graphify-refresh` hook, and `setup:ai` skips in a submodule.
   `node.md` and `python.md`: `linter` and `pre-commit` add them.
   `flutter.md`, `github.md`, `gitlab.md`: the hooks and tasks come from the
   tools of the `base` pack.
7. **`CLAUDE.md` and `README.md`:** the layout (`tools/packs.toml`, the ten
   tools, no `base` tool, no `_base`), the five keys of `tools.toml` with
   `default_envs`, the environment rules of D1 to D5, the docs folders of D18,
   and `test:repo-sync` with its hook in place of "no tests".
8. **`docs/design/`:** bring `bootstrap-config.md` (the `[tools.*]` and
   `[packs.*]` tables, `env`), `mise.md` (`default_envs`, the inherit rule),
   `packs.md` (`[packs.*]` with `env`) and `tools.md` (`default_envs`) up to
   this plan. Add D19 to `fnox.md`.

## Verification

- `rg -n 'tools/base|_base/|\bbase tool\b' README.md CLAUDE.md docs/sync`
  prints nothing.
- Every relative link in `docs/sync/**`, `README.md` and `CLAUDE.md` resolves.
- `mise x -- mise run code:precommit` passes (the wave gate).

## Guardrails

- Touch only docs. Do not edit `docs/plans/**`.
- Move with `mv`; delete with `rm`, never `git rm`.
- Write in ASD-STE100 Simplified Technical English, as the existing docs do.

## Commit

`docs: describe the tools and packs of the split base`
