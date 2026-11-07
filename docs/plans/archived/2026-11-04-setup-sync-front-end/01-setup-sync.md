# U1 — setup-sync: the Mac front-end on the shared libraries

- **Read first:** every owned file, top to bottom, before editing (the file is
  new; read `sync-engine/repo-sync` as the model of a front-end, and each
  shared library that it sources: `sync-engine/common`, `write`, `select`,
  `values`, `merge`, `manifest`, `mise`, `clone`).
- **Lazy-load:** the old engine
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/macos-setup/.config/mise/tasks/_scripts/setup-sync`
  (the functions that the Facts list, at their lines), the old entry task
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/macos-setup/.config/mise/tasks/updateall/configs/_default`
  (the mise check, `:74-89`).

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - The machine manifest (today `~/.cache/macos-setup/generated.txt`) becomes
>   `macos-generated.tsv`: tab-separated `<tool><TAB><path>`, not TOML. Only
>   setup-sync reads it; bash 3.2 reads TSV with no tool call.

> - Engine variables: `SYNC_ENGINE_CLONE` (temp copy, stub → child task),
>   `SYNC_ENGINE_URL`, `SYNC_ENGINE_CACHE`. Internal only, not user config:
>   test:repo-sync and test:setup-sync use them (today `REPO_SYNC_URL` 8x,
>   `MACOS_SETUP_REPO_URL` 5x, `MACOS_SETUP_REPO_PATH` 3x), and
>   `SYNC_ENGINE_URL`
>   tests an unpushed change. The user saw no value in them and let me keep
>   them on this reason. The forwarding file at the old repo-sync path accepts
>   `REPO_SYNC_CLONE` during the transition.

> - The engine starts gomplate, jq and yq as `mise x <tool>@latest --` from
>   `/`: no pins (drop setup-sync's gomplate@5.3.1, jq@1.8.2, yq@4.54.1), no
>   `#MISE tools.gomplate` directive for the engine. Works before a first sync,
>   on a new Mac, in test scratch folders. The template data stays per
>   front-end; only the delimiters and the gomplate call are shared.

> - Seeds on both sides use the setup-sync form: `[files."<target>"]` (target
>   path, no `.tmpl`), `mode = "seed"`, `requires = [...]` (exact paths
>   relative to the target root, each must exist). `seed = [...]` and
>   `[seed_requires]` go (exit 2 naming the new key). No `[files]` table =
>   `mode = "generated"`. The repo mise-entry `requires` (a file name searched
>   4 folders deep) is a second meaning of the word: solve it in the "mise
>   config" row.

> - The `.tmpl` suffix stays the only template marker; no `template = true`
>   key. **Why:** without the suffix, bootstrap's check-json/toml/yaml hooks,
>   dprint (global config excludes `**/*.tmpl` from shell) and mise (cd into
>   `tools/mise/`) read template sources as real files. The engine finds
>   `<target>` or `<target>.tmpl`; both → exit 2.

> - Stale rule on both sides: remove each file with the header that this run
>   does not write, then its empty parent folders up to the target root; a
>   file without the header is never removed. One manifest format:
>   `<tool><TAB><path>`, sorted, first line `# Written by bootstrap. Do not
>   edit this file.` (reader skips `#` lines), written only after a clean
>   run. Mac: `~/.cache/sync-engine/macos-generated.tsv`. Repo (user's
>   proposal): `.config/repo-generated.tsv`, committed (moves with the branch;
>   a cache copy breaks with clones, worktrees, other machines). Each
>   submodule has its own. The first repo sync has no manifest and removes
>   nothing; `migrate.md` lists old leftovers. No `git grep` header scan.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

> - The Mac side uses no environments (user, 2026-11-04): every generated Mac
>   mise config is `~/.config/mise/conf.d/<tool>/mise.toml`. **Why:** processes
>   without the shell startup files (first `mise bootstrap` of `setup`, Dock
>   apps such as VS Code, launchd, `ssh host cmd`) do not have `MISE_ENV=dev`,
>   and they lose `mise.dev.toml`. The engine keeps the shared code path for
>   `mise.<env>.toml`, so the Mac can use it later; until then a Mac tool with
>   `[mise_env.*]`, or an `env` key in the Mac selection, stops with exit 2.
>   The repo rule "no `conf.d/*/mise.toml`" stays for the repo side.

> - Values row, item 1: the Mac values live in the `macos-setup` repo, made
>   private, at its root (not iCloud). The iCloud config folder never
>   existed (the `tools/` model is not live), so no data moves. A new Mac
>   runs `mise x gh@latest -- gh auth login` in `setup` before the clone. The
>   engine reads a folder of that clone; `MACOS_SETUP_CONFIG_PATH` stays the
>   way to point at it.

> - Values row, item 2: one values file on each side. The values of a tool
>   are keys of its `[tools.<name>]` table, which also selects it; a table
>   for a base tool or a dependency only gives values. No `<tool>.toml`
>   files, no Mac `[overrides.<tool>]`, no "a `<tool>.toml` does not select"
>   rule. Repo: `.config/bootstrap.toml`, `[tools.<name>]` keeps `env` as its
>   only key until item 3. Mac file: `macos.toml` at the root of the private
>   `macos-setup` repo (the user wants the type in the file name; `macos`
>   matches the side name). Repo file keeps `bootstrap.toml` (a rename needs
>   a consumer migration for no function).

> - Values row, item 4: overrides use the repo shape on both sides.
>   `[overrides.mise.<path>]` valid on both (same rules, `# override:`
>   comment); `[overrides.mise_env.*]` → exit 2 on the Mac;
>   `[overrides.gitignore]` repo only (Mac `~/.config/git/ignore` is a git
>   tool file, changed with `[tools.git]` values). `[tools.<name>]` = values
>   for templates; `[overrides.mise.*]` = changes to generated mise config.
>   The repo rule "an override cannot add a key" is the input of item 5.

> - Values row, item 5 (values into generated mise config):
>   1. Map values (packages, aliases, `[env]`, defaults domains) move directly
>      into `[mise.*]` of `tools.toml` (`[defaults.packages]` →
>      `[mise.bootstrap.packages]`); changed with `[overrides.mise.*]`.
>   2. User additions go into `[mise.*]` of `macos.toml`; the engine generates
>      `~/.config/mise/conf.d/zz-macos/mise.toml` (loads last, wins; the Mac
>      twin of `zz-repo`). "An override cannot add a key" stays (else a new
>      alias lands in all 12 tools with `[shell_alias]`).
>   3. The engine renders the generated mise file with gomplate and the
>      tool's values, so a `[mise.*]` string can hold `<% %>` (e.g.
>      `PROJECTS_DIR = "<% .config.projects_dir %>"`); `required` and
>      `test.Fail` checks keep working.
>   4. A conditional section (fish `login_shell` bool) becomes a value:
>      `[mise.bootstrap.user] login_shell = "<% .config.login_shell %>"`.
>      All 6 templated Mac fragments become `tools.toml` tables.

