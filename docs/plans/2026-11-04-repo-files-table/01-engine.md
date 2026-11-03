# U1 — Engine: [files] tables, requires_found and merge mode

- **Read first:** every owned file, top to bottom, before editing
  (`sync-engine/write`, `sync-engine/repo-sync`; `sync-engine/merge` is new).
- **Lazy-load:** `MS/tools/macos-setup/.config/mise/tasks/_scripts/setup-sync`
  (read only: the header B:64-84 and the merge rules B:178-214, `_SS_MERGE_JQ`,
  `_ss_merge`,
  the `[files]` checks of `_ss_manifests` B:476-543 and `_ss_file`),
  `MS/.config/mise/tasks/test/setup-sync` (:1246-1347, the messages of the
  config errors), `sync-engine/common`.

## Ruling

From the design rulings of 2026-11-04:

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

> - Merge mode (`mode = "merge"`, JSON/YAML/TOML, `remove`, `remove_items`)
>   moves into the shared engine and is available to repo tools. No repo
>   tool uses it until a real need (candidates: `.vscode/settings.json`,
>   `.claude/settings.json`). Repo side order: merge, dprint format, compare.
>   `[overrides.gitignore]` stays as it is. Merge mode = tool author declares
>   enforced keys; overrides = consumer changes defaults; two mechanisms. The
>   three override shapes (`[overrides.gitignore]`, `[overrides.mise*]`, Mac
>   `[overrides.<tool>]`) get aligned in the "Values" row.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

> - Formatter: the parse check (jq for JSON, yq for YAML/TOML; a target that
>   does not parse → exit 2, old file stays) moves into the shared engine and
>   runs on both sides. The `format` key (`text`/`json`/`yaml`/`toml`,
>   default from the extension) is in the shared `[files]` schema. dprint
>   stays repo-only; repo order: render, parse check, merge, dprint,
>   compare. The Mac does not format.

From this plan:

