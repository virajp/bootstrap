# Adopt bootstrap on a Mac, or bring a Mac up to date

These instructions are for an AI agent that works with the owner of a Mac. Do
the steps in sequence to make the Mac get its tools and its config from
`bootstrap` (`https://github.com/virajp/bootstrap.git`). The contract of the
sync is in [`docs/design/setup-sync.md`](../design/setup-sync.md).

This file is a prompt. The owner can use it directly, for example: "Read
`docs/setup/index.md` of bootstrap and do its steps on this Mac." Use it for the
first adoption, and also later, to bring the Mac up to date.

## Two ways to use bootstrap

A Mac uses `bootstrap` in two ways, in this sequence:

1. **Adoption and update, by an AI agent.** The agent selects the tools with the
   owner and writes the values file `macos.toml`.
2. **Sync, each time after that, by a task.** `mise run updateall:configs`
   renders the selected tools into `$HOME`, then runs `mise bootstrap`. It needs
   no agent.

## Files in this folder

| File                                         | Use                                                                          |
| -------------------------------------------- | ---------------------------------------------------------------------------- |
| `index.md`                                   | This file: the first adoption, and the update.                               |
| [`migrate-dotfiles.md`](migrate-dotfiles.md) | The move of a Mac from the old symlink model of `dotfiles/` to `macos.toml`. |
| [`tools.md`](tools.md)                       | The reference of the Mac tools and the packs.                                |

Each rule of the sync has one place in `docs/design/`:

- [`docs/design/setup-sync.md`](../design/setup-sync.md): the keys of
  `macos.toml`, the selection, the stages, exit 4 and the symlink rule.
