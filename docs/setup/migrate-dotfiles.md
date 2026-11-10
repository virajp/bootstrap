# Migrate a Mac from the symlink model of `dotfiles/`

These instructions are for an AI agent that works with the owner of a Mac. They
move the Mac from the symlink model of `macos-setup/dotfiles/` to the repo model
of `bootstrap`: the values file `macos.toml` and the Mac sync. Do the steps in
sequence. They apply to each Mac, the Mac of the `macos-setup` owner too.

The migration is a one-time job. After the migration, use
[`index.md`](index.md) to bring the Mac up to date. The tools are in
[`tools.md`](tools.md). The contract of the sync is in
[`docs/design/mac.md`](../design/mac.md).

## Rules

- Ask the owner about each decision. Ask one question at a time.
- Each step gives its command and its pass condition. When a pass condition
  fails, stop and tell the owner. Do not go to the next step.
- Before a command writes to the real `$HOME` or changes the GitHub repo, tell
  the owner the command, and wait for a clear yes.
- Do not remove a symlink. See [step 7](#step-7-remove-no-symlink).
- Do not delete the `dotfiles/` folder. Delete only the files of
  [step 9](#step-9-delete-the-unwanted-files).

## Requirements

| Item              | Check                                                                                    |
| ----------------- | ---------------------------------------------------------------------------------------- |
| mise              | `readlink -f "$(command -v mise)"` ends with `/.local/bin/mise`.                         |
| The owner's clone | The clone of the owner's `macos-setup` repo has `dotfiles/`. The doc calls it `<clone>`. |
| A bootstrap clone | `git clone https://github.com/virajp/bootstrap.git <bootstrap>` passes.                  |
| A git repository  | `git -C <clone> rev-parse --show-toplevel` exits 0.                                      |
| A scratch folder  | `mktemp -d` gives an empty folder. The doc calls it `<scratch>`.                         |

This guide is for a Mac that has mise and the clone. A new Mac has no
`dotfiles/` model to move. It uses a parked step: the change of the `setup`
script of `macos-setup` (install mise and `bootstrap`, then the Mac sync). That
change has not landed.

## The sync command

This guide runs the CLI `bootstrap` through mise, because the symlink model does
not have it:

```text
mise x github:virajp/bootstrap@latest -- bootstrap <command>
```

In this file, `bootstrap` is short for this command. The CLI gets its own clone
of `bootstrap` in `~/.cache/bootstrap/`.

The symlink model has no task `updateall:configs`, and no
`MACOS_SETUP_REPO_PATH` in its mise env. Thus step 6 and step 8 run
`bootstrap mac sync`, and give the variable inline:

```text
MACOS_SETUP_REPO_PATH=<clone> bootstrap mac sync
```

Step 6 runs the binary of the CLI without `mise x`. See
[step 6](#step-6-compare-in-a-scratch-home) for the reason.

[Step 10](#step-10-do-the-first-real-run) runs `bootstrap mac init` in
`<clone>`. That command needs no variable, and it writes the task and the
variable.

## Step 1: Find the changes of the fork

Do this step in `<clone>`. When the remote is missing, add it. Then find the
changes:

```shell
git remote add upstream https://github.com/virajp/macos-setup.git
git fetch upstream
git diff "$(git merge-base HEAD upstream/main)" HEAD -- dotfiles/
```

**Pass:** The diff is the list of the changes of the fork.

- When `origin` is `virajp/macos-setup`, the repo has no fork changes. Start at
  step 2 with all of `dotfiles/`.
- When the owner has no read access to `virajp/macos-setup`, skip this step. The
  diff of the comparison in [step 6](#step-6-compare-in-a-scratch-home) is then
  the list of changes. Do steps 2 and 3 for that list after step 6.

## Step 2: Classify each change

Ask the owner about each change, one at a time. Put each change in one row:

| The change                         | Where it goes in `macos.toml`                                      |
| ---------------------------------- | ------------------------------------------------------------------ |
| A pack or a tool to select         | `[packs.<name>]` or `[tools.<name>]`                               |
| A value of a tool                  | A key of `[tools.<name>]`                                          |
| A change to the generated mise     | `[overrides.mise.*]`. It can change a key, and it cannot add a key |
| An own addition to the mise config | `[mise.*]`. The sync generates it into `conf.d/zz-macos/mise.toml` |
| A file that no tool ships          | [Step 3](#step-3-classify-the-files-that-no-tool-ships)            |

`macos/packs.toml` holds the packs, and [`tools.md`](tools.md) holds the tools.
The pack `base` is always selected.

**Pass:** Each change has one row and the answer of the owner.

## Step 3: Classify the files that no tool ships

The keys of the `[files]` and `[tasks]` tables of each `tools.toml` give the
targets that the tools ship. List the files of `dotfiles/` that are not a
target. Run this list in bash, in `<clone>`:

```text
for t in <bootstrap>/macos/tools/*/tools.toml; do
  mise x yq@latest -- yq -p toml -o yaml \
    '(.files // {} | keys | .[]), (.tasks // {} | keys | .[] | ".config/mise/tasks/" + .)' "$t"
done | LC_ALL=C sort -u > <scratch>/shipped.txt
git ls-files dotfiles | sed 's|^dotfiles/||' | grep -v '^\.config/mise/conf\.d/' |
  LC_ALL=C sort | LC_ALL=C comm -23 - <scratch>/shipped.txt
```

Ignore `.config/mise/tasks/updateall` in the answer: it is the old task, and
[step 9](#step-9-delete-the-unwanted-files) deletes it. Ask the owner about each
other file:

| The answer | Do                                                                                                    |
| ---------- | ----------------------------------------------------------------------------------------------------- |
| Obsolete   | Delete the file in step 9.                                                                            |
| Personal   | The file stays in `dotfiles/`. It gets a `[mise.dotfiles]` entry in `macos.toml`. See the note below. |
| Generic    | Write a proposal for a new tool in `bootstrap/macos/tools/`. The bootstrap owner decides.             |

Do not make a new `files/` folder.

### The note about `dotfiles.root`

A personal file stays in `dotfiles/` and gets a `[mise.dotfiles]` entry in
`macos.toml`. For such an entry, keep the setting `dotfiles.root` in
`[mise.settings]` of `macos.toml`, and point it at the `dotfiles/` folder of
`<clone>`:

```toml
[mise.settings]
dotfiles.root = "<clone>/dotfiles"

[mise.dotfiles]
"~/.config/zsh/02-initialisers.zsh" = { mode = "symlink" }
```

The live model sets `dotfiles.root` in the old `config.toml`, and the sync
replaces that file. The new file does not set it. Without the setting, mise
reads
under `~/.dotfiles`.

The mise docs give these rules for the source path of an entry:

- With no `source`, mise reads the same path relative to `$HOME`, under
  `dotfiles.root`. The entry above reads
  `<clone>/dotfiles/.config/zsh/02-initialisers.zsh`. Use this form.
- With a relative `source`, mise starts at the folder of the config file that
  declares the entry, not at `dotfiles.root`. Here that folder is
  `~/.config/mise/conf.d/zz-macos/`. Thus do not use a relative `source` for a
  personal file.

The new config also does not set `dotfiles.default_mode`. Thus write the `mode`
in each entry.

**Pass:** Each listed file has an answer. Each personal file has an entry.

## Step 4: Write `macos.toml`

Write `macos.toml` at the root of `<clone>`. Use `macos/examples/macos.toml` as
the model. Use the answers of steps 2 and 3. The migration writes this file by
hand, because step 6 needs it before the first `bootstrap mac init`. Move the
values of `[vars]` of
`dotfiles/.config/mise/conf.d/identity.toml` to these keys:

| `[vars]` key               | `macos.toml` key                                             |
| -------------------------- | ------------------------------------------------------------ |
| `name`                     | `name` of `[tools.git]`                                      |
| `github_email`             | `email_github` of `[tools.git]`                              |
| `gitlab_email`             | `email_gitlab` of `[tools.git]`                              |
| `signing_key`              | `signing_key` of `[tools.git]`                               |
| `editor`                   | `editor` of `[tools.gh]`                                     |
| `op_ssh_vault`             | `ssh_vault` of `[tools.1password]`                           |
| `ssh_user`                 | `user` of `[tools.ssh]`                                      |
| `projects_dir`             | `projects_dir` of `[tools.cli]`. The default is `~/Projects` |
| `pnpm_release_age_exclude` | None. The Mac has no pnpm config. A repository sets it       |

`identity.toml` has no value for two more answers:

- `config_path` of `[tools.mise]`: the path of `<clone>`. The sync writes it to
  the mise env as `MACOS_SETUP_REPO_PATH`.
- `dock_orientation` of `[tools.macos-defaults]`. The pack `base` always adds
  that tool, so this answer is always necessary. Ask the owner.

Do not commit or push `macos.toml` before step 5 passes. It holds personal
values.

Then list the questions of the selected tools:

```text
MACOS_SETUP_REPO_PATH=<clone> bootstrap mac questions --json
```

**Pass:** The command exits 0, and each row has an `answer` that is not `null`.
For a row with no answer, ask the owner, and write the answer in `macos.toml`.

## Step 5: Make the repo private

The owner runs the login, because it is interactive. Then run the change, after
the yes of the owner:

```text
mise x gh@latest -- gh auth login
mise x gh@latest -- gh repo edit <owner>/<repo> --visibility private --accept-visibility-change-consequences
mise x gh@latest -- gh repo view <owner>/<repo> --json visibility --jq .visibility
```

**Pass:** The last command prints `PRIVATE`. When GitHub refuses the change, for
example for a fork, stop and show the owner the message. The owner decides how
to get a private repo. Do not commit `macos.toml` to a public repo.

## Step 6: Compare in a scratch `HOME`

This step changes nothing in the real `$HOME`. It can write the
`bootstrap_commit` line of `macos.toml`, and that is correct.

**1. Render.** First get the path of the CLI with the real `HOME`. Run this in
bash:

```text
bin=$(mise x github:virajp/bootstrap@latest -- which bootstrap)
```

Then run this command in bash two times, in the same shell. The first run
stops with exit 4, because the init stage changed. That is correct. The second
run continues with the main stage:

```text
env -u XDG_CONFIG_HOME -u XDG_CACHE_HOME -u XDG_STATE_HOME -u XDG_DATA_HOME -u MISE_ENV \
  HOME=<scratch> MISE_DATA_DIR="$HOME/.local/share/mise" MISE_CACHE_DIR="$HOME/.cache/mise" \
  MACOS_SETUP_REPO_PATH=<clone> \
  "$bin" mac sync --no-bootstrap
```

The `XDG_*` variables stay out of the command, because otherwise the run writes
to the real cache. The command runs the CLI without `mise x`. The first run
writes a mise config into `<scratch>`, and that config sets
`MACOS_SETUP_REPO_PATH` from `config_path`. `mise x` with the scratch `HOME`
loads that config, and expands `~` to `<scratch>`. That value then replaces the
inline value, and the second run stops with exit 2. **Pass:** The first run
exits 4, and the second run exits 0.

**2. Compare the plain files.** Each file outside `.config/mise/conf.d/` is
compared byte for byte, after the removal of the header and of the scratch path.
Run this in bash:

```text
cd <scratch>
find . -type f -not -path './.cache/*' -not -path './.local/*' -not -path './.config/mise/conf.d/*' -print0 |
  while IFS= read -r -d '' f; do
    r=${f#./}
    [ -f "<clone>/dotfiles/$r" ] || { echo "NEW  $r"; continue; }
    sed -e '/Synced from bootstrap:/{N' -e 'd' -e '}' -e "s|<scratch>|$HOME|g" "$f" |
      cmp -s - "<clone>/dotfiles/$r" || echo "DIFF $r"
  done
```

**3. Compare the mise config.** Merge all `conf.d` files of each model, then
compare section by section: `tools`, `bootstrap`, `shell_alias`, `env`,
`dotfiles` and `settings`. Run this in bash, in `<clone>`. A run inside
`<scratch>` fails, because mise does not trust that config:

```text
merged() {
  local f files=()
  while IFS= read -r -d '' f; do
    files+=("$f")
  done < <(find "$1/conf.d" -type f -name '*.toml' -print0 | LC_ALL=C sort -z)
  mise x yq@latest -- yq ea -p toml -o json '. as $f ireduce ({}; . * $f)' "$1/config.toml" "${files[@]}"
}
merged dotfiles/.config/mise > <scratch>/old.json
merged <scratch>/.config/mise > <scratch>/new.json
for s in tools bootstrap shell_alias env dotfiles settings; do
  echo "== $s"
  command diff -u \
    <(mise x jq@latest -- jq -S ".$s" <scratch>/old.json) \
    <(mise x jq@latest -- jq -S ".$s" <scratch>/new.json)
done
```

**4. Compare the identity files.** The live files render from `[vars]` of
`identity.toml`, so compare with the live file in `$HOME`, not with the file in
`dotfiles/`. With the same answers, these must be equal:
`.config/git/identity`, `.config/git/github.config`,
`.config/git/gitlab.config`, `.config/git/allowed_signers`,
`.ssh/config.identity`, `.config/1Password/ssh/agent.toml` and
`.config/gh/config.yml`.

The Mac sync does not write `Library/Preferences/pnpm/config.yaml`, does not set
`PNPM_HOME` and does not install pnpm. A repository selects its own package
manager, so this difference is intended. Without `PNPM_HOME`, the pnpm store
moves from `~/Library/pnpm` to `~/.local/share/pnpm`, and each repository
downloads its packages again. The sync does not delete `~/Library/pnpm`.

**5. List each difference for the owner.** Show one list. Each row is a bug or
an intended change, and the owner decides.

- For an intended change, go on.
- For a change that the owner wants to keep from the old model, change
  `macos.toml` and render again.
- For a bug in `bootstrap`, stop and tell the bootstrap owner.

**Pass:** The owner approved each row of the list.

## Step 7: Remove no symlink

The links into `dotfiles/` stay, for example `~/.config/mise`, `~/.zshrc` and
`~/.ssh/config`. The Mac sync resolves each target through its links and writes
to the real path. Thus the generated files land in `dotfiles/` of the private
repo, and the repo keeps them as a backup.

Only a linked path has this backup. A path that is not a link, for example
`~/.config/fish`, is a plain file and has no backup.

Many paths exist in `dotfiles/` and in `macos/tools/`. The sync replaces them
through the links. This is correct.

**Pass:** `ls -l ~/.config/mise ~/.zshrc ~/.ssh/config` shows each link that the
Mac had at the start.

## Step 8: Do the `--dry-run` on the real `HOME`

Run the sync with `--dry-run`, and read the output with the owner:

```text
MACOS_SETUP_REPO_PATH=<clone> bootstrap mac sync --dry-run
```

**Pass:** The owner accepts each change. The run writes nothing. The run exits 1
at this step, because of a `failed` line for each task below `updateall/`: the
old file `dotfiles/.config/mise/tasks/updateall` is at a path where the sync
needs a folder. Step 9 deletes that file. Any other `failed` line stops the
migration.

## Step 9: Delete the unwanted files

This step comes after the comparison, because the comparison reads the old
files. After the yes of the owner, in `<clone>`, delete these files and make one
commit. Git keeps them for recovery. The commit also holds `macos.toml`, because
`bootstrap mac init` stops when a file that it changes has changes that are not
committed.

```text
rm dotfiles/.config/mise/conf.d/*.toml
rm dotfiles/.config/mise/tasks/updateall
rm <each obsolete file of step 3>
git add -u dotfiles
git add macos.toml
git commit -m "ops: delete the old mise fragments and the obsolete dotfiles"
```

- The 19 topic fragments `dotfiles/.config/mise/conf.d/*.toml`: if they stay,
  mise loads the old config and the new config together, and the old
  `[dotfiles]` entries continue to make links.
- The old task `updateall`: a file at the path where the sync needs a folder.
- The obsolete files of step 3.

**Pass:** `test -e dotfiles/.config/mise/tasks/updateall` fails, and
`git status --short` shows nothing.

## Step 10: Do the first real run

Do these commands in `<clone>`. First do a dry run, and read the change list
with the owner:

```text
bootstrap mac init --dry-run
```

**Pass:** The dry run exits 0, and its last line has `Dry run:`.

After the yes of the owner, do the real run:

```text
bootstrap mac init --yes
```

It writes the task `updateall:configs` and `MACOS_SETUP_REPO_PATH`, and it runs
`mise bootstrap`. It stops with exit 4 after the init stage, because the shells,
Homebrew or mise changed. Then open a new terminal, and run
`mise run updateall:configs` in it. For the other exit codes, see
[`index.md`](index.md#exit-codes).

**Pass:** The last run exits 0, `bootstrap_commit` of `macos.toml` is not empty,
and `mise dot status` shows each personal file of step 3.

Then, in `<clone>`, commit `macos.toml` and the files that the sync wrote
through the links. Ask the owner, then push them to the private repo:

```shell
git add macos.toml dotfiles
git commit -m "ops: move the Mac to macos.toml"
git push
```

## Step 11: What the owner does later

The migration deletes nothing else, and it never deletes the `dotfiles/` folder.
The folder holds the personal files and the backup. The owner decides what
happens to the rest of it. Tell the owner this, then stop.