Note: item 3 above is superseded by P08 D6. The engine renders a `tools.toml`
that holds `<%` before it generates, so the generated file holds the rendered
value. A `<% %>` expression sits inside a TOML basic string ("..."), and each
string literal inside the expression is a Go raw string in backticks.

> - Values row, item 6: `bootstrap_commit` on both sides (Mac: in
>   `macos.toml`, replaces `macos_setup_commit`), same write rules. Each Mac
>   sync with a new commit changes `macos.toml` in the macos-setup clone; the
>   user commits it (`ops: sync bootstrap`). The Values row is closed.

> - Phases: `stage` is in the shared `tools.toml` schema; only the Mac
>   front-end reads it (init stage today: bash, fish, zsh); a repo tool with
>   `stage` → exit 2. Exit 4 stays Mac-only; shared exit codes 2, 3, 5.
>   Environments (which config mise loads) and stages (sync order) stay
>   separate ideas.

> - Formatter: the parse check (jq for JSON, yq for YAML/TOML; a target that
>   does not parse → exit 2, old file stays) moves into the shared engine and
>   runs on both sides. The `format` key (`text`/`json`/`yaml`/`toml`,
>   default from the extension) is in the shared `[files]` schema. dprint
>   stays repo-only; repo order: render, parse check, merge, dprint,
>   compare. The Mac does not format.

> - Side-specific items: repo front-end keeps members, submodules,
>   `root_only`, branch model, merge tasks, `direct_commit`, pre-commit render
>   and `autoupdate`, `repo_sync --check` and its gate. Mac front-end keeps
>   `mise bootstrap`, the login shell check (`dscl`) with exit 4, mise at
>   `~/.local/bin/mise`, and the symlink resolve (repo-sync never follows a
>   link: a link can point outside the repo). The iCloud wait (`dataless`,
>   `MACOS_SETUP_ICLOUD_TIMEOUT`) goes. All rows of "The two engines today"
>   are closed (2026-11-04). Next: the other concerns of the artifact (risks,
>   phases, open question 4 on git history).

From the one merged packs model, verbatim:

> - The selection is `[packs.<name>]` and `[tools.<name>]` tables on both
>   sides. setup-sync drops `tools = [...]` in `macos-setup.toml`.

