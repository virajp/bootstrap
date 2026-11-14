# Adopt bootstrap, and change the tools of a repository, for vwf and stackgen

These instructions are for an AI agent that works in a consumer repository.
This file is the one entry point of `bootstrap` for the vwf and stackgen
plugins. A plugin reads only this file. This file sends the agent to the other
files of `docs/sync/`. Each instruction uses the CLI `bootstrap`. Install it
with `mise use -g github:virajp/bootstrap`.

## How a plugin uses this file

1. Clone `bootstrap` into a temporary directory. `--filter=blob:none` gets
   all the commits, for the update check of `index.md`, but only the files of
   the newest commit:

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --filter=blob:none https://github.com/virajp/bootstrap.git "$tmp"
   ```

2. Read `"$tmp"/docs/sync/vwf.md`, this file.
3. Give the instructions of the next section to the agent that reads this
   file.
4. Remove the clone after the agent completes: `rm -rf "$tmp"`.

The plugin never edits `.config/bootstrap.toml` itself. The agent that reads
this file does, with `bootstrap`.

In this file, "the clone" is that temporary directory. Each path in
`docs/sync/` and `repo/` is a path in the clone. The CLI `bootstrap` uses its
own clone in `~/.cache/bootstrap/`.

## Instructions

An instruction adopts `bootstrap` in a product (`init`), adds tools (`add`)
or removes tools (`remove`), or adds or removes Claude Code plugins. In a
tools instruction, a name in the list is a tool folder of `repo/tools/`, or a
pack of `repo/packs.toml`. It has this format:

```text
bootstrap: add [<tool>, …]
reason: <why>
```

```text
bootstrap: remove [<tool>, …]
reason: <why>
```

One message can hold several instructions. Each instruction has its own
`reason:` line, directly after it. For example:

```text
bootstrap: add [flutter]
reason: stackgen pack app-framework/flutter pinned
```

A plugins instruction names the scope, `user` or `project`. In `add`, each
item is `<name>@<marketplace>=<source>`, where `<source>` is the source of the
marketplace for `claude plugin marketplace add`. In `remove`, each item is
`<name>@<marketplace>`:

```text
bootstrap: add user plugins [<name>@<marketplace>=<source>, …]
reason: <why>
```

```text
bootstrap: remove project plugins [<name>@<marketplace>, …]
reason: <why>
```

For example:

```text
bootstrap: add user plugins [taste-skill@taste-skill=Leonxlnx/taste-skill]
reason: stackgen pack design-tool/claude-code pinned
```

An init instruction has one values block for each repository of the
product. A `repository:` line before the block gives the path of the
repository from the root of the base repository: `.` for the base, the path
of a git submodule, or the path of a sibling, for example `../acme-api`. The
block is the draft of the `.config/bootstrap.toml` of that repository, in the
format of `repo/examples/` of the clone. It also holds the tools and the
plugins, as `[tools.*]` and `[tools.claude.plugins.*]` tables. A `[packs.*]`
table of a block names a pack to add with `bootstrap add pack <pack…>`. It has
no `bootstrap_commit`: `bootstrap sync` writes it:

````text
bootstrap: init
reason: <why>
repository: <path>
```toml
<the draft values file of that repository>
```
````

For example, for a base repository with one sibling:

````text
bootstrap: init
reason: /vwf:init survey of acme
repository: .
```toml
environments = [ "dev" ]
product_name = "acme"
repo_name = "acme-product"
repo_url = "https://github.com/example/acme-product"

[tools.claude]

[tools.git.branch.main]
merge_model = "pr"
```
repository: ../acme-api
```toml
environments = [ "dev", "ci" ]
product_name = "acme"
repo_name = "acme-api"
repo_url = "https://github.com/example/acme-api"

[tools.python]

