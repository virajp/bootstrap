# Adopt bootstrap on a Mac, or bring a Mac up to date

These instructions are for an AI agent that works with the owner of a Mac. Do
the steps in sequence to make the Mac get its tools and its config from
`bootstrap` (`https://github.com/virajp/bootstrap.git`). The contract of the
Mac sync is in [`docs/design/mac.md`](../design/mac.md).

This file is a prompt. The owner can use it directly, for example: "Read
`docs/setup/index.md` of bootstrap and do its steps on this Mac." Use it for the
first adoption, and also later, to bring the Mac up to date.

## Two ways to use bootstrap

A Mac uses `bootstrap` in two ways, in this sequence:

1. **Adoption and update, by an AI agent.** The agent selects the tools with the
   owner, and runs `bootstrap mac init`. That command writes the values file
   `macos.toml`, then syncs the Mac.
2. **Sync, each time after that, by a task.** `mise run updateall:configs` runs
   `bootstrap mac sync`. It renders the selected tools into `$HOME`, then runs
   `mise bootstrap`. It needs no agent.

## Files in this folder

| File                                         | Use                                                                          |
| -------------------------------------------- | ---------------------------------------------------------------------------- |
| `index.md`                                   | This file: the first adoption, and the update.                               |
| [`migrate-dotfiles.md`](migrate-dotfiles.md) | The move of a Mac from the old symlink model of `dotfiles/` to `macos.toml`. |
| [`tools.md`](tools.md)                       | The reference of the Mac tools and the packs.                                |

Each rule of the sync has one place in `docs/design/`:

- [`docs/design/mac.md`](../design/mac.md): the keys of `macos.toml`, the
  selection, the stages, exit 4 and the symlink rule.
- [`docs/design/cli.md`](../design/cli.md): the commands of the
  CLI `bootstrap` and the exit codes.
- [`docs/design/packs.md`](../design/packs.md): the packs.
- [`docs/design/tools.md`](../design/tools.md): the keys of a `tools.toml`.
- [`docs/design/mise.md`](../design/mise.md): `[overrides.mise.*]` and
  `[mise.*]` of `macos.toml`.

## The CLI `bootstrap`

Run the CLI through mise. Before the first sync, the Mac does not have it on
`PATH`:

```text
mise x github:virajp/bootstrap@latest -- bootstrap <command>
```

A Mac sync before 2.0.2 wrote `version_prefix = "cli-v"` in
`~/.config/mise/conf.d/mise/mise.toml`. With that line, mise does not find the
release. Remove the line, then run the command. The next Mac sync writes the
file again without the line. If the setting `minimum_release_age` of mise
hides the release, use the version of `cli/VERSION` in place of `latest`.

In this file, `bootstrap` is short for this command. The CLI gets its own clone
of `bootstrap` in `~/.cache/bootstrap/`. This file uses these commands:

| Command                                         | Use                                                                                                                        |
| ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `bootstrap mac questions [tool…] --json`        | Lists the questions of the tools, with the proposal and the answer of each. Without a tool, it uses the selected tools.    |
| `bootstrap mac init`                            | Runs in the clone. Writes `macos.toml` with the members of `base`, or moves an old form, asks the questions, then syncs.   |
| `bootstrap mac add pack <pack> [--exclude <m>]` | Writes a `[tools.<name>]` table for each member of the pack that is not selected, and for each new dependency, then syncs. |
| `bootstrap mac add tool <tool…>`                | Writes a `[tools.<name>]` table for each tool and each new dependency, then syncs.                                         |
| `bootstrap mac remove tool <tool…>`             | Removes the tables of the tools, and of each dependency that no other tool needs, then syncs.                              |
| `bootstrap mac sync`                            | Syncs the Mac. The task `updateall:configs` runs it.                                                                       |

`bootstrap mac init` takes no tool names. The other `mac` commands find
`macos.toml` through
`MACOS_SETUP_REPO_PATH`, or `--repo-path <path>`. Before the first sync, give
that variable inline. The user config file of the CLI has no key for this path.

