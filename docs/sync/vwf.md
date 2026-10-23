# Adopt bootstrap, and change the tools of a repository, for vwf and stackgen

These instructions are for an AI agent that works in a consumer repository.
This file is the one entry point of `bootstrap` for the vwf and stackgen
plugins. A plugin reads only this file. This file sends the agent to the other
files of `docs/sync/`.

## How a plugin uses this file

1. Clone `bootstrap` into a temporary directory. `--filter=blob:none` gets
   all the commits, for the update check of `index.md`, but only the files of
   the newest commit:

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --filter=blob:none git@github.com:virajp/bootstrap.git "$tmp"
   ```

2. Read `"$tmp"/docs/sync/vwf.md`, this file.
3. Give the instructions of the next section to the agent that reads this
   file.
4. Remove the clone after the agent completes: `rm -rf "$tmp"`.

The plugin never edits `.config/bootstrap.toml` itself. The agent that reads
this file does.

In this file, "the clone" is that temporary directory. Each path in
`docs/sync/` and `tools/` is a path in the clone.

## Instructions

An instruction adopts `bootstrap` in a product, adds tools or removes tools,
or adds or removes Claude Code plugins. In a tools instruction, a name in the
list is a tool folder of
`tools/`, or a pack of `tools/packs.toml`. It has this format:

```text
bootstrap: add tools [<tool>, …]
reason: <why>
```

```text
bootstrap: remove tools [<tool>, …]
reason: <why>
```

One message can hold several instructions. Each instruction has its own
`reason:` line, directly after it. For example:

```text
bootstrap: add tools [flutter]
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

An adopt instruction has one values block for each repository of the
product. A `repository:` line before the block gives the path of the
repository from the root of the base repository: `.` for the base, the path
of a git submodule, or the path of a sibling, for example `../acme-api`. The
block is the draft of the `.config/bootstrap.toml` of that repository, in the
format of `examples/` of the clone. It also holds the tools, the packs and the
plugins, as `[tools.*]`, `[packs.*]` and `[claude.plugins.*]` tables. It has
no `bootstrap_commit`: `repo_sync` writes it:

````text
bootstrap: adopt
reason: <why>
repository: <path>
```toml
<the draft values file of that repository>
```
````

For example, for a base repository with one sibling:

````text
bootstrap: adopt
reason: /vwf:init survey of acme
repository: .
```toml
environments = [ "dev" ]
product_name = "acme"
repo_name = "acme-product"
repo_url = "https://github.com/example/acme-product"

[tools.claude]

[branch.main]
merge_model = "pr"
```
repository: ../acme-api
```toml
environments = [ "dev", "ci" ]
product_name = "acme"
repo_name = "acme-api"
repo_url = "https://github.com/example/acme-api"

[tools.python]

[claude.plugins.user]
"taste-skill@taste-skill" = "Leonxlnx/taste-skill"
```
````

The instruction is the consent. The plugin asked the user before it sent the
instruction. Do not ask the user again to apply it. An adopt instruction is
the consent to start the adoption: `index.md` still asks each of its
questions, and the values of the blocks are the defaults.

## Before each instruction

Do these checks for all the instructions of the message before you change a
file. If a check stops, change nothing, and give its message to the caller.

| Condition                                                                                      | Message                                                                                                   |
| ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| A tools or a plugins instruction, and the repository has no `.config/bootstrap.toml`.          | The repository did not adopt `bootstrap`. Send `bootstrap: adopt` first.                                  |
| A name is not a folder in `tools/` of the clone, and not a pack of `tools/packs.toml`.         | `<name>` is not a tool or a pack of `bootstrap`.                                                          |
| `remove` names a tool that another selected tool needs, through `depends` or a mise `depends`. | `<tool>` cannot be removed alone: `<other tool>` needs it. Remove `<other tool>` in the same instruction. |
| A name is `base`, or a tool or a pack that the `base` pack adds.                               | `<tool>` is in the `base` pack. The sync always adds it.                                                  |

To find whether a selected tool needs a tool, read `tools/<selected>/tools.toml`
in the clone:

- Its `depends` key names the tool.
- A mise `depends` of one of its `[mise.tools.<name>]` entries names a mise
  tool that a `[mise.tools.<name>]` entry of `tools/<tool>/tools.toml` defines.

A tool that the same instruction also removes does not count.

An adopt instruction has these checks too:

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

- `adopt` for a repository that already has `.config/bootstrap.toml`. The
  agent finds whether that repository is up to date, and does only the steps
  that are not.
- `adopt` with a path that does not exist. The agent asks whether to clone or
  to initialize that repository, or to skip it.

