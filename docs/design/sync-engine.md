# Sync engine design

This document holds the rules that both front-ends of the sync share: the cache
and the clone, the engine variables, the tool calls, the templates, the write of
one target, merge mode, the header, the manifest and the stale rule, the values
merge, `bootstrap_commit`, the exit codes and the release of the engine. It is
the one place of these rules.

See also: [packs.md](packs.md), [tools.md](tools.md), [mise.md](mise.md),
[repo-sync.md](repo-sync.md), [setup-sync.md](setup-sync.md).

## The files of `sync-engine/`

`sync-engine/` is not a tool. The sync never copies it to a consumer as a
folder.

| File         | Job                                                                                                                                                                                                        |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `clone`      | The cache, the lock, the fetch and the copy of the clone. A template of an entry task embeds this file.                                                                                                    |
| `common`     | The log, the output of a run, the file mode, the clean-up and its traps, the header detection, the record of the commit and the tool wrappers.                                                             |
| `write`      | The write of one target: the format, the parse check, the temp file, the seed gate, the unblock of a folder path, the checks of a `[files]` table, the compare, the mode, the dry-run diff and the rename. |
| `merge`      | Merge mode.                                                                                                                                                                                                |
| `select`     | The pack expansion and the depth-first walk of the `depends`.                                                                                                                                              |
| `manifest`   | The read and the write of the manifest, and the stale rule. Its `macos` link mode resolves each link, removes the real file and prunes up to the target root.                                              |
| `values`     | The merge of the values of the tools.                                                                                                                                                                      |
| `mise`       | The mise config generator, the environments of each tool and the checks of `[overrides.mise*]`. See [mise.md](mise.md).                                                                                    |
| `repo-sync`  | The repo front-end, the function `repo_sync`. See [repo-sync.md](repo-sync.md).                                                                                                                            |
| `setup-sync` | The Mac front-end, the function `setup_sync`. See [setup-sync.md](setup-sync.md).                                                                                                                          |
| `VERSION`    | The pin of the Go engine. See [The release of the engine](#the-release-of-the-engine).                                                                                                                     |

Each front-end sources `common`, `write`, `merge`, `select`, `manifest`,
`values`
and `mise` from its own folder.

### The Go module

`sync-engine/` is also the Go module `github.com/virajp/bootstrap/sync-engine`
of the new engine. The Go engine replaces the bash files phase by phase. The
bash files stay until phase P7. Until then, the Go engine keeps each output
line, file and exit code of the bash engine.

| Path                 | Job                                                                                                                                                      |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go.mod`             | The module.                                                                                                                                              |
| `cmd/sync-engine`    | The command, with the subcommands `repo`, `macos` and `version`. `version` prints `main.version`, which the release build sets. A local build has `dev`. |
| `internal/fail`      | The typed errors of the exit codes, and a collector where the highest code wins. See [The exit codes](#the-exit-codes).                                  |
| `internal/report`    | The output of a run, with the same bytes as `common`. See [The output of a run](#the-output-of-a-run).                                                   |
| `internal/source`    | The read of the clone: the packs file, the `tools.toml` of each tool, the files of a tool folder, and the checks of these files.                         |
| `internal/values`    | The values files of a repo and of the Mac, and the values merge. See [The values merge](#the-values-merge).                                              |
| `internal/selection` | The pack expansion, the checks of the packs, the depth-first walk of the `depends`, and the stages of the Mac.                                           |
| `internal/tmpl`      | `text/template` with the delimiters `<%` and `%>`, and our own copy of the gomplate functions that the templates use. It does not import gomplate.       |
| `internal/format`    | The parse check and merge mode: JSON in the form of jq, YAML and TOML through the yq command in-process. See [Merge mode](#merge-mode).                  |
| `internal/header`    | The header of a synced file: its two lines, and its detection. See [The header](#the-header).                                                            |
| `internal/target`    | The write of one target, and `bootstrap_commit`. See [The write of one target](#the-write-of-one-target).                                                |
| `internal/manifest`  | The manifest and the stale rule, in the link modes `repo` and `macos`. See [The manifest](#the-manifest-and-the-stale-rule).                             |
| `internal/misegen`   | The mise config generator of both sides: the environments, the checks of `[overrides.mise*]`, and the `conf.d` files. See [mise.md](mise.md).            |
| `internal/macos`     | The Mac front-end, `sync-engine macos`: the Go copy of `setup-sync`. See [setup-sync.md](setup-sync.md).                                                 |

The task `test:go` of this repository runs `go vet`, a `gofmt` check and
`go test` in the module. The golden files of `internal/report` hold the output
of the bash functions of `common`. The golden files of `internal/tmpl` hold the
output of the gomplate CLI for each `.tmpl` file and each inline template.

One difference from bash: to find a terminal on stderr, Go accepts any
character device, and the bash test `-t 2` accepts only a terminal. The two
differ only for a device such as `/dev/null`, where no output can be seen.

## The cache and the clone

The cache is `${XDG_CACHE_HOME:-~/.cache}/sync-engine/`. `SYNC_ENGINE_CACHE`
changes the folder. All the consumers of the machine and the Mac side share it.

| Entry                 | Use                                                                            |
| --------------------- | ------------------------------------------------------------------------------ |
| `repo/`               | One shallow clone of this repository, branch `main`.                           |
| `lock/`               | The lock folder, with the `pid` of its holder.                                 |
| `lock.takeover/`      | The short lock of the removal of a dead lock.                                  |
| `macos-generated.tsv` | The manifest of the Mac side. See [setup-sync.md](setup-sync.md#the-manifest). |

The URL is `https://github.com/virajp/bootstrap.git`. `SYNC_ENGINE_URL` changes
it. `_se_fetch` does these steps:

- It clones when `repo/` is missing or is an empty folder.
- It fetches the branch `main` when `repo/` is a clone whose `origin` matches
  the
  URL. It fetches from the URL, not from `origin`, because the compare drops the
  scheme, and the transport must not come from the cached `origin`. Then it runs
  `git reset --hard FETCH_HEAD` and `git clean -ffdx`. Do not keep work in the
  clone.
- It clones again when `repo/` is a clone with another `origin`, for example
  after
  a test with `SYNC_ENGINE_URL`.
- It stops when `repo/` is not a git clone and is not empty. It does not delete
  the
  folder.

The origin compare is normalized. Two URLs match when they are equal without the
scheme, the user name, the `user@host:path` form, a trailing `/` and a trailing
`.git`. Thus the SSH and the HTTPS URL of one repository match. The compare
reads
the raw config value, because `git remote get-url` expands `insteadOf`.

A git call that uses the network has no terminal prompt and uses SSH
`BatchMode`.
Thus a missing key fails and does not wait for a password.

`_se_copy` copies the clone to a temporary folder. The sync reads only this
copy.
Thus a change to the sync code applies in the same run, and the next run can
update the cache.

The sync uses the pushed `main` branch. Local changes have no effect on
consumers
until you push them. To test a change before you push it, set `SYNC_ENGINE_URL`
to
a local clone with a commit of the change on `main` of that clone. Use
`mise run setup:configs --dry-run` to show a diff. The next run without that
variable clones the default URL again.

### The old clone at the cache root

A git repository at the cache root is the old layout. When its raw `origin`
normalizes to the URL, the run removes each entry of the cache root, except
`repo/`, `lock/`, `lock.takeover/` and `macos-generated.tsv`. Another `origin`
or
no `origin` stops the task, and the task deletes nothing in the cache root. A
`.git` file (a worktree or a submodule) has no config file there, so it also
stops
the task.

### The legacy cleanup

The old stubs used `bootstrap/` and `bootstrap.lock` in
`${XDG_CACHE_HOME:-~/.cache}`. Each run removes these two while it holds the old
lock. When a live old stub holds that lock, the run skips the removal, and a
later
run does it. A failed removal does not fail the sync. The line
`Removed the old cache <path>` prints only when the old folder existed before
the removal and does not exist after it, so a run with only the old lock prints
nothing. The old Mac cache has its
own
step. See [setup-sync.md](setup-sync.md#the-entry-task).

### The lock

`lock/` makes a second run wait. `_se_lock` creates the folder and writes its
`pid`. A waiter removes the lock of a dead holder, and a lock without a holder
that
is older than one minute. The removal has its own lock, `lock.takeover/`:
without
it, a second waiter can remove the lock that the first waiter took after the
removal.

The release depends on the entry task:

| Entry task          | Release                                                       |
| ------------------- | ------------------------------------------------------------- |
| `setup:configs`     | After the copy of the clone, in each run.                     |
| `updateall:configs` | A dry run: after the copy. A real run: at the end of the run. |

### How an entry task embeds `clone`

An entry task must run before the consumer has a clone. Thus the sync embeds
`sync-engine/clone` in the template of the task when it renders it:
`<% file.Read (printf "%s/clone" .sync_engine) %>`. `.sync_engine` is the
absolute
path of the `sync-engine/` folder of the clone copy of the run. The entry task
runs
`_se_clone_init`, `_se_lock`, `_se_legacy_cleanup`, `_se_fetch` and `_se_copy`,
and
it sets a trap with `_se_clone_cleanup`.

Then the entry task starts its child task in a new process with
`SYNC_ENGINE_CLONE`
set to the copy. The child task loads the front-end from that copy.

## The engine variables

| Variable              | Use                                                                                                           |
| --------------------- | ------------------------------------------------------------------------------------------------------------- |
| `SYNC_ENGINE_CLONE`   | The clone copy of the run, which the entry task gives its child task.                                         |
| `SYNC_ENGINE_URL`     | The clone URL. A test uses it for a change that is not pushed.                                                |
| `SYNC_ENGINE_VERBOSE` | `1` prints the log line of each target in place of the list. See [The output of a run](#the-output-of-a-run). |
| `SYNC_ENGINE_CACHE`   | The cache folder. A test uses it.                                                                             |
| `REPO_SYNC_CLONE`     | The same copy, for a member that still has the old sync task. See the forwarding file below.                  |

The forwarding file is `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
of
this repository. The sync does not ship it. An old `setup/configs/sync` loads it
from the clone with `REPO_SYNC_CLONE`, and it loads `sync-engine/repo-sync` of
the
same clone. It keeps its old header. The cleanup plan removes it.

## The tool calls

The wrappers `_se_gomplate`, `_se_jq` and `_se_yq` of `common` start gomplate,
jq
and yq from `/`, so no mise config of a repository selects a version. The engine
calls gomplate, jq and yq only through these wrappers, and it has no pinned
version.

| State of the tool | Call                                     |
| ----------------- | ---------------------------------------- |
| Installed         | `MISE_OFFLINE=1 mise x <tool> -- <tool>` |
| Not installed     | `mise x <tool>@latest -- <tool>`         |

An installed tool runs with no network. mise resolves `@latest` from its list of
versions, and with `MISE_OFFLINE=1` it finds no versions, so an installed tool
runs
by its name alone. Thus the sync operates before a first sync, on a new machine
and
in a scratch folder. No task needs a `#MISE` line for gomplate, jq and yq, and
the sync operates in a shell without `MISE_ENV=dev`. Each path that a wrapper
gets is absolute.

gomplate reads the TOML files: the values file, the packs file (an awk scan also
gives the file order of its members), each `tools.toml` and the consumer file
`.config/fnox.toml`.

## Templates

- `.tmpl` is the only marker of a template. gomplate renders each `.tmpl` file,
  and
  the output file has no `.tmpl` suffix. Any other file is copied byte for byte.
- A JSON, YAML or TOML target is also a template when its source ends in
  `.tmpl`.
- The output file takes the mode of the `.tmpl` file. Make a `.tmpl` that
  renders a
  task or a script executable (`755`).
- The delimiters are `<% %>`. Do not use `{{ }}` or `[[ ]]` for template logic.
  mise uses `{{ }}` and bash uses `[[ ]]`, and both must reach the output
  unchanged.
- `.config` is the values of the tool of the template. `.all` is the values of
  each
  selected tool, by tool name. A tool without values has an empty table. See
  [The values merge](#the-values-merge).
- A top-level key `all` or `config` of the repo values file does not reach a
  template.

The other values of the context are different on each side. See
[repo-sync.md](repo-sync.md#the-template-context) and
[setup-sync.md](setup-sync.md#the-template-context).

## The write of one target

| Side    | Order                                                                                                 |
| ------- | ----------------------------------------------------------------------------------------------------- |
| `repo`  | Render or copy, parse check, merge, dprint, compare, write.                                           |
| `macos` | Render or copy, parse check, merge, compare, write. The Mac side has no formatter and adds no header. |

The repo side has more steps for two targets. See
[repo-sync.md](repo-sync.md#the-pre-commit-config) and
[repo-sync.md](repo-sync.md#overridesgitignore).

- A destination is replaced only after a complete render or copy. The temp file
  is
  next to the destination, so the final `mv` is atomic. A dry run puts the temp
  file
  in `${TMPDIR:-/tmp}`.
- A failure leaves the existing file as it was, and the other files continue.
- A file takes the mode of its source, also when it exists already. An existing
  merge target keeps its own mode, and a new one takes the mode of its source.
  When
  only the mode differs, the sync runs `chmod`.
- A dry run lists the target as `would change`, `would add` or `would seed`,
  with a unified diff when the file exists. A change of the mode alone lists as
  `would change`, with no diff. See [The output of a run](#the-output-of-a-run).
- A file with the bootstrap header at a path where a target needs a folder is
  removed. A file without the header, or a seed or a merge target of a tool of
  the
  run on the repo side, fails that target.

### Generated files and seeds

The modes of a file are in [tools.md](tools.md#the-files-table).

- A generated file is written on each run.
- A seed is written only when the target is missing. After that, the owner of
  the
  target owns it. An existing seed gets no parse check, because the sync does
  not
  write it. A seed has no manifest row and is never in the stale rule of its own
  run.

### The parse check

Each JSON, YAML or TOML target must parse, after its render or copy and before
its
merge and dprint. On the repo side, `pre-commit autoupdate` runs before the
parse
check of `.config/pre-commit-config.yaml`. jq reads JSON, and yq reads YAML and
TOML. The `format` key of its
`[files]` table selects the format, else the extension of the target without
`.tmpl`.

- A `.json` target without a `format` key, with a line that starts with `//`
  after
  spaces, is JSONC. jq cannot read it, so the sync does not examine it. This
  JSONC
  rule does not apply to a merge target.
- An empty JSON target does not parse.
- A target that does not parse keeps its existing file, the other files
  continue,
  and the run does not set `bootstrap_commit`.
- A dry run examines each target too.

## Merge mode

Merge mode is for a target that an app owns, for example
`~/.claude/settings.json`. The source holds only the keys that the tool
enforces.
The sync merges it into the live target, and the other keys stay. The format is
`json`, `yaml` or `toml`: the `format` key, else the extension. A merge target
of
format `text` fails. `sync-engine/merge` has the code of these rules.

The source has no bootstrap header. A YAML or TOML source can have the comment
`# Merged by bootstrap: <path>`. A JSON source has no comment. The sync strips
nothing from a merge source.

On each run, the sync does these steps:

1. It parses the live target.
2. It deletes each `remove` key path, and from the list at each `remove_items`
   `path`, each item that is equal to a listed item.
3. It merges the source over the result. The source wins over `remove`.
   `remove_items` wins over the source: a listed item also leaves the source
   list,
   so it is never added back, also on an absent target.

The merge:

- Tables merge key by key, deeply.
- A source scalar replaces the live value.
- A source value of another kind (table, list, scalar) than the live value
  replaces
  it.
- A list is the live list in its order, then each source item that is not equal
  to
  an item already in it, once. Items that the app or the user added stay.
- Equal means equal as a deep value. The key order of a table does not matter.
- A key that is not in the source stays as the app wrote it.

A `remove` path or a `remove_items` item that is not in the live target is not
an
error. A `remove_items` path whose live value is not a list is also not an
error.
No state records what the sync enforced, so a key that leaves the source stays
in
the live target: put its path in `remove`. When a list item of the source
changes,
put the old item in `remove_items`, or the live target keeps it next to the new
one.

**Output.** JSON is merged and written with jq: a 2-space indent, the key order
of
the live target, new keys at the end of their table, and one trailing newline.
YAML
and TOML are merged and written with yq in their own format, so the comments of
the
live target stay. In TOML, a comment on a table-header line can be lost, and the
spacing before an inline comment can change.

**Absent target.** A target that is missing, 0 bytes or only whitespace (or, for
YAML, only comments) is absent. It gets the source without the `remove_items`
items: JSON through the jq merge, YAML and TOML through yq as rendered, so the
comments of the source stay. A TOML target with only comments is an empty table,
not an absent target.

**Failures.** These keep the live target as it is: a live target that does not
parse, whose top level is not a table, or a YAML target with more than one
document;
a source whose top level is not a table, or a YAML source with more than one
document. An empty source (null or an empty table) is not an error: it enforces
nothing, only the removals act, and an absent target gets an empty table.

**The sides.** A merge target has no header and no manifest row, so the stale
rule
leaves it. When its tool leaves the selection, the file stays.

| Rule     | `repo`                                                                                                            | `macos`                                                     |
| -------- | ----------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| Path     | The path in the repository. A target that is a symlink, or is below a symlinked folder, stays as it is and fails. | The real path. See [setup-sync.md](setup-sync.md#symlinks). |
| Format   | dprint formats the result before the compare.                                                                     | No formatter.                                               |
| Requires | `requires` and `requires_found`.                                                                                  | `requires` only.                                            |

## The header

A file counts as a synced file (it "has the bootstrap header") when one of its
first
3 lines is a comment line that starts with a header text. The comment line has
an
optional indent, then `#`, `//` or `<!--`. A text directly after a `<!--` line
that
has nothing else also counts. A text in the middle of a comment does not count.

| Header text                 | Use                                              |
| --------------------------- | ------------------------------------------------ |
| `Synced from bootstrap: `   | The header of each side now.                     |
| `Synced from repo-config: ` | The old name of `bootstrap`, in older copies.    |
| `Generated by macos-setup`  | The header of the old Mac sync, in older copies. |

Which file gets the header, and where, is in
[tools.md](tools.md#the-header-of-a-source-file).

## The manifest and the stale rule

A manifest lists the generated files of the last clean run. Its first line is
`# Written by bootstrap. Do not edit this file.`. Each other line is
`<tool><TAB><path>`, sorted, with the path relative to the target root, one for
each
generated file of the run. A seed and a merge target get no row. A reader skips
each
`#` line. The sync writes the file only when its content changes, with an atomic
`mv` and the mode `644`. A write counts as `changed` or `new`. On the repo side,
a dry run lists the manifest as `would change` or `would add` when the content
differs. The dry run of the Mac side writes no manifest and
prints nothing for it.

| Side    | Manifest                      | Target root    | Details                                     |
| ------- | ----------------------------- | -------------- | ------------------------------------------- |
| `repo`  | `.config/repo-generated.tsv`  | The repository | [repo-sync.md](repo-sync.md#the-manifest)   |
| `macos` | `<cache>/macos-generated.tsv` | `$HOME`        | [setup-sync.md](setup-sync.md#the-manifest) |

The sync does not trust a manifest. A row with no TAB, an empty tool or path, an
absolute path, or an empty, `.`, `..` or `.git` segment (in any case) is skipped
with
a log line.

**The stale rule** examines each row of the old manifest whose path this run did
not
handle.

- A file that is not there, or that has no bootstrap header, stays with no log
  line.
- A listed path that is the same file as a handled path stays, which covers a
  case
  variant.
- Else the sync removes the file, then its empty parent folders, up to the
  target
  root and never past a symlinked folder. A dry run lists the file as
  `would remove`.
- A failed removal is a failure, and the other rows still run.
- The first sync has no manifest, so the stale rule removes nothing.

| Rule               | `repo` link mode                                                                                                    | `macos` link mode                 |
| ------------------ | ------------------------------------------------------------------------------------------------------------------- | --------------------------------- |
| A link             | A listed path that is a symlink, or is below a symlinked folder, stays. With verbose, it logs `kept <path> (link)`. | The sync removes the real file.   |
| Another repository | A file below a folder with its own `.git` stays. With verbose, it logs `kept <path> (another repository)`.          | Not examined.                     |
| Manifest file      | A manifest that is a symlink, is below a symlinked folder or is not a regular file is not read.                     | The sync reads the file as it is. |

## The values merge

`sync-engine/values` merges the values of each selected tool in this order:

1. `[defaults]` of the tool.
2. `[contributes.<tool>]` of each selected tool, in tool-name order.
3. `[tools.<name>]` of the values file, without `env`.

A contribution merges over `[defaults]`: tables merge key by key, a list appends
its
new items, and any other value replaces the default.

Two contributions that set one key to different values stop the sync, unless
both
values are tables or both are lists. A value in `[tools.<name>]` does not clear
this
conflict, because the sync merges the contributions first. Remove one of the
two.

The tools root of a front-end is `repo/tools/` for `repo-sync` and
`macos/tools/` for
`setup-sync`. A contribution to a tool of that tools root that is not selected
is
ignored. A contribution of a selected tool to a tool that is not in that tools
root
stops the sync.

A value of `[tools.<name>]` depends on the tool:

- A base or dependency tool is a tool that the pack `base` adds, also through a
  nested `default = true` pack (on the repo side, `"pack:security"`: `gitleaks`,
  `grype`, `osv-scanner`, and `_security_base` through `depends`), or a tool in
  the
  depends closure of a selected tool, a tool that a mise `depends` adds
  included. A
  `<tool> = {}` member that the pack does not add does not count. It is a base
  or
  dependency tool also when a `[tools.<name>]` table names it. For such a tool,
  a
  list appends its new items, tables merge key by key, and a value whose type
  differs
  from the merged value stops the sync.
- For any other tool, a value replaces the merged one, and tables merge key by
  key.

The merge runs before the mise config and the file loop, so a failure stops the
sync
before it writes a file of a tool, and `bootstrap_commit` stays. A dry run
merges too.
Across tools, each message logs and the highest code returns. Inside one target
tool,
the merge stops at the first contribution conflict or type change of that tool,
so a
second conflict of that tool shows only after the first is fixed.

When a contributor leaves the selection, its values are gone at the next render.
A
seed gets its contributions only when it is first written. A merge target gets
them
on each run.

The shape of these tables is in
[tools.md](tools.md#the-defaults-questions-and-contributes-tables).

## `bootstrap_commit`

After a sync without a failure, the front-end sets the top-level
`bootstrap_commit`
of its values file to the full commit of its clone. An agent compares it with
`main`
to find what changed since the last sync.

- A line that has the key before the first table is replaced. A line with the
  same
  value causes no write.
- Else the key goes after the header comments, when a blank line ends them, or
  else
  on the first line. A comment just above a key belongs to that key.
- Each other line stays. This is the only write to the values file.
- Through a symlink, the sync writes the real file, so the link stays.
- A real run prints the info line `Set bootstrap_commit to <short commit>`, and
  a dry run prints `Would set bootstrap_commit to <short commit>`. With verbose,
  the lines are `set bootstrap_commit <commit> in <path>` and
  `would set bootstrap_commit <commit> in <path>`.
- A missing `macos.toml` is created. See
  [setup-sync.md](setup-sync.md#macostoml). On the repo side, a missing values
  file
  stops the sync before this step.

## The output of a run

The code of the output is in `common`, so both front-ends print the same lines.
The sample is in
[docs/sync/tools/repo-sync.md](../sync/tools/repo-sync.md#tasks).

**The outcome of a target.** `_se_outcome <kind> <path> [text] [line]` replaces
each per-file log line. It increments the count of the kind and prints the
correct line. Each count is a plain variable, for bash 3.2.

| Kind        | Lines of the log                                      | List                | Dry-run word   | Table row      |
| ----------- | ----------------------------------------------------- | ------------------- | -------------- | -------------- |
| `changed`   | `updated`, `chmod`, `wrote` of a manifest that exists | `✎ changed`, yellow | `would change` | `Changed`      |
| `new`       | `updated`, `wrote` of a target that does not exist    | `✚ new`, green      | `would add`    | `New`          |
| `seeded`    | `updated` of a seed                                   | `❋ seeded`, blue    | `would seed`   | `Seeded`       |
| `removed`   | `removed`                                             | `✖ removed`, red    | `would remove` | `Removed`      |
| `failed`    | Each failure message of a target                      | `✗ failed`, red     | `failed`       | `Failed`       |
| `unchanged` | `unchanged`                                           | None                | None           | `= Unchanged`  |
| `kept`      | `kept seed`                                           | None                | None           | `◦ Kept seeds` |
| `skipped`   | `skip seed`, `skip merge`, `skip file`, `skip env`    | None                | None           | `⊘ Skipped`    |

- **A failed target:** It counts once, and it has one list line, from its first
  failure message: `failed <path>: <text>`. `<path>` is the target, and the text
  does not repeat it. A target that fails with no message gets the text
  `cannot write the target`. A later line for the same target, for example
  `kept existing`, prints only with verbose.
- **No count:** A failure of the mise config of a tool or of the manifest file
  has no count, and its message stays. The line `FAILED: <list>` stays, before
  the table. The other messages of `_se_log` stay too.

**The table and the last line.** `_se_summary <rc>` prints them once, as the
last
output of the run: after the stale rule, the manifest write and
`bootstrap_commit`. Each return after the file loop calls it, also on a failure.
A run that stops before the file loop has no table. Each run prints one table,
so
`--all` prints one table for each member.

- The table always has the rows `Changed`, `New`, `Seeded`, `Removed`,
  `Unchanged`, `Kept seeds`, `Skipped`, `Failed` and `Total`, in that order.
  `Total` is the sum of the rows.
- A row with a zero count is dim. Another row takes the color of its kind in the
  list. `Unchanged`, `Kept seeds`, `Skipped` and `Total` have no color.
- `n` is the sum of `Changed`, `New`, `Seeded` and `Removed`.

| Run              | Last line                                                                   |
| ---------------- | --------------------------------------------------------------------------- |
| Real run, code 0 | `✅ Synced: <n> files changed`, or `✅ Synced: no file changed`             |
| Dry run, code 0  | `✅ Dry run: <n> files would change`, or `✅ Dry run: no file would change` |
| Another code     | A red `❌ The sync failed with code <rc>: <failed count> files failed`      |

The last line uses `1 file` for one file. With no failed file, the line with
another code stops after `<rc>`.

**The style rule.** One rule turns the glyphs and the colors on and off
together:
`FORCE_COLOR` (not `0`) turns them on, `NO_COLOR` turns them off, and otherwise
they need a terminal on stderr and a `TERM` that is not `dumb`. This is the rule
of `style_enabled` of `helpers`. The engine does not call `helpers`, because the
local copy can be stale and the Mac side does not have it. `clone` has its own
copy of the rule, `_se_clone_styled`, because an entry task embeds `clone`
alone.

| Line      | Style on                       | Style off                         |
| --------- | ------------------------------ | --------------------------------- |
| List      | `<glyph> <word> <path>`, color | `<word> <path>`                   |
| Info      | `ℹ️ <text>`, no color          | `[INFO] <text>`                   |
| Last line | `✅ <text>` or `❌ <text>`     | `[OK] <text>` or `[ERROR] <text>` |
| Table     | Box characters                 | ASCII borders (`+`, `-`, `\|`)    |

These lines have no `sync-engine:` or `setup-sync:` prefix. `_se_say` of `clone`
uses `print_info` and `print_error` when they exist. Without them, it prints the
info glyph or the error glyph, or the prefix `sync-engine: ` with style off.

**Verbose.** `--verbose` of `setup:configs` and `updateall:configs` sets
`SYNC_ENGINE_VERBOSE=1`. A user can also set the variable, and only the value
`1`
turns it on. Each entry task consumes the flag and does not pass it to the
engine. Each target then prints its log line in place of its list line, for
example `unchanged  <path>`, `updated    <path>`, `kept seed  <path>`,
`skip env   <path>` or `would update <path>`. The table and the last line stay.
These lines print only with verbose:

- `using bootstrap at <commit>`, `tools:` and, on the repo side,
  `environments:`.
  By default, they are one info line. See
  [repo-sync.md](repo-sync.md#the-entry-tasks) and
  [setup-sync.md](setup-sync.md#the-output).
- `kept <path> (link)` and `kept <path> (another repository)`.
- `done: <ok>/<total> files in sync`, `<ok>/<total> files in sync; existing
  copies of the rest were left untouched` and, on the Mac side, `dry run:
  <ok>/<total> files checked, nothing written`.

The test tasks set the variable, so a case compares the log lines. The cases of
the default output unset it.

## The exit codes

A failure of one file does not stop the other files. A run returns the highest
code
of all its failures. This section is the one place of the codes 2, 3 and 5. Code
4 is
only for the Mac front-end: see
[setup-sync.md](setup-sync.md#stages-and-exit-4).

| Code | Meaning                                                             |
| ---- | ------------------------------------------------------------------- |
| 0    | Done, or nothing to change.                                         |
| 1    | Any other failure, for example a local file system error.           |
| 2    | A usage error, or a config that breaks a rule. See the lists below. |
| 3    | A name that the clone does not have. See the lists below.           |
| 5    | A failure of the network side or of the clone. See the lists below. |

**Code 2, on both sides:**

- A bad flag, and a side other than `repo` or `macos` in a call of a function of
  `sync-engine/mise`.
- A missing `mise`, and a sync without `SYNC_ENGINE_CLONE`.
- A `repo/` of the cache that is not a clone and not empty, and a repository at
  the cache root with another `origin` or no `origin`.
- A `tools.toml` that cannot be read or that breaks a rule, also an old key.
- A `[pack.*]` table that breaks a rule: an old key, a member that is not a
  table
  or is not on one line, another key in a member, a `default` that is not a
  boolean, a pack name that is also a tool folder, and a pack loop.
- A tool folder that is named `zz-repo` or `zz-macos` for its side, or whose
  name
  sorts after that name.
- A tool folder with both `<path>` and `<path>.tmpl`.
- A tool that ships a mise file that the sync also generates.
- A bad `[files]` table: another key, a target that fails, or a bad value.
- A bad `[defaults]`, `[questions]` or `[contributes]` table, and two
  contributions or a user value that conflict.
- A template that cannot render, and a target that does not parse.
- A merge source or a live merge target that cannot merge.
- A `[mise.*]` config that the generator cannot write: another section, a failed
  render, a rendered copy that is not valid TOML or that changes the key paths.
- A bad `[overrides.mise*]`: a path that no selected tool has, another key of
  `[overrides]`, an environment outside `environments`, and the `envs`, the
  `depends`, the `requires`, the `requires_found` or the `files` of a mise tool.

**Code 2, only on the repo side:**

- A missing values file, a values file in the old list format (`tools = [...]`,
  `packs = [...]`), and another key than `env` in a `[packs.<name>]` entry.
- A bad `environments`: an unknown name, or a list without `dev`.
- A bad `env`: a name that is not in `environments`, a list without `dev`, an
  empty list, or an item with a space.
- A name other than `dev` and `ci` in the `default_envs` or the `envs` of a
  `tools.toml`, and a `default_envs` without `dev`, also of a tool that is not
  selected.
- An environment that a mise tool does not allow in its `envs`, also of a
  dependency or through `default_envs`.
- `requires_found`, `requires` or `files` in a `[mise_env.<env>.tools]` entry.
- A `[files]` target that ends in `.tmpl`, with any file in the tool folder.
- Two tools that define the same mise tool.
- A `root_only` tool that a `[tools.*]` entry names in a submodule.
- A repository with neither `main` nor `develop`, and a bad `direct_commit`.
- A bad `[member.*]` table: a name with other characters, two members with one
  name, a folder that is not in the repository, a `folder` of `.`, and a `type`
  that is not correct.
- A member without a `setup:configs:sync` task, when `setup:configs` syncs that
  member with `--all` or `--<name>`.
- A bad `[overrides.gitignore]`, and the old `.config/repo-sync/gitignore`.
- A bad `.config/fnox.toml`.
- A merge target that is a symlink or is below a symlinked folder.
- A `.config/repo-generated.tsv` that is a symlink, is below a symlinked folder
  or is not a regular file.
- `--check` with `--dry-run`.

**Code 2, only on the Mac side:**

- A `mise` that is not at `~/.local/bin/mise`.
- A `MACOS_SETUP_CONFIG_PATH` that is not set or is not a folder.
- A `macos.toml` that breaks a rule, also a `[overrides.mise_env.*]` table.
- A nested run of `updateall:configs`.
- A pack `base` without the tool `macos-setup`.
- A key of a Mac `tools.toml` that is for repo tools only, a `requires_found` of
  a
  `[files]` table, and a `[mise_env.*]`, an `envs` or a `requires_found` of a
  mise
  tool.
- Two selected tools that ship one target, and a tool that ships a file below
  `.config/mise/conf.d/`.

**Code 3:**

- A tool or a pack that is not in the clone, also a member of a pack that is not
  a
  tool folder or a pack, and a tool in a `depends`.
- A contribution of a selected tool to a tool that is not in the clone.

**Code 5:**

- A clone or a fetch that fails.
- A `SYNC_ENGINE_CLONE` that is not usable, on the repo side.
- A `pre-commit autoupdate` that fails, or a missing `pre-commit`, on the repo
  side.

On the Mac side, a `SYNC_ENGINE_CLONE` without `sync-engine/` or `macos/tools/`
is
code 2. A child task that does not find its library in the clone returns code 1.

`repo_sync --check` returns the highest code of its checks, at most 2.

## The release of the engine

The Go engine is a binary of a GitHub release of `virajp/bootstrap`. mise
installs it with its `github:` backend. The workflow
`.github/workflows/release-engine.yml` builds the binary and makes the release.

### The pin

`sync-engine/VERSION` is the pin: the version of the engine that a clone of
`main` uses.

- It has one line, `MAJOR.MINOR.PATCH`, with no `v` and no comment.
- The tag `vX.Y.Z` is the source of the version. The binary gets `X.Y.Z` from
  the build, and the release workflow writes `X.Y.Z` to `VERSION`.
- Nobody edits `VERSION` by hand. A manual change can name a version that has no
  release.

### The rule

The release exists before the pin moves. The workflow commits `VERSION` only as
its last step, after a successful build and release. If a step fails before it,
`VERSION` does not change, so no consumer gets a version without a release.

A tool change that needs a new feature of the engine lands after the pin commit.
Before that commit, consumers run the old engine with the new clone.

### Release a version

**CAUTION:** Do not tag a commit that is not on `main`. The workflow stops, and
the ruleset `all-tags` does not let you move or delete the tag.

1. Land the engine change on `main` with `mise run code:merge <branch>`.
2. Pull `main`.
3. Select the next version `X.Y.Z`.
4. Make a signed tag on the commit of `main`: `git tag -s vX.Y.Z -m "vX.Y.Z"`.
5. Push the tag: `git push origin vX.Y.Z`.
6. Monitor the run of `release-engine` in GitHub Actions.
7. When the run is done, pull `main`. Its last commit is
   `ops: pin sync-engine X.Y.Z`.

The ruleset `all-tags` requires a signed tag. The tag must be
`vMAJOR.MINOR.PATCH`, with no leading zero and no suffix.

### What the workflow does

The workflow starts on a push of a tag `v*.*.*`. Its one job runs on `macos-15`,
an arm64 runner, with the permission `contents: write` only.

1. It examines the tag. The tag must be `vMAJOR.MINOR.PATCH`, and its commit
   must be on `main`. Else the run stops.
2. It installs the Go version of `sync-engine/go.mod`.
3. It builds `./cmd/sync-engine` in `sync-engine/` for `darwin/arm64`, with
   `CGO_ENABLED=0` and `-ldflags "-X main.version=X.Y.Z"`.
4. It runs `sync-engine version`. The output must be `X.Y.Z`.
5. It writes the SHA-256 checksum of the binary.
6. It creates the GitHub release `vX.Y.Z` with two assets:
   `sync-engine-darwin-arm64` and `sync-engine-darwin-arm64.sha256`.
7. It commits `sync-engine/VERSION` = `X.Y.Z` on `main`.

The build is for macOS arm64 only. No Linux build and no Intel build exist.

mise reads release assets, not workflow artifacts. It selects the asset of the
OS and the CPU, and it removes the OS and CPU suffix from a single-file binary.
Thus `sync-engine-darwin-arm64` installs as `sync-engine`. mise also checks the
download against the digest of the asset or the checksum file.

### The pin commit

The ruleset `all-branches` requires signed commits. A `git push` from the runner
is not signed, so the ruleset refuses it. Thus the last step uses the GraphQL
mutation `createCommitOnBranch`, and GitHub signs that commit.

- The step reads the head of `main`, and gives it to the mutation as
  `expectedHeadOid`.
- If `main` moves before the mutation, the mutation fails. The step reads the
  head again and tries again, up to five times.
- If `VERSION` on `main` is already `X.Y.Z`, the step does nothing.
- The commit is from `github-actions[bot]`. A commit of the workflow token
  starts no other workflow.

If the last step fails, the release exists and the pin does not move. Run the
failed job again in GitHub Actions. The release step finds the release and does
not change it, and the last step tries the pin commit again.
