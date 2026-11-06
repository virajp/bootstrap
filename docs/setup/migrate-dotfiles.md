# Migrate a Mac from the symlink model of `dotfiles/`

These instructions are for an AI agent that works with the owner of a Mac. They
move the Mac from the symlink model of `macos-setup/dotfiles/` to the repo model
of `bootstrap`: the values file `macos.toml` and the Mac sync. Do the steps in
sequence. They apply to each Mac, the Mac of the `macos-setup` owner too.

The migration is a one-time job. After the migration, use
[`index.md`](index.md) to bring the Mac up to date. The tools are in
[`tools.md`](tools.md). The contract of the sync is in
`docs/design/setup-sync.md`.

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
| A scratch folder  | `mktemp -d` gives an empty folder. The doc calls it `<scratch>`.                         |

This guide is for a Mac that has mise and the clone. A new Mac has no
`dotfiles/` model to move. It uses a parked step: the change of the `setup`
script of `macos-setup` (clone `bootstrap` over HTTPS, `gh auth login`, then the
Mac sync). That change has not landed.

## The two forms of the sync command

The symlink model has no task `updateall:configs`. A sync writes that task. The
first real run uses the task form, in [step 10](#step-10-do-the-first-real-run):

```text
MACOS_SETUP_CONFIG_PATH=<clone> mise run updateall:configs
```

Later runs get the variable from the generated mise env, and they do not need it
inline.

The direct form runs the file `sync` of the bootstrap clone:

```text
MACOS_SETUP_CONFIG_PATH=<clone> SYNC_ENGINE_CLONE=<bootstrap> \
  bash <bootstrap>/macos/tools/macos-setup/.config/mise/tasks/updateall/configs/sync
```

Use the direct form only in these two steps, because the task cannot exist there
yet:

- [Step 6](#step-6-compare-in-a-scratch-home): the scratch `HOME` is empty, and
  its first render writes the task.
- [Step 8](#step-8-do-the---dry-run-on-the-real-home): `~/.config/mise` still
  links into `dotfiles/`, and the old file
  `dotfiles/.config/mise/tasks/updateall` is at the path where the task needs a
  folder. [Step 9](#step-9-delete-the-unwanted-files) deletes that file.

Both forms accept `--dry-run` and `--no-bootstrap`.

An older `macos-setup` sync can leave a task with the same name. That task does
not use `bootstrap`. The task form is correct only when this command passes:

```shell
grep -q 'Synced from bootstrap' ~/.config/mise/tasks/updateall/configs/_default
```

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

List these files. Run this list in bash, in `<clone>`:

```text
git ls-files dotfiles | while IFS= read -r f; do
  r=${f#dotfiles/}
  case $r in .config/mise/conf.d/*) continue ;; esac
  found=
  for t in <bootstrap>/macos/tools/*/; do
    if [ -f "$t$r" ] || [ -f "$t$r.tmpl" ]; then found=1; break; fi
  done
  [ -n "$found" ] || echo "$r"
done
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
the model. Use the answers of steps 2 and 3. Move the values of `[vars]` of
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

- `config_path` of `[tools.macos-setup]`: the path of `<clone>`.
- `dock_orientation` of `[tools.macos-defaults]`. The pack `base` always adds
  that tool, so this answer is always necessary. Ask the owner.

Do not commit or push `macos.toml` before step 5 passes. It holds personal
values.

**Pass:** `mise x yq@latest -- yq -p toml -o json . <clone>/macos.toml` prints
JSON with each key above.

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

**1. Render.** Run this in bash. The first run stops with exit 4, because the
init stage changed. That is correct. The second run uses the task, which the
first run wrote:

```text
env -u XDG_CONFIG_HOME -u XDG_CACHE_HOME -u XDG_STATE_HOME -u XDG_DATA_HOME -u MISE_ENV \
  HOME=<scratch> MISE_DATA_DIR="$HOME/.local/share/mise" MISE_CACHE_DIR="$HOME/.cache/mise" \
  MACOS_SETUP_CONFIG_PATH=<clone> SYNC_ENGINE_CLONE=<bootstrap> \
  bash <bootstrap>/macos/tools/macos-setup/.config/mise/tasks/updateall/configs/sync --no-bootstrap

cd <scratch> && env -u XDG_CONFIG_HOME -u XDG_CACHE_HOME -u XDG_STATE_HOME -u XDG_DATA_HOME -u MISE_ENV \
  HOME=<scratch> MISE_DATA_DIR="$HOME/.local/share/mise" MISE_CACHE_DIR="$HOME/.cache/mise" \
  MACOS_SETUP_CONFIG_PATH=<clone> mise run updateall:configs --no-bootstrap
```

The `XDG_*` variables stay out of the command, because otherwise the run writes
to the real cache. **Pass:** The first run exits 4, and the second run exits 0.

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

Run the direct form with `--dry-run`, and read the output with the owner. The
task cannot exist yet: see
[the two forms of the sync command](#the-two-forms-of-the-sync-command).

```text
MACOS_SETUP_CONFIG_PATH=<clone> SYNC_ENGINE_CLONE=<bootstrap> \
  bash <bootstrap>/macos/tools/macos-setup/.config/mise/tasks/updateall/configs/sync --dry-run
```

**Pass:** The owner accepts each change. The run writes nothing. The run exits 1
at this step, because of three `failed` lines: the old file
`dotfiles/.config/mise/tasks/updateall` is at a path where the sync needs a
folder. Step 9 deletes that file. Any other `failed` line stops the migration.

## Step 9: Delete the unwanted files

This step comes after the comparison, because the comparison reads the old
files. After the yes of the owner, in `<clone>`, delete these files and make one
commit. Git keeps them for recovery.

```text
rm dotfiles/.config/mise/conf.d/*.toml
rm dotfiles/.config/mise/tasks/updateall
rm <each obsolete file of step 3>
git add -u dotfiles
git commit -m "ops: delete the old mise fragments and the obsolete dotfiles"
```

- The 19 topic fragments `dotfiles/.config/mise/conf.d/*.toml`: if they stay,
  mise loads the old config and the new config together, and the old
  `[dotfiles]` entries continue to make links.
- The old task `updateall`: a file at the path where the sync needs a folder.
- The obsolete files of step 3.

**Pass:** `test -e dotfiles/.config/mise/tasks/updateall` fails, and
`git status --short -- dotfiles/.config/mise/conf.d` shows nothing.

## Step 10: Do the first real run

The Mac now has mise, the clone and no task. Get the task, and do the dry run,
as
[step 5 of `index.md`](index.md#step-5-do-a-dry-run) says. There, `$tmp` is
`<bootstrap>`. Also render the task when the `grep` of
[the two forms of the sync command](#the-two-forms-of-the-sync-command) fails,
because then the task is an old one.

**Pass:** The dry run exits 0 and has no `failed` line.

After the yes of the owner, do the real run. Use the absolute path of the clone:

```text
MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone> mise run updateall:configs
```

It runs `mise bootstrap`. It stops with exit 4 after the init stage, because the
shells, Homebrew or mise changed. Then open a new terminal, and run the same
command again.

**Pass:** The last run exits 0, `bootstrap_commit` of `macos.toml` is not empty,
and `mise dot status` shows each personal file of step 3.

Then, in `<clone>`, commit `macos.toml` and the files that the sync wrote
through
the links, and push them to the private repo:

```shell
git add macos.toml dotfiles
git commit -m "ops: move the Mac to macos.toml"
git push
```

## Step 11: What the owner does later

The migration deletes nothing else, and it never deletes the `dotfiles/` folder.
The folder holds the personal files and the backup. The owner decides what
happens to the rest of it. Tell the owner this, then stop.