An agent has no terminal for a prompt of the CLI. Thus give each answer with
`--answer <tool>.<key>=<value>`, also an empty answer with
`--answer <tool>.<key>=`, and give `--yes` after the owner agrees with
the change list. Without `--yes`, a command that writes stops with exit 2
(`changes need --yes`).

In a terminal, `bootstrap mac init` opens the TUI, which is for a person. The
CLI is for an agent: `--no-tui`, `--yes`, `--dry-run` and `--answer` each keep
`bootstrap mac init` on the CLI path. Do not run `bootstrap mac tui`.

## Rules

- Each decision of the owner is one question. Ask one question at a time. Before
  you ask, find each value that you can, and show it as the proposal.
- Do not delete `dotfiles/`, and do not remove a symlink. The sync keeps a link
  and writes through it to its source.
- Do not edit a file that the sync writes. That file has the header
  `Synced from bootstrap`, and the next sync replaces it. Put each change in
  `macos.toml`.
- Let `bootstrap mac init`, `bootstrap mac add` and `bootstrap mac remove`
  write the `[tools.<name>]` tables and the answers. Write by hand only
  `[overrides.mise.*]` and `[mise.*]`. `macos.toml` keeps no record of a pack:
  do not write a `[packs.<name>]` table.
- Do not write a secret in `macos.toml`. It holds names, public keys and paths,
  and the owner commits it.
- A command that writes stops when a file that it changes has changes that are
  not committed. Commit them first, or ask the owner.

## Requirements

Examine each item, in this sequence. If an item fails, tell the owner, and stop.

1. The machine is a Mac. Command: `uname -s`. Pass: the output is `Darwin`.
2. mise is the standalone install at `~/.local/bin/mise`. Command:
   `readlink -f "$(command -v mise)"`. Pass: the output ends with
   `/.local/bin/mise`. For any other path, for example a Homebrew copy, the sync
   stops. The install command is `curl https://mise.run | sh`.
3. `gh` has a login. Command: `mise x gh@latest -- gh auth status`. Pass:
   exit 0. If it fails, ask the owner to run
   `mise x gh@latest -- gh auth login`.
4. The owner has a clone of a private `macos-setup` repository. Ask for its path
   (for example `~/Projects/github.com/<owner>/macos-setup`). The clone must be
   a git repository. Command:
   `mise x gh@latest -- gh repo view <owner>/<repo> --json visibility --jq
   .visibility`,
   in the clone. Pass: the output is `PRIVATE`. `macos.toml` holds personal
   values, so a public repository fails this item. To make it private, ask the
   owner, then run
   `mise x gh@latest -- gh repo edit <owner>/<repo> --visibility private
   --accept-visibility-change-consequences`.