D5: "The seed key goes from every tools.toml, the empty ones too. A seed key
with any value stops the sync with code 2 and names [files.\"<target>\"] mode =
\"seed\". A tools.toml now has four keys that the sync reads without a
default: default_envs, depends, root_only and extensions."

D7: "It does not write the target. A generated target keeps the behavior of
the mise files list of today: a copy with the bootstrap header is removed (dry
run: would remove <path> (no <file>)), else the log says skip file <path> (no
<file>). A seed logs skip seed <path> (no <file>). A merge target logs skip
merge <path> (no <file>) and stays as it is. <file> names the first entry that
is not there, in the order of the lists: requires first, then requires_found.
Plan 5 replaces the removal with the manifest."

D8: "Every tool of the clone, selected or not, in the same pass that reads all
manifests today (_rs_mise_index). The rules: a known key in each [files]
table, a known mode and format, a target that a file of the tool produces
(<target> or <target>.tmpl), no tool folder with both <p> and <p>.tmpl, merge
only with json, yaml or toml, remove and remove_items only with merge and of
the right shape, the old keys. Thus the test of bootstrap finds an error
before a consumer selects the tool."

D9: "requires: a list of exact paths relative to the repository root
(MISE_PROJECT_ROOT); each must exist (-e); an absolute path or a .. part stops
with code 2. requires_found: a list of file names; each must be found by the
rule of `_rs_has_file` (git tracks it or does not ignore it, at the root or down
to 4 folders below it, no submodule); a name with a / stops with code 2. Both
lists must be non-empty when present. A target with both lists needs every
entry of both."

D10: "sync-engine/merge holds a port of `_SS_MERGE_JQ` and `_ss_merge` with the
rules of the setup-sync header (Merge files). The repo order is render, parse
check, merge with the live target, dprint, compare. The repo side follows no
symlink: the target path is the path in the repository. A merge target has no
header, is never removed by the sync, and keeps the mode of an existing target
(a new one takes the mode of its source). The format of a merge target is its
format key, else its extension; the JSONC rule of plan 2 does not apply to
it."

D11: "One message, code 2, per old key per tools.toml (an entry with an old
requires and an old files gives two): '<tools root>/<tool>/tools.toml: seed is
not supported; use [files.\"<target>\"] with mode = \"seed\"'; '<tools
root>/<tool>/tools.toml: [seed_requires] is not supported; use requires in
[files.\"<target>\"]'; '<tools root>/<tool>/tools.toml: files in
[mise.tools.<x>] is not supported; use requires_found in
[files.\"<target>\"]'; and for requires = \"<name>\" in a [mise.tools.<x>]
entry: '<tool>/tools.toml: requires of mise tool <x> is the old key - write
requires_found = [\"<name>\"]'."

D12: "The generated mise.<env>.toml drops requires_found, as it drops
requires today. In a [mise_env.<env>.tools] entry, requires_found, requires and
files stop the sync with code 2 and the message says that the key is valid
only in [mise.tools]."

D13: "No. [overrides.mise.tools.<x>.requires_found] is locked like envs,
depends and requires, and the lock message names requires_found."

D14: "No. A merge source carries no two-line bootstrap header, because the
app owns the target. A YAML or TOML merge source may carry the comment #
Merged by bootstrap: <path>; a JSON merge source carries no comment. The
engine strips nothing. The test fixtures follow this. The docs unit rewrites
the header rule of CLAUDE.md to: each file that is neither a seed nor a merge
source."

D15: "_se_unblock of sync-engine/write gets a parameter: a list of paths to
spare (the seeds and the merge targets). The repo side passes the list; it has
no wrapper of its own."

D16: "Mode 755, as every file of sync-engine/ that has a shebang (plan 1,
D18)."

## Edits

1. **`sync-engine/merge`** (new, mode 755 (D16); shebang and a short
   comment, no header).
   - Port `_SS_MERGE_JQ` as `_SE_MERGE_JQ` and `_ss_merge` as
     `_se_merge <src> <live> <format> <remove-json> <remove-items-json> <out>
     <show>`
     (or an equal signature; state it in the comment). Keep the rules of the
     setup-sync header: removals, then the deep merge where the source wins
     a scalar and a list adds each new item once; an absent, empty or
     whitespace-only target (YAML: comments only) gets the source; a live
     target that does not parse, is not a table, or has two YAML documents
     fails with 2 and stays; a source that is not a table or has two YAML
     documents fails with 2; YAML and TOML keep the comments of the live
     file through yq. Replace the pinned calls with `_se_jq` and `_se_yq`.
     Remove the symlink resolve and the iCloud code; the caller gives the
     paths.
2. **`sync-engine/write`** — `_se_format` already takes an explicit format.
   Add the helpers that both sides need for `[files]`:
   - `_se_files_key_ok <key>`: one of `mode`, `requires`, `requires_found`,
     `format`, `remove`, `remove_items`.
   - The shape checks of `remove` (a list of non-empty lists of strings) and
     `remove_items` (a list of tables with `path`, a non-empty list of
     strings, and `items`, a list), with the messages of the setup-sync test
     (`remove must be a list of key paths`, `remove_items must be a list of
     tables`, `remove works only with mode = "merge"`,
     `mode = "merge" needs a json, yaml or toml target`), each prefixed with
     `<tool>/tools.toml: [files."<target>"] `.
3. **`sync-engine/repo-sync`**
   - `_RS_MANIFEST_TEMPLATE`: remove the `seed`, `seed_requires`,
     `mise_requires` and `mise_files` lines. Add one line per `[files]`
     entry key, for example `files <target> <key> <json-value>` (use
     `data.ToJSON` or the gomplate equal; verify with Context7), one line
     `mise_found <mise-tool> <name>` per item of a mise entry's
     `requires_found`, and one line `old <key> <detail>` for each old key
     that is present: the top-level `seed`, `seed_requires`, and the mise
     entry keys `requires` and `files`.
   - `_rs_mise_index`: for each tool of the clone, apply D8, D9 (the shape
     part) and D11. Each problem stops with code 2 with the exact message of
     D11; `<tools root>` is `tools` today. A source check: for each `[files]`
     target, `tools/<tool>/<target>` or `tools/<tool>/<target>.tmpl` must be
     a file; for each file of the tool folder, `<p>` and `<p>.tmpl` must not
     both exist (`<tool>: <p> and <p>.tmpl both exist - keep one`).
   - `_rs_resolve`: build the per-file data from the `files` lines in place
     of `_RS_SEEDS`, `_RS_SEED_REQUIRES` and `_RS_FILE_REQUIRES`: the mode,
     the format, `requires`, `requires_found`, `remove` and `remove_items` of
     each `<tool>/<target>`. A mise entry with `requires_found` replaces
     `_RS_MISE_REQUIRES` (every name must be found).
   - `_rs_file_skips`, `_rs_skip_required`: evaluate both lists of each
     `[files]` target (D9) and act by mode (D7). The `requires` check uses
     `-e "$MISE_PROJECT_ROOT/<path>"`.
   - `_rs_mise_env`: a mise tool with `requires_found` gets its environments
     only when every name is found.
   - `_rs_file`: a seed is a `[files]` target with `mode = "seed"`; the seed
     gate is `_se_seed_kept`. The format is `_se_format "$out" "$tmp" "<format
     key>"`.
     A merge target: render, parse check, then `_se_merge` with the live
     target at `$MISE_PROJECT_ROOT/<target>`, then the dprint stage, then
     `_rs_file_finish`. `_rs_file_finish` keeps the mode of an existing merge
     target (D10).
   - Every removal skips a merge target and a seed. `_se_unblock` (in
     `sync-engine/write`) gets a parameter with the list of paths to spare;
     the repo side passes its seeds and merge targets, with no wrapper (D15).
   - The mise generation: drop `requires_found` from the generated
     `mise.<env>.toml`, as `requires` today. A `requires_found`, `requires`
     or `files` key in a `[mise_env.<env>.tools]` entry stops with code 2:
     `<tool>/tools.toml: <key> of mise tool <x> in [mise_env.<env>.tools] is
     valid only in [mise.tools]`
     (D12).
   - The overrides: add `requires_found` to the locked keys of a mise tool;
     the lock message names it (D13).
   - A merge source is read as it is: the engine strips no comment (D14).
   - The header comment: the manifest keys (four required keys, the
     `[files]` table and its keys, `requires_found` in a mise entry), the
     merge order, and the new code 2 problems.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n` passes on `sync-engine/merge`, `sync-engine/write` and
  `sync-engine/repo-sync`.
- `grep -nE
  '_RS_SEEDS|_RS_SEED_REQUIRES|_RS_FILE_REQUIRES|mise_files|seed_requires <'
  sync-engine/repo-sync`
  finds only the lines that report an old key.
- `grep -n '5.3.1\|1.8.2\|4.54.1' sync-engine/merge` finds nothing.
- `test:repo-sync` runs green only after U2 and U3; it is U3's gate.

## Guardrails

- Do not touch `tools/**` (U2) or the test task (U3).
- Do not change `[overrides.gitignore]` or the overrides of mise.
- No tool of bootstrap gets a merge target in this plan.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- gomplate, jq and yq run only through `_se_gomplate`, `_se_jq` and
  `_se_yq`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add [files] tables, requires_found and merge mode to repo tools` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
