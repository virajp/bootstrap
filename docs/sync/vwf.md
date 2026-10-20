# Change the tools of a repository for vwf and stackgen

These instructions are for an AI agent that works in a consumer repository.
This file is the one entry point of `bootstrap` for the vwf and stackgen
plugins. A plugin reads only this file. This file sends the agent to the other
files of `docs/sync/`.

## How a plugin uses this file

1. Make a shallow clone of `bootstrap` into a temporary directory:

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --depth 1 git@github.com:virajp/bootstrap.git "$tmp"
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

An instruction adds tools or removes tools. It has this format:

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

The instruction is the consent. The plugin asked the user before it sent the
instruction. Do not ask the user again to apply it.

## Before each instruction

Do these checks for all the instructions of the message before you change a
file. If a check stops, change nothing, and give its message to the caller.

| Condition                                                                                      | Message                                                                                                   |
| ---------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| The repository has no `.config/bootstrap.toml`.                                                | The repository did not adopt `bootstrap`. The adoption is `docs/sync/index.md`.                           |
| A tool name is not a folder in `tools/` of the clone.                                          | `<tool>` is not a tool of `bootstrap`.                                                                    |
| `remove` names a tool that another selected tool needs, through `depends` or a mise `depends`. | `<tool>` cannot be removed alone: `<other tool>` needs it. Remove `<other tool>` in the same instruction. |
| `remove` names `base`.                                                                         | `base` is necessary in each repository.                                                                   |

To find whether a selected tool needs a tool, read `tools/<selected>/tools.toml`
in the clone:

- Its `depends` key names the tool.
- A mise `depends` of one of its `[mise.tools.<name>]` entries names a mise
  tool that a `[mise.tools.<name>]` entry of `tools/<tool>/tools.toml` defines.

A tool that the same instruction also removes does not count.

These cases are not a stop:

- `add` of a tool that is already in `tools` changes nothing for that tool.
- `remove` of a tool that is not in `tools` changes nothing for that tool.

## Add tools

1. Add each tool name to `tools` in `.config/bootstrap.toml`. Keep the list
   sorted, with no duplicates.
2. Find the tools that each added tool adds, then the tools that these tools
   add in turn. Read `tools/<tool>/tools.toml` in the clone. A tool adds:
   - each tool of its `depends` key;
   - for each mise `depends` of one of its `[mise.tools.<name>]` entries, the
     tool whose `tools.toml` has the `[mise.tools.<name>]` entry of that mise
     tool.
3. For each added tool and each tool of step 2, read `docs/sync/<tool>.md` in
   the clone. Do the items of its "Adapt to the repository" and "Migration"
   sections.
4. Run `mise run setup:configs`.

## Remove tools

Do the checks of "Before each instruction" first. A tool that another selected
tool needs cannot be removed alone.

1. Find the dependency-only tools that leave. Make two sets of tools, with
   the rule of step 2 of "Add tools":
   - The first set is the tools of `tools` in `.config/bootstrap.toml` before
     the change, and each tool that they add in turn.
   - The second set is the same tools without the removed tools, and each
     tool that they add in turn.

   A tool that is in the first set, is not in the second set, and is not in
   `tools` is a dependency-only tool that leaves. Step 3 deletes its files
   too.
2. Remove each tool name from `tools` in `.config/bootstrap.toml`. Do not
   change `tools` for a dependency-only tool. It was never in the list.
3. For each removed tool and each dependency-only tool of step 1, read "What
   the tool ships" in `docs/sync/<tool>.md` in the clone:
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
   the removed tools and of the dependency-only tools of step 1.

## After the change

If `mise run setup:configs` fails, read its exit code in "Step 6: Run the
sync" of `docs/sync/index.md`.

1. Show the changed files to the caller, with `git status --short`.
2. Show the `reason:` of each instruction to the caller.
3. Do not commit. The caller commits.
