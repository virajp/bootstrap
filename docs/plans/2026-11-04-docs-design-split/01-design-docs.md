# U1 — The design docs of the merged model

- **Read first:** every file of `docs/design/`, `CLAUDE.md`,
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/docs/setup-sync.md`,
  and the header comment of each file of `sync-engine/`.
- **Lazy-load:** the body of a `sync-engine/` file when a rule needs a
  check against the code; the `index.md` and the unit files of the
  earlier `2026-11-04-*` plans in `docs/plans/` or `docs/plans/archived/`
  (they quote the design rulings);
  `docs/sync/index.md`.

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Contract docs: one place for each rule. `docs/design/` holds the full
contract: new `sync-engine.md`, `repo-sync.md`, `setup-sync.md` (from
macos-setup), and `packs.md`, `tools.md`, `mise.md` change to the merged
model. More files are fine, but CLAUDE.md and `docs/sync/index.md` load
them lazily ("read X when you change Y"). CLAUDE.md keeps layout, change
steps, task rules, test rule, pointers. Script headers keep usage, flags,
a pointer (the 300-line setup-sync header contract goes).
`macos-setup/CLAUDE.md` becomes short (macos.toml, how to edit, how to
sync)."

"Side-specific items: repo front-end keeps members, submodules,
`root_only`, branch model, merge tasks, `direct_commit`, pre-commit render
and `autoupdate`, `repo_sync --check` and its gate. Mac front-end keeps
`mise bootstrap`, the login shell check (`dscl`) with exit 4, mise at
`~/.local/bin/mise`, and the symlink resolve (repo-sync never follows a
link: a link can point outside the repo). The iCloud wait (`dataless`,
`MACOS_SETUP_ICLOUD_TIMEOUT`) goes. [...]"

From this plan:

D3: "sync-engine.md: the rules that both sides share (cache, clone, lock,
engine variables, tool calls, templates, the write order, the parse check,
merge mode, the header and its detection, the manifest and the stale rule,
the values merge, the shared exit codes 2, 3 and 5, bootstrap_commit).
repo-sync.md: the repo front-end; it links to sync-engine.md for the exit
codes. setup-sync.md: the Mac front-end; it describes exit 4 and links to
sync-engine.md for the other codes. packs.md: the packs model. tools.md: the
tools.toml schema, the [files] keys and the rule for the header of a source
file. mise.md: the mise generator, environments, [overrides.mise*], zz-repo
and zz-macos. bootstrap-config.md: the keys of .config/bootstrap.toml.
fnox.md: no change. Each other doc links to the one place."

D4: "No. bootstrap-config.md stays the one place of the keys of
.config/bootstrap.toml; repo-sync.md links to it. setup-sync.md holds the
keys of macos.toml, because no other doc has them."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-macos-tools-apps."

## Edits

Write each rule from the code and the current text, not from memory. When
the code and a text differ, the code wins; report the text as
`DOCS FALSIFIED:`. Each doc starts with an H1, one paragraph on its scope,
and a line that names the other design docs it links to.

1. **`docs/design/sync-engine.md`** (new) — the shared engine:
   - The files of `sync-engine/` and the job of each.
   - The cache `~/.cache/sync-engine/` (`repo/`, `lock/`,
     `macos-generated.tsv`), the clone URL, branch `main`, the normalized
     origin compare, the non-git folder (exit 2), the old clone outside
     `repo/`, the legacy cleanup, the lock and its release (a dry run
     releases it early), and how a stub embeds `sync-engine/clone`
     (`.sync_engine`).
   - The engine variables `SYNC_ENGINE_CLONE`, `SYNC_ENGINE_URL`,
     `SYNC_ENGINE_CACHE`, and the forwarding file with `REPO_SYNC_CLONE`.
   - Tool calls: `mise x <tool>@latest --` from `/`, no pins.
   - Templates: `.tmpl` as the only marker, the `<% %>` delimiters, the mode
     of the output file, the values `.config` and `.all`.
   - The write order of each side (repo: render, parse check, merge, dprint,
     compare; Mac: render, parse check, merge, compare), seeds, the parse
     check, merge mode (`remove`, `remove_items`).
   - The header, its detection in the first 3 lines, and the old texts.
   - The manifest (`<tool><TAB><path>`, first line, written after a clean
     run) and the stale rule.
   - The values merge order (defaults, contributions, `[tools.<name>]`) and
     the rules of `[contributes.<tool>]`.
   - The shared exit codes 2, 3 and 5 (D3); `bootstrap_commit` and its
     write rule.
   - Links to `packs.md`, `tools.md`, `mise.md`, `repo-sync.md`,
     `setup-sync.md`.
2. **`docs/design/repo-sync.md`** (new) — the repo front-end: the entry
   tasks (`setup:configs`, `setup:configs:sync`, the gate
   `code:lint:repo-sync`), how it reads `.config/bootstrap.toml` (keys:
   link to `bootstrap-config.md`, D4), pack `base`, members and
   submodules, `root_only`, the template context (`.members`,
   `.member_folders`, `.member_types`, `.submodules`, `.tools`,
   `.extensions`, `.precommit`, `.config`, `.all`), the branch model and the
   merge tasks, `direct_commit`, the pre-commit render and `autoupdate`,
   `[overrides.gitignore]`, the fnox profile rules (link to `fnox.md` for
   the rest), dprint, `--check` and `--dry-run`, the manifest
   `.config/repo-generated.tsv`, `[claude.plugins.*]` (read by `setup:ai`),
   `requires_found`. For the exit codes, link to `sync-engine.md` (D3).
   Take the text from the sections How the sync operates,
   Templates and Branch model in the merge tasks of `CLAUDE.md`.
3. **`docs/design/setup-sync.md`** (new) — the Mac front-end, from
   `MS/docs/setup-sync.md`, updated to the code of `sync-engine/setup-sync`
   and the rulings: `macos.toml` and its keys, `MACOS_SETUP_CONFIG_PATH`
   (required), the selection, `[overrides.mise.*]` (no `mise_env`),
   `[mise.*]` of `macos.toml` into `zz-macos`, stages and exit 4, the login
   shell check, `mise bootstrap --yes`, mise at `~/.local/bin/mise`, the
   symlink resolve, the manifest `macos-generated.tsv`, the flags
   `--dry-run`, `--no-bootstrap`, `-h`, the entry task
   `updateall:configs` and its lock. Describe exit 4 here, and link to
   `sync-engine.md` for the other exit codes (D3). Drop every old item: the
   iCloud config
   folder and wait, `<tool>.toml` values files, `tools = [...]`,
   `[overrides.<tool>]`, `generated.txt`, the pinned tool versions,
   `macos_setup_commit`, `MACOS_SETUP_REPO_URL`, `MACOS_SETUP_REPO_PATH`.
4. **`docs/design/packs.md`** — the merged packs model for both sides:
   `repo/packs.toml` and `macos/packs.toml`, the entry shape, `"pack:<name>"`,
   `default`, selection with `[packs.<name>]` and `[tools.<name>]`,
   dedupe, pack `base` on each side, `_security_base` as a `depends` (no
   `required`), and the open point (no way to drop one default member).
   Remove each passage of the old model.
5. **`docs/design/tools.md`** — the shared `tools.toml` schema: the keys of
   each side, `stage` (Mac only), `[files."<target>"]` with `mode`,
   `requires`, `requires_found` (repo only), `format`, `remove`,
   `remove_items`, `[defaults]`, `[questions]`, `[contributes.<tool>]`,
   `[[precommit]]`, `[mise.*]` (link to `mise.md`), the folder rules
   (`_` common tools, `zz-repo`), and the rule that two tools must not ship
   the same path, and the rule for the header of a source file (which file
   gets which header, in which comment syntax and at which line; D3). Take
   the key rules from the Layout rule section of
   `CLAUDE.md`.
6. **`docs/design/mise.md`** — the generator of both sides: the valid
   sections (with `[bootstrap.*]` and `[dotfiles]`), the repo environments
   and their rules, the Mac `conf.d/<tool>/mise.toml` with no environment,
   `[overrides.mise.*]` and `[overrides.mise_env.*]`, the render with
   gomplate, `zz-repo` and `zz-macos`. Take the text from the mise config
   section of `CLAUDE.md`. Keep the task sections that it has today.
7. **`docs/design/index.md`** — the references list names every design
   doc, with one line each on what it holds (D3).
8. **`docs/design/bootstrap-config.md`** — it keeps the keys of the repo
   values file (D4). Replace each passage that restates a rule of another
   design doc with a link to it.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check docs/design` passes.
- The three new files exist, and `docs/design/index.md` links to each
  design doc.
- `grep -rn 'iCloud\|generated.txt\|macos_setup_commit\|gomplate@5' docs/design`
  finds nothing, except a line that names an old item as gone.
- Spot checks of "one place": the shared exit codes 2, 3 and 5 are
  described in `sync-engine.md` only; `setup-sync.md` describes exit 4 and
  links to `sync-engine.md`, and `repo-sync.md` links to it too; `grep -l
  'remove_items'
  docs/design/*.md` names `tools.md` and at most `sync-engine.md`.
- Each section of How the sync operates, Templates, Branch model in the
  merge tasks and mise config of `CLAUDE.md` has each of its rules in a
  design doc. Report the result as one `DECIDED:` line, because U3 removes
  those sections.

## Guardrails

- Touch only `docs/design/**`. Report a passage of another doc as
  `DOCS FALSIFIED:`.
- Do not change a rule; move and update its text only. A rule that the code
  does not have is a `GAP:`, not a new rule.
- Write all prose in ASD-STE100 Simplified Technical English. Write
  identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: write the design docs of the merged sync model` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