- [`docs/design/packs.md`](../design/packs.md): the packs.
- [`docs/design/tools.md`](../design/tools.md): the keys of a `tools.toml`.
- [`docs/design/mise.md`](../design/mise.md#the-overrides): `[overrides.mise.*]`
  and `[mise.*]` of `macos.toml`.
- [`docs/design/sync-engine.md`](../design/sync-engine.md#the-exit-codes): the
  exit codes.

## Rules

- Each decision of the owner is one question. Ask one question at a time. Before
  you ask, find each value that you can, and show it as the proposal.
- Do not delete `dotfiles/`, and do not remove a symlink. The sync keeps a link
  and writes through it to its source.
- Do not edit a file that the sync writes. That file has the header
  `Synced from bootstrap`, and the next sync replaces it. Put each change in
  `macos.toml`.
- Do not write a secret in `macos.toml`. It holds names, public keys and paths,
  and the owner commits it.

## Requirements

Examine each item, in this sequence. If an item fails, tell the owner, and stop.

1. The machine is a Mac. Command: `uname -s`. Pass: the output is `Darwin`.
2. mise is the standalone install at `~/.local/bin/mise`. Command:
   `readlink -f "$(command -v mise)"`. Pass: the output ends with
   `/.local/bin/mise`. For any other path, for example a Homebrew copy, the sync
   stops. The install command is `curl https://mise.run | sh`.
3. `gh` has a login. Command: `mise x gh@latest -- gh auth status`. Pass:
   exit 0.
   If it fails, ask the owner to run `mise x gh@latest -- gh auth login`.
4. The owner has a clone of a private `macos-setup` repository. Ask for its path
   (for example `~/Projects/github.com/<owner>/macos-setup`). Command:
   `mise x gh@latest -- gh repo view <owner>/<repo> --json visibility --jq
   .visibility`,
   in the clone. Pass: the output is `PRIVATE`. `macos.toml` holds personal
   values, so a public repository fails this item. To make it private, ask the
   owner, then run
   `mise x gh@latest -- gh repo edit <owner>/<repo> --visibility private
   --accept-visibility-change-consequences`.
5. The machine can reach `https://github.com/virajp/bootstrap.git`. Clone it
   into
   a temporary folder, and keep the folder until the end. Each step below reads
   this clone. Pass: `git clone` exits 0.

   ```bash
   tmp=$(mktemp -d)
   git clone --quiet --filter=blob:none https://github.com/virajp/bootstrap.git "$tmp"
   ```

A new Mac with no mise and no clone cannot pass these items. It uses the parked
change of the `setup` script of `macos-setup`: the script clones `bootstrap`
over
HTTPS, runs `mise x gh@latest -- gh auth login`, then runs the Mac sync. If the
`setup` script of the clone does not have this change, it installs the old
`dotfiles/` model. Do not run it. Tell the owner that this step is parked, and
stop.

## Which path

Examine the Mac. Do the first path that agrees:

| The Mac                                                                                 | Path                                                                                           |
| --------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `~/.config/mise` is a link into a `dotfiles/` folder, and the clone has no `macos.toml` | The old model. Do [`migrate-dotfiles.md`](migrate-dotfiles.md) first. Come back if it says so. |
| The clone has no `macos.toml`                                                           | [First adoption](#first-adoption).                                                             |
| The clone has `macos.toml`                                                              | [Update](#update-to-a-newer-bootstrap_commit).                                                 |

Commands: `readlink ~/.config/mise` and `test -f <path of the
clone>/macos.toml`.

## First adoption

### Step 1: Select the packs and the tools

Read `"$tmp"/macos/packs.toml` and `"$tmp"/docs/setup/tools.md`.

The pack `base` is always added. Do not select it, and do not ask about it. Show
its members to the owner. A tool of `base` cannot go away.

For each other pack, in the order of the file, show the description of each
member. Mark the `default = true` members. Ask these two questions, one at a
time:

1. Do you want this pack? `[packs.<name>]` adds only the default members. It
   adds all or none of them.
2. Which other members of this pack do you want? Each gets a `[tools.<name>]`
   table. If the owner wants only some default members, do not select the pack:
   give each of them a `[tools.<name>]` table.

Then ask whether the owner wants a tool of `tools.md` that is not on the list.
The sync adds the `depends` of each tool. Show the tools that this adds.

Pass: the owner agrees with the full list of the packs and the tools.

### Step 2: Ask the questions of each selected tool

A selected tool is a tool of the list of step 1, a tool of `base`, or a tool in
`depends`, transitively. For each one, read its questions:

```bash
mise x yq@latest -- yq -p toml -o json '.questions // {}' "$tmp/macos/tools/<tool>/tools.toml"
```

For each `[questions.<key>]`, do these items. Leave `config_path` of
`macos-setup` for step 4.

1. If it has `detect`, run that shell command. If the command exits 0 with
   output, show the output as the proposal. Else there is no proposal.
2. Ask its `question`. Show `example` only as the form of the answer. Do not use
   it as an answer.

Pass: each key has an answer that is not empty. A template that needs a missing
answer fails, and the sync stops with code 2.

### Step 3: Write `macos.toml`

Use `"$tmp"/macos/examples/macos.toml` as the model. Show the full draft to the
owner, and ask whether to change a value. Then write the file at the root of the
clone.

| Key                       | Write                                                                                  |
| ------------------------- | -------------------------------------------------------------------------------------- |
| `[packs.<name>]`          | An empty table for each pack of step 1. No `base`.                                     |
| `[tools.<name>]`          | A table for each tool of step 1, and a table with the answers for each tool of step 2. |
| `[overrides.mise.<path>]` | A different default of a tool, if the owner asks for it. It cannot add a key.          |
| `[mise.*]`                | The own mise config of the owner, if any. It goes to `conf.d/zz-macos/mise.toml`.      |

Do not write these. Each of the first four stops the sync:

- An `env` key in a table. The Mac side has no environments.
- `[overrides.mise_env.*]`.
- A key in a `[packs.<name>]` table.
- Any other top-level key.
- A value for `bootstrap_commit`. Each clean real run of the sync writes it.

Command: `mise x yq@latest -- yq -p toml '.' <path of the clone>/macos.toml`.
Pass: exit 0.

### Step 4: Answer the question `config_path`

Propose the path of the clone: `git rev-parse --show-toplevel`, in the clone.
Ask
the question `config_path` of `macos-setup`. Write the answer in `macos.toml`:

```toml
[tools.macos-setup]
config_path = "~/Projects/github.com/<owner>/macos-setup"
```

The sync writes this value to the generated mise env as
`MACOS_SETUP_CONFIG_PATH`. `~` at the start becomes `$HOME`.

Pass: `macos.toml` is in the folder of `config_path`.

### Step 5: Do a dry run

The task `updateall:configs` comes with the sync, so it is not on a Mac before
its first sync. An older `macos-setup` sync can leave a task with the same name,
and that task does not use `bootstrap`. Thus examine the task file, and do not
use the task name. Do the first item that agrees:

1. `grep -q 'Synced from bootstrap'
   ~/.config/mise/tasks/updateall/configs/_default`
   exits 0: the task is ready. Go to the dry run.
2. `test -e ~/.config/mise/tasks/updateall/configs/_default || test -f
   ~/.config/mise/tasks/updateall`
   exits 0: this is an old task. Do not run it. Stop, and do
   [`migrate-dotfiles.md`](migrate-dotfiles.md).
3. Else the task is missing. Render it from the clone, as the first sync does:

```bash
tool="$tmp/macos/tools/macos-setup/.config/mise/tasks/updateall/configs"
task="$HOME/.config/mise/tasks/updateall/configs"
mkdir -p "$task"
printf '{"sync_engine":"%s/sync-engine"}\n' "$tmp" > "$tmp/context.json"
mise x gomplate@latest -- gomplate --left-delim '<%' --right-delim '%>' \
  -c ".=$tmp/context.json" -f "$tool/_default.tmpl" -o "$task/_default"
chmod 755 "$task/_default"
```

Run the dry run. On the first run, give the variable inline. Use the absolute
path of the clone:

```text
MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone> mise run updateall:configs --dry-run
```

Read the diff with the owner. If it replaces a file that the owner wants to
keep, stop, and ask the owner where to put that content (see "Rules").

Pass: exit 0, and the last line has `Dry run:`. A dry run never stops with
exit 4.

### Step 6: Do the real run

Ask the owner whether to run the sync. Then run the same command without
`--dry-run`:

```text
MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone> mise run updateall:configs
```

Later runs get the variable from the generated mise env, so they need only
`mise run updateall:configs`.

On a first run, exit 4 is usual: the init stage changed the shells, Homebrew or
mise. Ask the owner to open a new terminal (the login shell changed), then run
the task again. Other exit codes are in [Exit codes](#exit-codes).

Pass: exit 0, and `bootstrap_commit` of `macos.toml` is not empty. A failed run
does not set it.

### Step 7: Commit `macos.toml`

Show `git status --short` of the clone to the owner. Commit `macos.toml` in the
clone as `ops: sync bootstrap`. Ask the owner before you push.

Pass: `git status --short -- macos.toml` shows nothing.

## Update to a newer `bootstrap_commit`

1. Do the requirements. Then read `bootstrap_commit` of `macos.toml`. If it is
   empty, no sync completed: do step 2 to step 6 of the first adoption.
2. Find what changed. Command, in the bootstrap clone:

   ```text
   git -C "$tmp" log --oneline <commit>..main -- macos sync-engine docs/setup
   ```

   Pass: the command exits 0. If it prints nothing, tell the owner that the
   tools are up to date. Then ask whether to do a sync, and go to item 5.
3. For each new tool, and each new member of a pack, ask whether the owner wants
   it. For each new `[questions.<key>]` of a selected tool, ask it as in step 2
   of the first adoption.
4. Change `macos.toml` where a tool or a pack changed. Show each change to the
   owner first. Pass: `mise x yq@latest -- yq -p toml '.' macos.toml` exits 0.
5. Do a dry run: `mise run updateall:configs --dry-run`. Read the diff with the
   owner. Pass: exit 0.
6. Ask the owner whether to run the sync, then do the real run:
   `mise run updateall:configs`. Pass: exit 0. On exit 4, the owner opens a new
   terminal, and you run the task again.
7. Each sync with a new commit changes `macos.toml`. On a Mac that
   [`migrate-dotfiles.md`](migrate-dotfiles.md) moved, the links stay, so each
   sync also writes through them into `dotfiles/` of the clone. That folder is
   the backup. Show `git status --short` of the clone to the owner. Commit
   `macos.toml`, and `dotfiles` if the clone has that folder, in the clone as
   `ops: sync bootstrap`. Ask the owner before you push. Pass:
   `git status --short -- macos.toml dotfiles` shows nothing.

## Exit codes

| Code | Cause                                          | Action                                                                                  |
| ---- | ---------------------------------------------- | --------------------------------------------------------------------------------------- |
| 0    | Done, or nothing to change.                    | Continue.                                                                               |
| 1    | Another failure, for example a file error.     | Read the lines on stderr.                                                               |
| 2    | A usage error, or a config that breaks a rule. | Read the message. Correct `macos.toml`, or tell the `bootstrap` owner about a bad tool. |
| 3    | A tool or a pack that `macos/` does not have.  | Correct the `[tools.*]` and `[packs.*]` tables.                                         |
| 4    | The init stage changed.                        | Open a new terminal, then run the task again.                                           |
| 5    | The clone or the fetch failed.                 | Examine the network and the access.                                                     |

A failure of one file does not stop the other files. After a failure, the
manifest and `bootstrap_commit` stay as they are, and no `mise bootstrap` runs.