[tools.claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"
```
````

A `[tools.git.branch.main]` or `[tools.git.branch.develop]` table of a block
can also have `direct_commit = true`, to allow a direct commit to that
branch. See "Step 2: Select the branch model" of `index.md`.

The instruction is the consent. The plugin asked the user before it sent the
instruction. Do not ask the user again to apply it. An init instruction is
the consent to start the adoption: `index.md` still asks each of its
questions, and the values of the blocks are the defaults.

## Before each instruction

Do these checks for all the instructions of the message before you change a
file. If a check stops, change nothing, and give its message to the caller.

| Condition                                                                                  | Message                                                                 |
| ------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------- |
| A tools or a plugins instruction, and the repository has no `.config/bootstrap.toml`.      | The repository did not adopt `bootstrap`. Send `bootstrap: init` first. |
| A name is not a folder in `repo/tools/` of the clone, and not a pack of `repo/packs.toml`. | `<name>` is not a tool or a pack of `bootstrap`.                        |
| A name of `add` is `base`, or a required tool of the `base` pack.                          | `<tool>` is in the `base` pack. The sync always adds it.                |
| A name of `remove` is a pack.                                                              | `<name>` is a pack. `bootstrap` has no remove pack: name its tools.     |

`bootstrap add tool <tool…>`, `bootstrap add pack <pack…>` and
`bootstrap remove tool <tool…>` do the other checks. For example,
`bootstrap remove tool <tool…>` refuses a tool that another selected tool needs,
and names that tool. Give each message of a command that stops to the caller.

An init instruction has these checks too:

| Condition                                                                                  | Message                                                                   |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| The base repository is not a git work tree.                                                | `bootstrap` needs a git repository. Run `git init` first.                 |
| The instruction has no `repository:` line, or a block has no `repository:` line before it. | Each values block needs a `repository: <path>` line before it.            |
| Two blocks have the same path.                                                             | `<path>` has more than one values block. Keep one.                        |
| A path exists, but it is not the root of a git work tree.                                  | `<path>` is not the root of a git repository.                             |
| A block is not valid TOML.                                                                 | The values block of `<path>` is not valid TOML: `<error>`.                |
| A `[tools.*]` or `[packs.*]` name of a block is not a tool or a pack of the clone.         | `<name>` in the block of `<path>` is not a tool or a pack of `bootstrap`. |

A plugins instruction has one more check:

| Condition                                                    | Message                                                                  |
| ------------------------------------------------------------ | ------------------------------------------------------------------------ |
| An item of `add` has no `=<source>`, or a plugin has no `@`. | `<item>` is not `<name>@<marketplace>=<source>`.                         |
| The scope is not `user` or `project`.                        | `<scope>` is not a plugin scope of `bootstrap`. Use `user` or `project`. |

These cases are not a stop:

- `init` for a repository that already has `.config/bootstrap.toml`. The
  agent finds whether that repository is up to date, and does only the steps
  that are not.
- `init` with a path that does not exist. The agent asks whether to clone or
  to initialize that repository, or to skip it.
- `add` of a tool that is already selected changes nothing for that name.
  `bootstrap add tool <tool…>` tells so.
- `add` of a plugin that the table of that scope already has replaces its
  source.

## Before a command that writes

`bootstrap` writes no target that has changes that are not committed. It
stops with code 2, names each file, and writes nothing. Give that message to
the caller. Do not commit or discard the changes yourself.

Each command that writes needs `--yes`, because the agent has no terminal.
The instruction is the consent, so give `--yes` after the dry run.

## Init

1. Read `docs/sync/index.md` in the clone, and do its steps in each
   repository of the instruction. Use the clone of this file as the clone of
   its step 1.
2. Use the order of "A product with more than one repository" of `index.md`:
   the deepest submodules first, the siblings, then the base repository.
3. For each repository, use its values block as the defaults: of the tools
   in step 1, of the branch model in step 2, and of the values in step 6. Ask
   each question of `index.md`, also when the block has the
   answer. Only the user can skip a question.
4. In step 5 of `index.md`, give the `[tools.*]` names of the block that are
   not in the `base` pack to `bootstrap add tool <tool…>`, and the
   `[packs.*]` names to `bootstrap add pack <pack…>`. Write the other values
   of the block in its step 6. Do not write a `[packs.*]` table: it is an old
   form.
5. For a repository that already has `.config/bootstrap.toml`, start with "A
   repository that already has `.config/bootstrap.toml`" of `index.md`.
6. A repository without a block gets no defaults. Ask the user whether to
   include it.

## Add

1. For each tool of the instruction, read `docs/sync/tools/<tool>.md` in the
   clone. Do the items of its "Adapt to the repository" and "Migration"
   sections.
2. Get the questions of the tools, and find each answer as step 3 of
   `index.md` tells you:

   ```text
   bootstrap questions <tool…> --json
   ```

3. Run `bootstrap add tool <tool…> --answer <tool>.<key>=<value> --dry-run`.
   `bootstrap add tool <tool…>` adds each tool and each tool of its `depends`,
   asks the questions of the new tools, and syncs. Examine the change list.
4. Run the same command with `--yes` in place of `--dry-run`.

For a pack, read `docs/sync/packs/<pack>.md`, and do its items. Then do steps
3 and 4 with `bootstrap add pack <pack> --answer <tool>.<key>=<value>`. It
writes a `[tools.<name>]` table for each member of the pack that is not
selected, and for each new dependency. The values file keeps no record of the
pack. Each new tool gets its `default_envs`.

## Remove

1. Run `bootstrap remove tool <tool…> --dry-run`. Examine the change list.
2. Run the same command with `--yes` in place of `--dry-run`.

`bootstrap remove tool <tool…>` removes the `[tools.<tool>]` table of each tool,
and each dependency that no other selected tool needs. The sync then removes the
files of these tools that `.config/repo-generated.tsv` lists, and their items in
the shared files. A seed of a removed tool stays:
`bootstrap remove tool <tool…>` lists it. Ask the user whether to delete it.

`bootstrap remove tool <tool…>` refuses a `required` tool and a tool that
another selected tool needs. Give its message to the caller. To remove a tool
that another tool needs, put both tools in one instruction.

`bootstrap` has no remove pack: the values file keeps no record of a pack.
Remove the tools of the pack with `bootstrap remove tool <tool…>`.

## Add plugins

1. If the `claude` tool is not selected, do "Add" for `claude` first.
2. In `.config/bootstrap.toml`, add each plugin to
   `[tools.claude.plugins.user]` or `[tools.claude.plugins.project]`, the
   table of the scope: `"<name>@<marketplace>" = "<source>"`. Make the table
   if it is missing. Keep the keys sorted.
3. Run `mise run setup:ai`. It installs each missing plugin.

## Remove plugins

1. Remove each plugin from the table of the scope in `.config/bootstrap.toml`.
   Keep the table, also when it is empty.
2. Uninstall each plugin at that scope:
   `claude plugin uninstall --scope <scope> <name>@<marketplace>`. `setup:ai`
   does not uninstall a plugin.

## Gradle builds in the Groovy DSL

`bootstrap` uses the Kotlin DSL for Gradle. [`kotlin-dsl.md`](kotlin-dsl.md)
in the clone is a static prompt that migrates a Gradle build from the Groovy
DSL (`*.gradle`) to the Kotlin DSL (`*.gradle.kts`).

After an `init` or an `add` instruction, examine each repository that
changed. Its first section, "When the migration applies", gives the condition:
an executable `gradlew` at the root or in `android/`, and a `*.gradle` file
that git tracks in that folder. If the condition is true, tell the caller, and
do the steps of `kotlin-dsl.md` only when the user agrees. The instruction is
not the consent for this migration.

## After the change

If a command of `bootstrap` or `mise run setup:configs` fails, read its exit
code in "The exit codes" of `docs/sync/index.md`.

1. Show the changed files to the caller, with `git status --short`, in each
   repository that changed.
2. Show the `reason:` of each instruction to the caller.
3. For a tools or a plugins instruction, do not commit. The caller commits.
   For an init instruction, commit only if the user chose it in step 5 of
   `index.md`.