5. The machine can reach `https://github.com/virajp/bootstrap.git`. Clone it
   into a temporary folder, and keep the folder until the end. The steps below
   read the packs and the tools from this clone. Pass: `git clone` exits 0.

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --filter=blob:none https://github.com/virajp/bootstrap.git "$tmp"
   ```

A new Mac with no mise and no clone cannot pass these items. It uses the parked
change of the `setup` script of `macos-setup`: the script installs mise and
`bootstrap`, then runs the Mac sync. If the `setup` script of the clone does not
have this change, it installs the old `dotfiles/` model. Do not run it. Tell the
owner that this step is parked, and stop.

## Which path

Examine the Mac. Do the first path that agrees:

| The Mac                                                                                 | Path                                                                                           |
| --------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `~/.config/mise` is a link into a `dotfiles/` folder, and the clone has no `macos.toml` | The old model. Do [`migrate-dotfiles.md`](migrate-dotfiles.md) first. Come back if it says so. |
| The clone has no `macos.toml`                                                           | [First adoption](#first-adoption).                                                             |
| The clone has `macos.toml`                                                              | [Update](#update-to-a-newer-bootstrap_commit).                                                 |

Commands: `readlink ~/.config/mise` and
`test -f <path of the clone>/macos.toml`.

## First adoption

Do each command of these steps in the clone.

### Step 1: Select the packs and the tools

Read `"$tmp"/macos/packs.toml` and `"$tmp"/docs/setup/tools.md`.

`bootstrap mac init` writes a table for each member of the pack `base`. Show
its members to the owner. A required member cannot go away. Ask whether the
owner wants each member that is not required, for example
`visual-studio-code` and `uv`.

For each other pack, in the order of the file, show its description and the
description of each member. Each member is a default: `mac add pack` adds
each one. Ask these two questions, one at a time:

1. Do you want this pack?
2. Which members of this pack do you not want? Step 5 gives each of them as
   `--exclude <member>`.

Then ask whether the owner wants a tool of `tools.md` that is not in a pack of
the list. The sync adds the `depends` of each tool. Show the tools that this
adds.

Pass: the owner agrees with the full list of the packs and the tools.

### Step 2: Ask the questions

Get the questions of `base`, then of the tools of step 1. In the second
command, name each tool of step 1 and each member of each pack of step 1 that
the owner did not leave out:

```text
MACOS_SETUP_REPO_PATH="$PWD" bootstrap mac questions --json
MACOS_SETUP_REPO_PATH="$PWD" bootstrap mac questions <tool…> --json
```

Each row has `tool`, `key`, `question`, `proposal`, `source`, `answer`,
`group`, `note` and `optional`. An `answer` of `""` means no answer, and the
row is answered. For each row with no answer, do these items:

1. If `source` is `detect`, show `proposal` as the proposal. If `source` is
   `example`, show it only as the form of the answer. Do not use it as an
   answer.
2. For `mise.config_path`, propose the path of the clone:
   `git rev-parse --show-toplevel`. The sync writes this answer to the mise env
   as `MACOS_SETUP_REPO_PATH`.
3. Ask the `question`. Show the `note`, when it is not empty, under the
   question.
4. The owner can leave a row with `optional` = `true` empty. Give it as
   `--answer <tool>.<key>=`.
5. The rows with the same `tool` and a `group` that is not empty need at
   least one answer that is not empty. Give each empty answer of the group as
   `--answer <tool>.<key>=`.

Pass: each other row has an answer that is not empty.

### Step 3: Do a dry run of `bootstrap mac init`

Give one `--answer` for each answer of step 2 of a tool of `base`:

```text
bootstrap mac init --answer <tool>.<key>=<value> … --dry-run
```

A tool in `depends` can add a question that step 2 did not show. Then the
command names it: `the question <tool>.<key> has no answer`. For a group, the
command names the group: `the group <tool>.<group> has no answer`, with the
hint `give at least one with --answer <tool>.<key>=<value>: <key>, <key>`. Ask
that question, or the questions of that group, as in step 2, add each
`--answer`, and run the command again.

Read the change list with the owner. If it replaces a file that the owner wants
to keep, stop, and ask the owner where to put that content (see "Rules").

Pass: exit 0, and the last line has `Dry run:`. A dry run never stops with
exit 4.

### Step 4: Do the real run of `bootstrap mac init`

Ask the owner whether to run the sync. Then run the same command with `--yes`
in place of `--dry-run`:

```text
bootstrap mac init --answer <tool>.<key>=<value> … --yes
```

On a first run, exit 4 is usual: the init stage changed the shells, Homebrew or
mise. Ask the owner to open a new terminal (the login shell changed). Then run
`mise run updateall:configs`. The first run wrote that task and
`MACOS_SETUP_REPO_PATH` into the mise config. Other exit codes are in
[Exit codes](#exit-codes).

Pass: exit 0, and `bootstrap_commit` of `macos.toml` is not empty. A failed run
does not set it.

### Step 5: Add the packs and the tools

Do each command with `--dry-run` first, read the change list with the owner,
then run it again with `--yes` in place of `--dry-run`. Give the answers of
step 2 of the tools that the command adds:

```text
bootstrap mac add pack <pack> --exclude <member> … --answer <tool>.<key>=<value> … --dry-run
bootstrap mac add tool <tool…> --answer <tool>.<key>=<value> … --dry-run
bootstrap mac remove tool <tool…> --dry-run
```

Run `mac add pack` one time for each pack of step 1: `--exclude` needs exactly
one pack name. Run `mac add tool` for the other tools of step 1, and
`mac remove tool` for each member of `base` that the owner does not want.

Pass: each command exits 0.

### Step 6: Write the tables that the CLI does not write

Do this step only when the owner wants a different default of a tool, or an
own mise config. Add these tables to `macos.toml`:

| Table                     | Write                                                                             |
| ------------------------- | --------------------------------------------------------------------------------- |
| `[overrides.mise.<path>]` | A different default of a tool, if the owner asks for it. It cannot add a key.     |
| `[mise.*]`                | The own mise config of the owner, if any. It goes to `conf.d/zz-macos/mise.toml`. |

Do not write a value for `bootstrap_commit`, or a key that
[`docs/design/mac.md`](../design/mac.md) does not give. Show the file to the
owner, then run `mise run updateall:configs`.

Pass: `mise x yq@latest -- yq -p toml '.' macos.toml` exits 0, and the sync
exits 0.

### Step 7: Commit `macos.toml`

Show `git status --short` of the clone to the owner. Commit `macos.toml` in the
clone as `ops: sync bootstrap`. Ask the owner before you push.

Pass: `git status --short -- macos.toml` shows nothing.

## Update to a newer `bootstrap_commit`

Do each command of these items in the clone.

1. Do the requirements. Then read `bootstrap_commit` of `macos.toml`. If it is
   empty, no sync completed: do step 1 to step 7 of the first adoption.
2. Find what changed. Command:

   ```text
   git -C "$tmp" log --oneline <commit>..main -- macos cli docs/setup
   ```

   Pass: the command exits 0. If it prints nothing, tell the owner that the
   tools are up to date. Then ask whether to do a sync, and go to item 5.
3. For each new tool, and each new member of a pack, ask whether the owner
   wants it. `macos.toml` keeps no record of a pack, so a new member of a pack
   comes only with `mac add tool` or `mac add pack`. Put each new tool that the
   owner wants on the list of tools.
4. Get the questions, as in step 2 of the first adoption. Name the new tools of
   item 3 in the second command. Ask each question with no answer. If the
   command stops with `run bootstrap mac init`, `macos.toml` has an old form.
   Then
   go to item 5: the dry run names each question with no answer.
5. Do a dry run: `bootstrap mac init --answer … --dry-run`. It also moves an
   old form, for example a file without `format = 2`, the `[packs.*]` tables,
   or the table of the old Mac entry tool to `[tools.mise]` (see
   `docs/sync/migrate.md`). Read the change list with the owner. Pass: exit 0.
6. Ask the owner whether to run the sync. Then do the real run: the same command
   with `--yes` in place of `--dry-run`. Pass: exit 0. On exit 4, the owner
   opens
   a new terminal, and you run `mise run updateall:configs`. Then add the new
   tools of item 3 with `bootstrap mac add tool <tool…> --answer … --dry-run`,
   and the same command with `--yes`.
7. Each sync with a new commit changes `macos.toml`. On a Mac that
   [`migrate-dotfiles.md`](migrate-dotfiles.md) moved, the links stay, so each
   sync also writes through them into `dotfiles/` of the clone. That folder is
   the backup. Show `git status --short` of the clone to the owner. Commit
   `macos.toml`, and `dotfiles` if the clone has that folder, in the clone as
   `ops: sync bootstrap`. Ask the owner before you push. Pass:
   `git status --short -- macos.toml dotfiles` shows nothing.

## Exit codes

| Code | Cause                                          | Action                                                                                                                               |
| ---- | ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| 0    | Done, or nothing to change.                    | Continue.                                                                                                                            |
| 1    | Another failure, for example a file error.     | Read the lines on stderr.                                                                                                            |
| 2    | A usage error, or a config that breaks a rule. | Read the message. Correct `macos.toml`, or tell the `bootstrap` owner about a bad tool. For `run bootstrap mac init`, do the update. |
| 3    | A tool or a pack that `macos/` does not have.  | Correct the `[tools.*]` tables, or the names of the command.                                                                         |
| 4    | The init stage changed.                        | Open a new terminal, then run `mise run updateall:configs`.                                                                          |
| 5    | The clone or the fetch failed.                 | Examine the network and the access.                                                                                                  |

A failure of one file does not stop the other files. After a failure, the
manifest and `bootstrap_commit` stay as they are, and no `mise bootstrap` runs.