> - `[packs.<name>]` adds each `default = true` member (option A). A new
>   `default = true` member reaches every consumer of the pack; a
>   `default = false` member arrives only with `[tools.<name>]`. Still open,
>   as in `docs/design/packs.md`: no way to drop one default member.

Plan decisions:

D1: "No. When MACOS_SETUP_CONFIG_PATH is unset or empty, setup_sync stops with
exit 2 and names the variable. When the folder does not exist, it also stops
with exit 2. The macos-setup tool sets the variable in its generated mise
config (P10)."

D2: "description, stage and depends (each optional; depends defaults to []),
[defaults], [questions.<key>], [contributes.<tool>], [files."<target>"] and
[mise.*]. The repo-only keys default_envs, root_only, extensions,
[[precommit]] and [mise_env.*], and requires_found in a [files] table, stop the
run with exit 2 and a message that says the key is for repo tools only. Any
other key also stops the run with exit 2."

D3: "U1 calls the own-config entry of sync-engine/mise (P08 item 7), which
generates the [mise.*] tables of macos.toml into
~/.config/mise/conf.d/zz-macos/mise.toml with the header `# Synced from
bootstrap: macos.toml [mise]. Do not edit this file.` and `# The next sync
replaces it, and a manual change is lost.` U1 does not build the file itself.
setup-sync writes it in the main stage, and its manifest line is
`zz-macos<TAB>.config/mise/conf.d/zz-macos/mise.toml`. A Mac tool folder named
zz-macos, or one whose name sorts after zz-macos in byte order, stops the run
with exit 2."

D4: ".config (the values of the tool), .all (the values of every selected
tool, by name), .tools (the selected tool names), .env (HOME, USER and
MACOS_SETUP_CONFIG_PATH) and .sync_engine (the absolute path of the
sync-engine/ folder of SYNC_ENGINE_CLONE, as the repo front-end gives it since
P01). The delimiters are <% %>."

D5: "Port the rule of the old engine: setup-sync selects pack base and its
depends only. A clean real run then calls _se_record_commit (P02 item 9) with
the new-file header `# The values of the Mac sync of bootstrap. See
docs/setup/index.md.`, which creates macos.toml with that header and the
bootstrap_commit line. A symlinked macos.toml is written through to its real
file."

D6: "No. U1 edits only sync-engine/setup-sync. Each Mac-only step (the symlink
resolve, the stages, the login shell check, mise bootstrap, the mise path
check) is a function in setup-sync that calls the shared libraries. When a
library has no entry that a wrapper can use, U1 returns UNRESOLVED and names
the missing entry."

D7: "In setup_sync, as its first check after the flags. The check and its
message are the ones of the old entry task, so a ~/.local/bin/mise link into
another folder also fails."

D11: "Each message keeps the `setup-sync: ` prefix of the log function. A
message that named <tool>.toml, [overrides.<tool>] or macos-setup.toml now
names [tools.<tool>] of macos.toml, or macos.toml."

D13: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there."

D16: "Yes. U1 calls sync-engine/manifest with the link mode macos (P05 item
5): the stale rule resolves links, removes the real file, then removes empty
parents up to $HOME, never past a symlinked folder. On the Mac, ~/.config/mise
is a link into dotfiles/, and the mise config of a deselected tool must go.
The old case "a stale file through a link" stays. U1 passes the mode and edits
no library (D6)."

D17: "The rule of the old engine stays (B:1564-1580): a failure in the init
stage stops the run before the main stage, with no mise bootstrap and no
manifest write. Inside one stage, a failed file stops no other file."

## Edits