- `add` of a tool that already has a `[tools.<tool>]` entry (or a pack that
  already has a `[packs.<pack>]` entry) changes nothing for that name.
- `remove` of a name that has no entry changes nothing for that name.
- `add` of a plugin that the table of that scope already has replaces its
  source.

## Adopt

1. Read `docs/sync/index.md` in the clone, and do its steps in each
   repository of the instruction. Use the clone of this file as the clone of
   its step 1.
2. Use the order of "A product with more than one repository" of `index.md`:
   the deepest submodules first, the siblings, then the base repository.
3. For each repository, use its values block as the defaults: of the tools
   and the packs in step 1, and of the values in step 3. Ask each question of
   `index.md`, also when the block has the answer. Only the user can skip a
   question.
4. For a repository that already has `.config/bootstrap.toml`, start with "A
   repository that already has `.config/bootstrap.toml`" of `index.md`.
5. A repository without a block gets no defaults. Ask the user whether to
   include it.

## Add tools

1. In `.config/bootstrap.toml`, add a `[tools.<tool>]` table for each tool,
   and a `[packs.<pack>]` table for each pack. Write no `env` key: the tool
   then gets its `default_envs`. Keep the tables sorted.
2. Find the tools that each added tool adds, then the tools that these tools
   add in turn. For a pack, start with the tools of its `depends` and its
   `required` in `tools/packs.toml`, and expand each `p:<pack>` in turn. Read
   `tools/<tool>/tools.toml` in the clone. A tool adds:
   - each tool of its `depends` key;
   - for each mise `depends` of one of its `[mise.tools.<name>]` entries, the
     tool whose `tools.toml` has the `[mise.tools.<name>]` entry of that mise
     tool.
3. For each added tool and each tool of step 2, read
   `docs/sync/tools/<tool>.md` in the clone. For each added pack, also read
   `docs/sync/packs/<pack>.md`. Do the items of its "Adapt to the repository"
   and "Migration" sections.
4. Run `mise run setup:configs`.

## Remove tools

Do the checks of "Before each instruction" first. A tool that another selected
tool needs cannot be removed alone.

1. Find the dependency-only tools that leave. Make two sets of tools, with
   the rule of step 2 of "Add tools":
   - The first set is the tools of the `[tools.*]` and `[packs.*]` entries of
     `.config/bootstrap.toml` before the change, the tools of the `base` pack,
     and each tool that they add in turn.
   - The second set is the same, without the removed tools and packs, and
     each tool that they add in turn.

   A tool that is in the first set, is not in the second set, and has no
   `[tools.*]` entry is a dependency-only tool that leaves. Step 3 deletes its
   files too.
2. Remove the `[tools.<tool>]` or `[packs.<pack>]` table of each name from
   `.config/bootstrap.toml`. A dependency-only tool has no table.
3. For each tool that is in the first set of step 1 and not in the second
   set, read "What the tool ships" in `docs/sync/tools/<tool>.md` in the
   clone:
   - Delete each file of kind `Synced`. These files start with the bootstrap
     header.
   - Delete each file of kind `Template` that has `Synced from bootstrap` in
     its first three lines. The path in the row is the output path, with no
     `.tmpl` suffix.
   - Keep each file of kind `Seed` or `Seed, template`. Each of them is a
     seed, and the repository owns it.
   - Do not delete a file of kind `Generated`. Step 4 removes it.
   - Keep a file of kind `Template` without that header in its first three
     lines. Keep a file of any other kind.
4. Run `mise run setup:configs`. The sync removes the generated mise files of
   the tools of step 3.

## Add plugins

1. In `.config/bootstrap.toml`, add each plugin to `[claude.plugins.user]` or
   `[claude.plugins.project]`, the table of the scope:
   `"<name>@<marketplace>" = "<source>"`. Make the table if it is missing.
   Keep the keys sorted.
2. If no selected tool or pack adds the `claude` tool, do "Add tools" for
   `claude` first.
3. Run `mise run setup:ai`. It installs each missing plugin.

## Remove plugins

1. Remove each plugin from the table of the scope in `.config/bootstrap.toml`.
   Keep the table, also when it is empty.
2. Uninstall each plugin at that scope:
   `claude plugin uninstall --scope <scope> <name>@<marketplace>`. `setup:ai`
   does not uninstall a plugin.

## After the change

If `mise run setup:configs` fails, read its exit code in "Step 6: Run the
sync" of `docs/sync/index.md`.

1. Show the changed files to the caller, with `git status --short`, in each
   repository that changed.
2. Show the `reason:` of each instruction to the caller.
3. For a tools or a plugins instruction, do not commit. The caller commits.
   For an adopt instruction, commit only if the user chose it in step 6 of
   `index.md`.