1. **`sync-engine/setup-sync`** (new). A library that a task sources, not a
   task: `#!/usr/bin/env bash`, then a short header comment. It has no
   bootstrap sync header, because the sync never copies `sync-engine/`.
   1. **Header comment.** Usage
      (`source "${SYNC_ENGINE_CLONE}/sync-engine/setup-sync"` then
      `setup_sync [--dry-run] [--no-bootstrap]`), the flags, the variables
      (`SYNC_ENGINE_CLONE`, `SYNC_ENGINE_CACHE`, `MACOS_SETUP_CONFIG_PATH`),
      the return codes (0 done; 1 other failure, also a failed
      `mise bootstrap`; 2 usage or config error; 3 a tool or a pack that
      `macos/` does not have; 4 the init stage changed). At most 60 lines.
      P15 moves the contract into `docs/design/setup-sync.md`; do not copy the
      300-line contract of B.
   2. **Libraries.** Source `common`, `write`, `select`, `values`, `merge`,
      `manifest` and `mise` from the folder of this file, the way
      `sync-engine/repo-sync` sources them. Use their log function with the
      prefix `setup-sync`, their mode and cleanup functions, their header
      detection, their record commit and their `mise x <tool>@latest --`
      wrappers. Write no copy of a function that a library has.
   3. **`setup_sync` flags.** `--dry-run`, `--no-bootstrap`, `-h`/`--help`
      (usage on stdout, return 0). Any other argument: the message
      `unknown flag '<arg>'`, the usage on stderr, return 2.
   4. **First checks, in this order, each with return 2:**
      - mise (D7): resolve `command -v mise` through every link, as the old
        entry task does (`:74-89`); when the real path does not end in
        `/.local/bin/mise`, log the old message
        ``mise must be installed with `curl https://mise.run | sh` at
        ~/.local/bin/mise, not from a package manager (found: <path>)``.
      - `SYNC_ENGINE_CLONE` set and a folder with `sync-engine/` and
        `macos/tools/`; else
        `SYNC_ENGINE_CLONE is not a clone of bootstrap - run mise run
        updateall:configs`.
      - `MACOS_SETUP_CONFIG_PATH` (D1): unset or empty →
        `MACOS_SETUP_CONFIG_PATH is not set - set it to the folder of
        macos.toml`;
        not a folder →
        `the config folder <path> does not exist - set MACOS_SETUP_CONFIG_PATH
        to the folder of macos.toml`.
   5. **`macos.toml`.** Read `<config>/macos.toml` with gomplate as one TOML
      document (D5: a missing file is an empty selection). Accept only these
      top-level keys, else return 2 with a message that lists them:
      - `bootstrap_commit`: a string.
      - `[tools.<name>]`: a table of values. An `env` key → return 2 with
        `[tools.<name>] env is for repo tools only - the Mac side has no
        environments`.
      - `[packs.<name>]`: an empty table; a key in it → return 2.
      - `[overrides.mise.<path>]`: passed to `sync-engine/mise`.
        `[overrides.mise_env.*]` → return 2 with a message that the Mac side
        has no environments. Any other key of `[overrides]` → return 2 with
        a message that names `[tools.<name>]` for template values and
        `[overrides.mise.<path>]` for mise config.
      - `[mise.*]`: the user config of D3.
      - `tools = [ ... ]` (a list) → return 2 with
        `tools = [...] is the old format - write a [tools.<name>] table for each
        tool`.
   6. **Tool folders.** For each folder of `$SYNC_ENGINE_CLONE/macos/tools/`
      (byte order): a folder without `tools.toml` → return 2; a folder named
      `zz-macos` or sorted after it → return 2 (D3). Validate each
      `tools.toml` with the keys of D2. Port the old-key messages of B:
      `seed` and `[seed_requires]` name the `[files]` table, `stage = "shell"`
      names `stage = "init"`, and a `stage` other than `init` names the only
      stage. Validate `[files]` keys (`mode`, `requires`, `format`, `remove`,
      `remove_items`) and `[questions]` keys (`question`, a string, required;
      `detect`; `example`) with the shared rules where a library has them.
   7. **Selection.** Call `sync-engine/select` with the root `macos/`
      (`macos/tools/`, `macos/packs.toml`). Pack `base` is always added; a
      `packs.toml` without `[pack.base]`, or a base without
      `macos-setup = { default = true }`, → return 2 with
      `pack base does not have the tool macos-setup`. Then the
      `default = true` members of each `[packs.<name>]`, each
      `[tools.<name>]`, and the depends closure. An unknown tool or pack →
      return 3.
   8. **Values.** Call `sync-engine/values` with the `[defaults]`,
      `[contributes.*]` and `[tools.<name>]` keys of the selection. The list
      rules for base and dependency tools are the library's (P07).
   9. **File plan** (port `_ss_plan`, `B:931-1005`). For each selected tool,
      each file of its folder except `tools.toml` gives a target relative to
      `$HOME`, without `.tmpl`. Return 2 for: a `<target>` and a
      `<target>.tmpl` in one tool; two tools that give one target; a
      `[files]` table whose target no file gives; a file under
      `.config/mise/conf.d/` (message: the sync generates it from `[mise.*]`
      of `tools.toml`). Keep each file's `mode`, `requires`, `format`,
      `remove` and `remove_items`.
   10. **Stages** (port `B:1564-1618` and the stage split of `_ss_resolve`).
       The init stage is each selected tool with `stage = "init"` and,
       transitively, its depends; the main stage is each other tool and
       `zz-macos`. In each stage, write the mise files first (the generated
       `conf.d/<tool>/mise.toml` files of its tools and each shipped
       `.config/mise/**` file), then the other files.
   11. **Render and write.** For each file: render a `.tmpl` with gomplate and
       the context of D4 (else copy the bytes); then the write library (parse
       check, compare, mode, dry-run diff, rename, seed gate) and, for
       `mode = "merge"`, the merge library. Before the write, resolve the
       target through each symlink of the file or of a parent folder (port
       `_ss_real`, `B:344-375`): the compare, the temp file, the rename and a
       stale removal act on the real path. A link whose source is missing is
       written as a new file at the source path; a seed behind such a link
       counts as missing. A file with the header where a target needs a
       folder is removed; any other blocker fails that target (port
       `_ss_unblock` through the shared unblock, if a library has it).
   12. **mise config.** For each selected tool with `[mise.*]`, call
       `sync-engine/mise` with side `macos` and the `[overrides.mise.*]` of
       `macos.toml`; it writes `~/.config/mise/conf.d/<tool>/mise.toml`. For
       `[mise.*]` of `macos.toml`, call the own-config entry of
       `sync-engine/mise` (P08 item 7, D3); do not build the `zz-macos` file
       in setup-sync. The generated files go through the same write step as
       other files.
   13. **Manifest.** Call `sync-engine/manifest` with the file
       `<cache>/macos-generated.tsv` (the cache is the expression of
       `sync-engine/clone`:
       `${SYNC_ENGINE_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/sync-engine}`)
       the target root `$HOME` and the link mode `macos` (D16). Port the
       stage-aware lines of
       `_ss_generated_lines` and `_ss_write_generated`: the init stage applies
       the stale rule to the lines of init-stage tools only, and at exit 4 it
       rewrites only those lines; the main stage applies it to every other
       line (also the lines of a tool that is no longer selected) and writes
       the whole manifest. Seeds and merge files are not listed.
   14. **Run flow** (port `setup_sync`, `B:1540-1618`). A dry run renders
       both stages, prints the stale removals, prints
       `would set bootstrap_commit <commit>` and never stops. A real run: the
       init stage when it is not empty (render, stale, `mise bootstrap --yes`
       unless `--no-bootstrap`, the login shell compare with
       `dscl . -read ~ UserShell`); a change returns 4 with
       `open a new terminal, then run: mise run updateall:configs`. Then the
       main stage, the manifest, `mise bootstrap --yes` unless
       `--no-bootstrap`, and `bootstrap_commit`. Port `_ss_bootstrap`
       (`B:1413-1421`) as it is. Inside one stage, a failed file stops no
       other file. A failure in the init stage stops the run before the main
       stage (D17). After any failure, the run keeps the manifest, runs no
       `mise bootstrap`, writes no commit, and returns the highest code.
   15. **`bootstrap_commit`.** The commit is
       `git -C "$SYNC_ENGINE_CLONE" rev-parse HEAD`. Write it with the record
       commit `_se_record_commit` of `sync-engine/common`, key
       `bootstrap_commit`, file `<config>/macos.toml`, and the new-file header
       `# The values of the Mac sync of bootstrap. See docs/setup/index.md.`
       (D5). A symlinked `macos.toml` is written through to its real file.
   16. **Not ported:** `_ss_download`, `_ss_dataless`,
       `MACOS_SETUP_ICLOUD_TIMEOUT`, `_ss_snapshot`, the selection template of
       `macos-setup.toml`, the pins of `B:281-283`, the header constant of
       `B:285`, any dprint call.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `/bin/bash -n sync-engine/setup-sync` passes (bash 3.2).
- `/bin/bash -c 'source sync-engine/setup-sync && setup_sync --help'` exits 0
  and prints the usage.
- `grep -nE 'dataless|ICLOUD|gomplate@5|jq@1|yq@4|dprint|Generated by
  macos-setup' sync-engine/setup-sync`
  finds nothing.
- `grep -n 'macos-generated.tsv' sync-engine/setup-sync` and
  `grep -n 'zz-macos' sync-engine/setup-sync` each find a line.

## Guardrails

- Touch only `sync-engine/setup-sync` (D6). Do not edit another file of
  `sync-engine/`, the old engine in `MS`, or a test.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- Start gomplate, jq and yq only through the wrappers of `sync-engine/common`
  (`mise x <tool>@latest --` from `/`). System commands (`git`, `awk`, `sed`,
  `dscl`, `stat`) get no prefix.
- Write only in `$HOME`, the sources of its symlinks, the cache folder, the
  `bootstrap_commit` line of `macos.toml`, and temp folders that the cleanup
  removes.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the setup-sync front-end on the shared engine` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
