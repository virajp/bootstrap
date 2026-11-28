# The Mac tools

This file is the reference of the Mac tools of `bootstrap`. An AI agent reads
it to select the tools of a Mac (see [`index.md`](index.md)).

## How a tool works

Each tool is a folder `macos/tools/<tool>/` with a `tools.toml`. Its `[files]`
table gives each target below `$HOME` and its source in `files/`. An entry with
`root = "backup"` gives a target below the backup clone, the key `backup` of
`macos.toml`, instead. Its `[tasks]`
table gives each task below `~/.config/mise/tasks/` and its source in `tasks/`.
A tool can also have scripts in `scripts/`. The sync runs a script, and never
copies it. The schema of `tools.toml` is in
[`docs/design/tools.md`](../design/tools.md). The rules of the Mac sync are in
[`docs/design/mac.md`](../design/mac.md).

`bootstrap mac sync` renders the selected tools into `$HOME`, and the
`root = "backup"` targets into the backup clone. It installs nothing. The task
`mise run updateall:configs` of the tool `mise` runs that
command, and `updateall` then runs `mise bootstrap`. Only the first setup of a
Mac, which `bootstrap mac migrate` does, runs `mise bootstrap` itself.

`macos.toml`, at the fixed place `~/.config/bootstrap/macos.toml`, selects
the tools:

- `format = 2`: the form of the file.
- `backup`: the backup clone, for example the private `macos-setup` clone.
- `bootstrap_commit`: the commit of the last clean run that changed a file.
  The sync writes it.
- `[tools.<name>]`: a table. It adds the tool. A table for a required tool of
  the pack `base` only gives values.
- `[mise.*]`: the own mise config of the owner. It goes to
  `~/.config/mise/conf.d/zz-macos/mise.toml`.

`macos.toml` keeps no record of a pack. The sync also adds the required members
of the pack `base` and each tool in `depends`, transitively.
`bootstrap mac add tool <tool…>`, `bootstrap mac add pack <pack…>`,
`bootstrap mac remove tool <tool…>`, `bootstrap mac replace tool <old> <new>`
and `bootstrap mac detect --add` change the `[tools.<name>]` tables.
`bootstrap mac detect --add` also adds an installed package that no tool has to
`[mise.bootstrap.packages]` or `[mise.tools]` of `macos.toml`.

A key of `[tools.<name>]` sets a value of that tool. It answers a question of
the tool, or it replaces a key of `[defaults]` of its `tools.toml`.

`[mise.bootstrap.packages]` of a `tools.toml` lists the Homebrew formulae
(`brew:<name>`), the casks (`brew-cask:<name>`) and the Mac App Store apps
(`mas:<adam id>`). A dev tool that mise installs goes in `[mise.tools]`, for
example `usage` of `mise`. mise, not a Brewfile, installs these packages.
`upgrade:brew` upgrades them with `mise bootstrap packages upgrade` and
`mise bootstrap packages apply`. The Brewfile holds only the VS Code
extensions. Each tool with a `mas:` package depends on `mas`. A tool holds one
app or one package. The exceptions are `fonts` (the Nerd Fonts casks), `zsh`
(zsh and zinit) and `mise` (`usage`, `github:virajp/bootstrap` and
`ca-certificates`). The shells load the drop-in files of `~/.config/bash/`,
`~/.config/zsh/` and `~/.config/fish/conf.d/` in the order of their names. A
tool adds its own drop-in, for example `01a-homebrew.zsh` of `homebrew` and
`01b-mise.zsh` of `mise`.

Each task of a Mac tool is a Python uv script with a `.py` name, and mise
removes `.py` from the task name. Each task has
`# MISE tools={uv="latest", python="3.14"}`, and the Mac config of the tool
`mise` sets `task.run_auto_install = false`, so `mise` depends on `python` and
`uv`. A task that uses a function of the library `helpers.py` of the tool
`mise` (the print functions of the Mac tasks, `task_exists` and
`run_task_if_exists`) imports it with this preamble:

```python
sys.dont_write_bytecode = True
sys.path.insert(0, os.path.expanduser("~/.config/mise/tasks/_scripts"))
import helpers  # noqa: E402
```

`func:claude`, `func:code` and `keychain:delete` use no function of the
library, so they do not import it. `updateall:configs` imports no library, so
it runs before the first sync writes `helpers.py`. It runs under `mise run` and
directly, with uv on PATH.

Each section below gives these items for one tool:

- **Packs:** the packs that hold the tool, or none. `(required)` marks a
  member of `base` that each Mac must have.
- **Stage:** `migrate` for a tool with `stage = "init"`, a tool of the init
  stage: the first setup of `bootstrap mac migrate` writes and installs it. A
  tool without this item has no stage of its own. See
  [`docs/design/mac.md`](../design/mac.md).
- **Depends:** the tools that the sync adds with this tool.
- **Files:** each target, and its mode. `copy` is copied on each run.
  `template` is rendered on each run. `seed` is written only when it is
  missing. `merge` adds the items of a merge spec (`files/<name>.merge.toml`)
  to the live file, and the other items of the live file stay. The keys
  `format`, `perm`, `dir_perm`, `when` and `root` of the `[files]` table
  follow. `<backup>/` starts the target of a `root = "backup"` entry: a real
  file in the backup clone, the key `backup` of `macos.toml`.
- **Tasks:** each task file below `~/.config/mise/tasks/`. `library` marks a
  library of `_scripts/`, which is not a task.
- **Scripts:** the scripts in `scripts/` of the tool. The sync runs them, and
  does not copy them.
- **Contributes to `<tool>`:** the keys that the tool adds to the values of
  the tool `<tool>` (`[contributes.<tool>]`), for example `extensions` of
  `visual-studio-code`, or its line of `skill.lines`, a line of the skill of
  `<tool>`. The section has one item for each tool.
- **Defaults:** the keys of `[defaults]`. A key of `[tools.<name>]` in
  `macos.toml` replaces it.
- **Mise sections:** the `[mise.*]` tables of `tools.toml`. They become
  `~/.config/mise/conf.d/<tool>/mise.toml`, which the sync generates.
- **Runtime:** the `runtime` paths of `tools.toml`: the files and folders
  below `$HOME` that the program of the tool writes. They must not go into
  git, so the backup clone does not link a folder that holds one: it links the
  synced items of that folder one by one. See
  [`docs/design/mac.md`](../design/mac.md#the-backup-clone).
- **Skill `<skill>`:** the Claude Code skill that the tool owns, in short: its
  rules, and where the sync writes it. Each `[files]` entry of the skill has
  `when = "claude.skills"`. See
  [`docs/design/tools.md`](../design/tools.md#skills).
- **Question `<key>`:** a value that the owner of the Mac gives, with a link to
  the `tools.toml` that holds its `[questions.<key>]` table. The text of the
  question is only in that table. `bootstrap mac questions` shows it. The
  section has one item for each question.

## The packs

`macos/packs.toml` holds the 11 packs. Each member of a pack is a default:
`bootstrap mac add pack <pack>` writes a table for each one.
`bootstrap mac migrate` adds the pack `base` to a new `macos.toml`. A tool can
be
in more than one pack, or in no pack. Only a member of `base` can be required.

| Pack             | Required members                                                                                                                                                                                                                                                                                                                                                                                                                                    | Other members                                                                                                                                       |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `base`           | `mise`, `homebrew`, `bash`, `zsh`, `fish`, `git`, `git-tools`, `ssh`, `cli`, `fnox`, `node`, `touchid-sudo`, `starship`, `network`, `macos-defaults`, `gh`, `ruby`, `fonts`, `mas`, `age`, `bat`, `coreutils`, `curl`, `doggo`, `eza`, `fzf`, `gping`, `httping`, `ipcalc`, `jq`, `lazygit`, `prettyping`, `ripgrep`, `zoxide`, `diff-so-fancy`, `glab`, `btop`, `pstree`, `tree`, `watch`, `wget`, `yq`, `vorssaint`, `pre-commit`, `python`, `uv` | `visual-studio-code`                                                                                                                                |
| `claude-apps`    | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `claude-desktop`, `claude-cli`, `claude-status`, `codeburn`                                                                                         |
| `essentials`     | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `alfred`, `easyfind`, `the-unarchiver`, `brave-browser`, `github-desktop`, `google-chrome`, `tailscale`, `taskexplorer`, `battery-health`, `setapp` |
| `1password-apps` | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `1password`, `1password-cli`, `1password-safari`                                                                                                    |
| `dev`            | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `lcov`, `orbstack`, `superfile`, `sublime-text`                                                                                                     |
| `terminal`       | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `ghostty`, `warp`                                                                                                                                   |
| `mobile-dev`     | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `xcode`, `developer`, `testflight`, `transporter`, `android-studio`                                                                                 |
| `iwork`          | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `keynote`, `numbers`, `pages`                                                                                                                       |
| `ms-office`      | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `microsoft-excel`, `microsoft-powerpoint`, `microsoft-word`                                                                                         |
| `disk`           | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `daisydisk`, `disk-doctor`, `disk-map`, `duplicate-detective`                                                                                       |
| `media`          | —                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `vlc`, `handbrake`                                                                                                                                  |

## The tools

### `1password`

1Password (the app) as SSH agent and git commit signer. The SSH agent offers
the keys of the vault `ssh_vault`. The cask goes to `/Applications` (`appdir`
of its package), not to `~/Applications` like the other casks, and the git
signer `gpg.ssh.program` is
`/Applications/1Password.app/Contents/MacOS/op-ssh-sign`. Only the detect of
the question `ssh_vault` runs `op`, the command of the tool `1password-cli`.

- **Packs:** `1password-apps`
- **Depends:** `ssh`, `git`
- **Files:**
  - `~/.config/1Password/ssh/agent.toml` — template
- **Contributes to `ssh`:** `identity_agent`
- **Contributes to `git`:** `gpg.ssh.program`, and the line `1password` of
  `skill.lines` (the skill `git-identity`)
- **Mise sections:** `[mise.bootstrap.packages]`
- **Question `ssh_vault`:**
  [`macos/tools/1password/tools.toml`](../../macos/tools/1password/tools.toml)

### `1password-cli`

The 1Password CLI (op).

- **Packs:** `1password-apps`
- **Depends:** `1password`
- **Files:** none
- **Contributes to `fnox`:** the line `1password-cli` of `skill.lines` (the
  skill `secrets`)
- **Mise sections:** `[mise.tools]`

### `1password-safari`

1Password for Safari, from the Mac App Store.

- **Packs:** `1password-apps`
- **Depends:** `mas`, `1password`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `age`

age, the encryption tool.

- **Packs:** `base` (required)
- **Files:** none
- **Contributes to `fnox`:** the line `age` of `skill.lines` (the skill
  `secrets`)
- **Mise sections:** `[mise.bootstrap.packages]`

### `alfred`

Application launcher and productivity software.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `android-studio`

Tools for building Android applications.

- **Packs:** `mobile-dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `bash`

bash from Homebrew, its rc files and its drop-in folder ~/.config/bash/.
`~/.bash_profile` loads `~/.bashrc`, and `~/.bashrc` loads each
`~/.config/bash/*.bash` file in name order, so the drop-ins of other tools
(`01a-homebrew.bash`, `01b-mise.bash`) run after `00-bootstrap.bash`.
`00-bootstrap.bash` sets the XDG folders, `GPG_TTY` (mise cannot set it,
because it needs the live terminal) and `FNOX_CONFIG_DIR`. It adds
`~/.local/bin` at the end of `PATH`, in the shell config and not in
`[env] _.path` of mise. `/etc/shells` is outside `$HOME`, so the line of
Homebrew's bash in it stays a `[mise.dotfiles]` declaration.

- **Packs:** `base` (required)
- **Stage:** `migrate`
- **Depends:** `homebrew`, `mise`
- **Files:**
  - `~/.bash_profile` — copy
  - `~/.bashrc` — copy
  - `~/.config/bash/00-bootstrap.bash` — copy
- **Contributes to `visual-studio-code`:** `extensions`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.dotfiles]`

### `bat`

bat, an alternative to cat, and its alias.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `battery-health`

Battery Health 2, from the Mac App Store.

- **Packs:** `essentials`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `bitwarden`

Bitwarden, the password manager.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `brave-browser`

Brave Browser. Its `[mise.env]` sets `CHROME_EXECUTABLE` to the Brave app in
`~/Applications`, so a tool that reads that variable starts Brave.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.env]`

### `btop`

btop, the resource monitor for the terminal.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `cardhop`

Cardhop, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `claude`

The shared Claude files: settings.json, CLAUDE.md, the npm hook, the skills of
the Mac tools and the upgrade:ai task. `claude-cli` and `claude-desktop` depend
on it, so the sync adds it with either one.

- **Packs:** none
- **Depends:** `jq`. The npm hook reads its JSON input with jq.
- **Files:**
  - `~/.claude/CLAUDE.md` — merge
  - `~/.claude/hooks/npm-normalize.sh` — copy, `perm` `755`
  - `~/.claude/settings.json` — merge
- **Tasks:**
  - `~/.config/mise/tasks/upgrade/ai.py`
- **Defaults:** `settings`, `lists`, `skills`. `skills = true` (the default)
  ships the skills of the Mac tools: `secrets` of `fnox`, and `mac-setup` of
  `mise` and `git-identity` of `git` in the backup clone. Each skill entry has
  `when = "claude.skills"`, so `skills = false` stops all of them.
- **Mise sections:** none
- **Runtime:** `~/.claude/history.jsonl`, `~/.claude/hooks`,
  `~/.claude/projects`, `~/.claude/sessions`, `~/.claude/skills`. Thus the
  backup clone links the synced files of `~/.claude` one by one. Other
  programs also write items into `~/.claude/skills` and `~/.claude/hooks`, so
  the backup clone links each synced item of these two folders one by one,
  and each other item stays a real item in `$HOME`.
- **Settings:** Claude Code rewrites `~/.claude/settings.json`, so the merge
  spec adds only `$schema`, the install keys of `env`
  (`CLAUDE_INSTALL_METHOD`, `DISABLE_AUTOUPDATER`: Homebrew installs Claude
  Code, so it does not update itself), with `claude-status` its status lines
  and hook, and the `PreToolUse` item of the npm hook (matcher `Bash`). It
  removes `deniedMcpServers` and `env.CONTEXT7_RUNNER`, which bootstrap set
  before. A key that the spec does not name stays as Claude Code or the owner
  wrote it, and a key that the user or Claude Code changes keeps its value.
  The own keys of the user are in `settings` of `[tools.claude]`. `lists` maps
  a dotted path, for example `"permissions.allow"`, to the items that the sync
  adds to that list, and an item that the user adds in Claude Code stays.
- **CLAUDE.md:** bootstrap owns the section between the `bootstrap:start` and
  `bootstrap:end` lines at the top of the file, and each sync writes it again.
  The own rules of the user go after the `bootstrap:end` line, and they stay.
- **npm hook:** before each Bash command of Claude Code, the hook changes the
  command when pnpm is on PATH: each `npm ci` becomes
  `pnpm install --frozen-lockfile` (also before `;`, `&`, `|`, `)`, `<` and
  `>`), `npx` becomes `pnpm exec` with its flags kept, and each other `npm`
  becomes `pnpm`. It finds a command at the start of the line, after `;`, `&`
  or `|`, and after `$(`. It changes nothing, and prints `{}`, when pnpm is not
  on PATH, or when the command holds a newline, a quote or a backslash. It
  returns only `updatedInput`, with the whole `tool_input` and only the
  command changed, and no permission decision: the changed command goes
  through the normal permission rules. The change comes before the permission
  check, so an npm-only deny or ask rule, for example `Bash(npm publish:*)`,
  does not match the changed pnpm command. Add a matching pnpm rule, for
  example `Bash(pnpm publish:*)`, in `lists` of `[tools.claude]`.
- **upgrade:ai:** a hidden task that `updateall` runs when it exists. It
  calls the `claude` command of `claude-cli`. It adds the marketplaces
  `anthropics/claude-plugins-official` and `virajp/claude-plugins` when they
  are missing, and updates each marketplace. It updates each plugin of the
  user scope, but not a plugin of the `builtin` marketplace
  (`<name>@builtin`): that plugin ships inside Claude Code, and
  `claude plugin update` cannot find it. Then it removes the unused plugins
  of the user scope (`claude plugin autoremove`).

### `claude-cli`

Claude Code: its output styles, the func:claude task and the cc* aliases.
`func:claude` takes an optional `name`, `-n/--new` and `-m/--model`. It starts
`claude` with a session named after the current folder, or `<folder>-<name>`
with the argument `<name>`, and resumes that session with `--resume` unless
`--new` is given. It adds `--remote-control`, `--dangerously-skip-permissions`
and `--allow-dangerously-skip-permissions`. `cc` uses the default model, and
`ccf`, `cch`, `cco` and `ccs` give `--model` with Fable, Haiku, Opus and Sonnet
(each with `[1m]`).

- **Packs:** `claude-apps`
- **Depends:** `claude`, `cli`
- **Files:**
  - `~/.claude/output-styles/ste100.md` — copy
  - `~/.claude/output-styles/terse.md` — copy
- **Tasks:**
  - `~/.config/mise/tasks/func/claude.py`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `claude-desktop`

Claude Desktop.

- **Packs:** `claude-apps`
- **Depends:** `claude`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `claude-status`

The status line of Claude Code (from the virajp/tap tap). When this tool is
selected, the merge spec of `~/.claude/settings.json` of `claude` adds
`statusLine`, `subagentStatusLine` and the `PostToolUse` hook of
`claude-status`. The tool declares the tap `virajp/tap` in
`[mise.bootstrap.brew.taps]`, and its package is
`brew:virajp/tap/claude-status`.

- **Packs:** `claude-apps`
- **Depends:** `claude-cli`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `cli`

The projects folder, the generic aliases and the func:*, cleanupDS and
keychain:delete tasks. Each command-line package is a tool of its own.
`projects_dir` (default `~/Projects`) sets `PROJECTS_DIR`, with `~` expanded,
because a value of `[mise.env]` does not expand it. `projects_dir` also sets the
alias `cdgh`. The tools `git` and `mise` also read `projects_dir`.
`CLOUD_PATH` is the iCloud Drive folder.

- **Packs:** `base` (required)
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/cleanupDS.py`
  - `~/.config/mise/tasks/func/backup.py`
  - `~/.config/mise/tasks/func/code.py`
  - `~/.config/mise/tasks/func/listStartupItems.py`
  - `~/.config/mise/tasks/func/zip.py`
  - `~/.config/mise/tasks/keychain/delete.py`
- **Defaults:** `projects_dir`
- **Mise sections:** `[mise.env]`, `[mise.shell_alias]`

### `codeburn`

codeburn, the usage and cost report of the AI tools, and its aistats alias.

- **Packs:** `claude-apps`
- **Depends:** `claude-cli`, `node`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.shell_alias]`

### `coreutils`

coreutils, the GNU core utilities.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `curl`

curl, which transfers data with URLs.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `daisydisk`

DaisyDisk, from the Mac App Store.

- **Packs:** `disk`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `dasel`

dasel, the query tool for JSON, YAML and TOML, and its fish completion. The
fish file loads the completion only when `dasel` is on PATH.

- **Packs:** none
- **Files:**
  - `~/.config/fish/conf.d/53-dasel.fish` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `developer`

Developer, from the Mac App Store.

- **Packs:** `mobile-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `diff-so-fancy`

diff-so-fancy, the git diff formatter. The tool `git` depends on it, and its
alias `diff` runs `diff-so-fancy`. The git config sets no pager.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `disk-doctor`

Disk Doctor, from the Mac App Store.

- **Packs:** `disk`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `disk-map`

Disk Map, from the Mac App Store.

- **Packs:** `disk`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `doggo`

doggo, the DNS lookup utility, and the dig alias.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `dprint`

dprint, the code formatter, and its global config.

- **Packs:** none
- **Files:**
  - `~/.config/dprint/dprint.jsonc` — merge
- **Contributes to `visual-studio-code`:** `extensions`
- **Defaults:** `settings`, `blocks`, `excludes`, `plugins`
- **Mise sections:** `[mise.tools]`
- **Plugins:** The global config formats the base formats only: TOML, JSON,
  Markdown, YAML, shell and Python. A repository gets its own
  `dprint.jsonc`, which extends no other config. The ruff plugin
  (`npm:@dprint/ruff`, the block `ruff`) formats each `.py` file, also each
  Python task below `~/.config/mise/tasks/`. The shuck block formats the
  other task files, and skips `*.tmpl` and `*.py`. A key, an exclude or a
  plugin that the user adds stays. The sync finds a plugin by its name, so
  it keeps the version in the file. When the file exists and `dprint` is on
  PATH, `updateall` runs `dprint config update` on the file, and that step
  updates the versions.

### `duplicate-detective`

Duplicate Detective, from the Mac App Store.

- **Packs:** `disk`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `easyfind`

Find files, folders, or contents in any file.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `eza`

eza, a modern replacement for ls, and the ls and tree aliases.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `final-cut-pro`

Final Cut Pro, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `fish`

fish from Homebrew, its config, drop-in folder conf.d/ and functions, and the
login shell. fish loads `conf.d/` in name order: `01-env` and `02-path` set
the XDG folders, `HOMEBREW_PREFIX` and the PATH before mise activates at
`06-mise.fish`, so the tools of mise win on PATH. `02-path.fish` puts
`~/.local/bin` on `PATH`, not `[env] _.path` of mise. mise puts `_.path`
before its own tool paths, so with `_.path` the standalone copies in
`~/.local/bin` (`uv` and `uvx`) come before the tools of mise. `02-path.fish`
uses `fish_add_path --global`. A universal path stays in `fish_variables`
after its line is removed, and the global scope is built again at each start.
`/etc/shells` is outside `$HOME`, so the line of Homebrew's fish is a
`[mise.dotfiles]` entry, as the lines of `bash` and `zsh`. The login shell is
`login_shell` of `[mise.bootstrap.user]`. `mise bootstrap` applies both.

- **Packs:** `base` (required)
- **Stage:** `migrate`
- **Depends:** `homebrew`, `mise`
- **Files:**
  - `~/.config/fish/conf.d/01-env.fish` — copy
  - `~/.config/fish/conf.d/02-path.fish` — copy
  - `~/.config/fish/conf.d/fish_frozen_key_bindings.fish` — seed
  - `~/.config/fish/conf.d/fish_frozen_theme.fish` — seed
  - `~/.config/fish/config.fish` — copy
  - `~/.config/fish/functions/fish_title.fish` — seed
- **Contributes to `visual-studio-code`:** `extensions`
- **Defaults:** `login_shell`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.bootstrap.user]`,
  `[mise.dotfiles]`
- **Runtime:** `~/.config/fish/fish_variables`. Thus the backup clone links
  `conf.d/`, `functions/` and `config.fish`, not `~/.config/fish`.

### `fnox`

fnox: secrets from the macOS Keychain, with a profile per consumer, and the
Claude skill `secrets`. fnox has one provider, `global`: it reads each secret
from the macOS Keychain (service `fnox`, prefix `global/`). `env = false`
exports no secret to the shell: only the command of `fnox exec -P <profile>`
gets it. `if_missing = "error"` stops a command when a secret is missing. The
profile `brew` gives `GITHUB_API_TOKEN` and `HOMEBREW_GITHUB_API_TOKEN` to the
command, from the Keychain item `GITHUB_API_TOKEN`. The second profile is
`fnox`. It gives `AGE_KEY`.

- **Packs:** `base` (required)
- **Files:**
  - `~/.config/fnox/config.toml` — copy
  - `~/.claude/skills/secrets/SKILL.md` — template, `when = "claude.skills"`
  - `~/.claude/skills/secrets/references/conventions.md` — copy,
    `when = "claude.skills"`
  - `~/.claude/skills/secrets/references/local.md` — seed,
    `when = "claude.skills"`
- **Defaults:** `skill.lines`, an empty table. `1password-cli` and `age` add
  their lines.
- **Mise sections:** `[mise.bootstrap.packages]`
- **Skill `secrets`:** secrets come from fnox (the macOS Keychain) or op. Give
  a secret to a command with `fnox exec` or `op run`. Never print a secret,
  never write it to a file, never put it in the chat. The skill ships only
  when the `claude` tool is selected and its value `skills` is `true`.

### `fonts`

The Nerd Fonts casks (Fira Code, Hack, Iosevka, Iosevka Term, Meslo LG).

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `fzf`

fzf, the command-line fuzzy finder.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `gcloud`

The Google Cloud CLI (gcloud) and its env. `CLOUDSDK_PYTHON` names the Python
3.14 of the `python` tool. Without it, gcloud runs the system `python3` (3.9 on
macOS) and fails. The path names the minor version 3.14, not `latest`, because
`latest` moves to each newer Python, and the gcloud virtual environment in
`~/.config/gcloud/virtenv` then runs without its packages.

- **Packs:** none
- **Depends:** `python`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.env]`

### `gh`

the GitHub CLI (gh) and its config: ssh as git protocol, the editor, the
aliases. `gh config set` rewrites `config.yml`, so the file is a merge target:
the spec adds only `version`, `git_protocol`, `editor`, `pager` and the alias
`co`, and gh owns each other key.

- **Packs:** `base` (required)
- **Depends:** `git`
- **Files:**
  - `~/.config/gh/config.yml` — merge
- **Contributes to `git`:** the line `gh` of `skill.lines` (the skill
  `git-identity`)
- **Mise sections:** `[mise.bootstrap.packages]`
- **Runtime:** `~/.config/gh/hosts.yml`. Thus the backup clone links
  `config.yml`, not `~/.config/gh`.
- **Question `editor`:**
  [`macos/tools/gh/tools.toml`](../../macos/tools/gh/tools.toml)

### `ghostty`

Ghostty, the terminal emulator, and its config. The config uses the font
MesloLGS Nerd Font, so the tool depends on `fonts`.

- **Packs:** `terminal`
- **Depends:** `fonts`
- **Files:**
  - `~/.config/ghostty/config` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `git`

git, its global config, the commit identity per forge (GitHub, GitLab), the
git aliases and the Claude skill `git-identity` of the backup clone. `config`
includes `identity`. `identity` holds `[user]`, and includes `github.config` or
`gitlab.config` by the URL of a remote (`hasconfig:remote.*.url:`) and by the
folder of the repository (`gitdir:`): a GitHub remote or a folder below
`<projects_dir>/github.com/` gets `github.config`, and a GitLab remote or a
folder below `<projects_dir>/gitlab.com/` gets `gitlab.config`. `projects_dir`
is a value of the `cli` tool, so `git` depends on `cli`. The config signs each
commit and each annotated tag with SSH, and the `program` line of
`[gpg "ssh"]` renders only when `1password` gives `gpg.ssh.program`.
`push.default` is `current`, so the alias `gp` and the git alias `p` push only
the current branch.

- **Packs:** `base` (required)
- **Depends:** `cli`, `diff-so-fancy`
- **Files:**
  - `~/.config/git/allowed_signers` — template
  - `~/.config/git/config` — template
  - `~/.config/git/github.config` — template
  - `~/.config/git/gitlab.config` — template
  - `~/.config/git/identity` — template
  - `~/.config/git/ignore` — copy
  - `~/.github/git-commit-guidelines.md` — copy
  - `<backup>/.claude/skills/git-identity/SKILL.md` — template,
    `root = "backup"`, `when = "claude.skills"`
  - `<backup>/.claude/skills/git-identity/references/conventions.md` — copy,
    `root = "backup"`, `when = "claude.skills"`
  - `<backup>/.claude/skills/git-identity/references/local.md` — seed,
    `root = "backup"`, `when = "claude.skills"`
- **Defaults:** `core_editor`, and `skill.lines`, an empty table. `1password`,
  `ssh`, `gh` and `glab` add their lines.
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`
- **Skill `git-identity`:** the identity for each forge comes from `includeIf`,
  never a local `user.email`. The 1Password SSH signer signs the commits. If a
  signature fails, tell the user to unlock 1Password, and never turn off
  signing. Use `gh auth status` and `glab auth status`, and never print a
  token. `<backup>` is the backup clone (`macos-setup`), the key `backup` of
  `macos.toml`. Without `backup`, the sync skips the skill.
- **Question `name`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `email_github`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `email_gitlab`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `signing_key`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)

`name` and `signing_key` need an answer. `email_github` and `email_gitlab` are
in the group `email`: give at least one email. The note of `email_github` tells
the owner that the SSH key that signs the commits is the same for GitHub and
GitLab. With both emails empty, the sync stops with code 2 and the message
`the group git.email has no answer in macos.toml`. An empty email changes these
files:

- `~/.config/git/identity`: `email` is `email_github` when it is not empty,
  else `email_gitlab`.
- `~/.config/git/github.config`: no `email` line when `email_github` is empty.
- `~/.config/git/gitlab.config`: no `email` line when `email_gitlab` is empty.
- `~/.config/git/allowed_signers`: no signer line for an empty email.

### `git-tools`

git-tools, more git commands.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `github-desktop`

GitHub Desktop.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `glab`

glab, the GitLab CLI.

- **Packs:** `base` (required)
- **Files:** none
- **Contributes to `git`:** the line `glab` of `skill.lines` (the skill
  `git-identity`)
- **Mise sections:** `[mise.bootstrap.packages]`

### `google-chrome`

Google Chrome.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `gping`

gping, ping with a graph.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `handbrake`

Open-source video transcoder.

- **Packs:** `media`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `homebrew`

Homebrew: its environment, the brew alias, the Brewfile, its upgrade task and
its shellenv in bash, zsh and fish. `HOMEBREW_CASK_OPTS` gives `brew` the cask
folder of mise, `~/Applications`. The alias `brew` runs
`fnox -c $HOME/.config/fnox/config.toml exec -P brew -- brew`, so `brew` gets
the GitHub token of the profile `brew` of fnox. Thus the tool depends on
`fnox`. The shellenv files (`01a-`, `05-`) run before the mise drop-ins
(`01b-`, `06-`), so the tools of mise win on PATH. The Brewfile holds only the
VS Code extensions of `visual-studio-code`, with no line when that tool is not
selected.

`upgrade:brew` updates Homebrew and runs `upgrade:tailscale` when that task
exists. Then it runs `mise bootstrap packages prune`, which removes each
package that no tool and no `[mise.bootstrap.packages]` of `macos.toml`
declares, and it upgrades and applies the declared packages. `updateall` runs
`upgrade:brew` under `fnox exec --profile brew` when fnox is on `PATH`, else
without the token, with a warning. Its other `mise bootstrap` step gets no
token from fnox. `brew:casks` runs by hand only: it installs the Brewfile, and
removes each cask and VS Code extension that the Brewfile does not hold.

- **Packs:** `base` (required)
- **Depends:** `fnox`
- **Files:**
  - `~/.config/bash/01a-homebrew.bash` — copy
  - `~/.config/brewfile` — template
  - `~/.config/fish/conf.d/05-homebrew.fish` — copy
  - `~/.config/zsh/01a-homebrew.zsh` — copy
- **Tasks:**
  - `~/.config/mise/tasks/brew/casks.py`
  - `~/.config/mise/tasks/upgrade/brew.py`
- **Contributes to `mise`:** the line `homebrew` of `skill.lines` (the skill
  `mac-setup`)
- **Mise sections:** `[mise.env]`, `[mise.shell_alias]`

### `httping`

httping, which pings a web server over HTTP, and the hping alias.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `imagemagick`

ImageMagick, the image manipulation tool.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `ipcalc`

ipcalc, the IP address calculator.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `jq`

jq, the command-line JSON processor.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `keynote`

Keynote, from the Mac App Store.

- **Packs:** `iwork`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `kindle`

Kindle, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `lazygit`

lazygit, a simple terminal UI for git.

- **Packs:** `base` (required)
- **Depends:** `git`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `lcov`

lcov, the code coverage tool.

- **Packs:** `dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `macos-defaults`

macOS settings: Finder, Dock, Mission Control, window management, menu bar clock
and trackpad. `dock_orientation` is `bottom`, `left` or `right`. Its detect
proposes the current position of the Dock, else `bottom`. Another value stops
the render, and the sync stops. A `defaults` value is not a template and cannot
hold `$HOME`, so the hook `post-defaults` writes `NewWindowTargetPath` of
Finder with `$HOME`. It also runs `chflags nohidden ~/Library`, and restarts
`cfprefsd`, Finder, the Dock, `SystemUIServer` and Control Center, so they read
the new values. The `pmset` and `nvram` settings are in the task
`upgrade:power` of `mise`.

- **Packs:** `base` (required)
- **Files:** none
- **Contributes to `mise`:** the line `macos-defaults` of `skill.lines` (the
  skill `mac-setup`)
- **Mise sections:** `[mise.bootstrap.macos]`, `[mise.bootstrap.hooks]`
- **Question `dock_orientation`:**
  [`macos/tools/macos-defaults/tools.toml`](../../macos/tools/macos-defaults/tools.toml)

### `magnet`

Magnet, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `mas`

Mac App Store command-line interface.

- **Packs:** `base` (required)
- **Files:** none
- **Contributes to `mise`:** the line `mas` of `skill.lines` (the skill
  `mac-setup`)
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-excel`

Microsoft Excel, from the Mac App Store.

- **Packs:** `ms-office`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-powerpoint`

Microsoft PowerPoint, from the Mac App Store.

- **Packs:** `ms-office`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-word`

Microsoft Word, from the Mac App Store.

- **Packs:** `ms-office`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `mise`

mise: its user-level config, its activation in bash, zsh and fish, ~/.zprofile,
the entry tasks of the Mac sync and the Claude skill `mac-setup` of the backup
clone. The activation files read `projects_dir` of the `cli` tool, so the tool
depends on `cli`. The entry tasks are Python uv scripts, so the tool also
depends on `python` and `uv`: both are then in the init stage, and the first
setup installs them before its exit 4. `~/.zprofile` puts `~/.local/bin`, the
folder of mise, on PATH before the activation files run. Each activation file
exports `MISE_ENV=dev`. An interactive shell gets the full `mise activate`, and
a non-interactive shell gets `mise activate --shims`.

The activation of mise runs after `brew shellenv` and the `PATH` exports:
`01b-mise` sorts after `01a-homebrew` in bash and zsh, and `06-mise.fish`
after `02-path.fish` and `05-homebrew.fish` in fish. mise prepends its tool
paths to `PATH`, so its tools win, and a later change of `PATH` takes that
place back. The fish file sets `MISE_FISH_AUTO_ACTIVATE` to `0`, because the
Homebrew formula of mise ships `vendor_conf.d/mise-activate.fish`, which sorts
after it and runs a full `mise activate` again.
`MISE_BREW_CASK_OPT_APPDIR` of `~/.config/mise/config.toml` puts each cask that
mise installs into `~/Applications`. The place of `macos.toml` is fixed, at
`~/.config/bootstrap/macos.toml`, so the tool has no question `config_path`:
`mac migrate` moves an old answer to the key `backup`. See
[mac.md](../design/mac.md#the-place-of-macostoml).

- **Packs:** `base` (required)
- **Depends:** `cli`, `python`, `uv`
- **Files:**
  - `~/.config/bash/01b-mise.bash` — template
  - `~/.config/fish/conf.d/06-mise.fish` — template
  - `~/.config/mise/config.toml` — copy
  - `~/.config/mise/miserc.toml` — copy
  - `~/.config/zsh/01b-mise.zsh` — template
  - `~/.zprofile` — copy
  - `<backup>/.claude/skills/mac-setup/SKILL.md` — template,
    `root = "backup"`, `when = "claude.skills"`
  - `<backup>/.claude/skills/mac-setup/references/conventions.md` — copy,
    `root = "backup"`, `when = "claude.skills"`
  - `<backup>/.claude/skills/mac-setup/references/local.md` — seed,
    `root = "backup"`, `when = "claude.skills"`
- **Tasks:**
  - `~/.config/mise/tasks/_scripts/helpers.py` — library
  - `~/.config/mise/tasks/updateall/_default.py`
  - `~/.config/mise/tasks/updateall/configs/_default.py`
  - `~/.config/mise/tasks/upgrade/macos/download.py`
  - `~/.config/mise/tasks/upgrade/macos/upgrade.py`
  - `~/.config/mise/tasks/upgrade/power.py`
- **Contributes to `visual-studio-code`:** `extensions`
- **Defaults:** `skill.lines`, an empty table. `homebrew`, `mas`, `setapp`
  and `macos-defaults` add their lines.
- **Mise sections:** `[mise.tools]`, `[mise.bootstrap.packages]`,
  `[mise.shell_alias]`
- **Skill `mac-setup`:** to install or remove a tool or an app, or to change a
  setting, edit `macos.toml` and run `mise run updateall:configs`.
  `mise run updateall` also installs the packages. Never run `brew install`,
  `mise use -g` or `defaults write` directly. The dev tools come from mise,
  the others from Homebrew. A file with the `Synced from bootstrap` header is
  changed in `bootstrap`. `SKILL.md` shows one bullet for each key of
  `skill.lines`, in key order. `references/local.md` is a seed: the own rules
  of the Mac. `<backup>` is the backup clone (`macos-setup`). The sync writes
  the skill only when the `claude` tool is selected and its value `skills` is
  true. Without `backup`, the sync skips the skill.
- **CA certificates:** `[mise.bootstrap.packages]` installs
  `brew:ca-certificates`.
  It is in this tool, because each Mac must have it.
- **Entry of the Mac sync:** `updateall:configs` runs `bootstrap mac sync`,
  with the flags `--dry-run`, `--no-bootstrap` and `--verbose`, and exits
  with its code. `BOOTSTRAP_BIN`, a local build, replaces the `bootstrap` of
  mise. `updateall` upgrades the Mac. It does these steps, in this order:

  1. `mise self-update`.
  2. `updateall:configs`. On exit 4 it stops with exit 4, and on another
     failure it stops with that code. After an exit 4, open a new terminal,
     then run `updateall` again.
  3. `mise install --include-task-tools`, `mise upgrade` and `mise prune`.
  4. With the `dprint` tool, `dprint config update` of
     `~/.config/dprint/dprint.jsonc`.
  5. `upgrade:brew`.
  6. `mise bootstrap --skip macos-defaults,packages`, `mise reshim` and
     `mise doctor`.
  7. `upgrade:npm` and `upgrade:ai`.
  8. Once a week: `mise bootstrap --only macos-defaults,defaults` and
     `upgrade:power`.
  9. `upgrade:macos:download`, `cleanupDS` of `PROJECTS_DIR`, and
     `tailscale:daemon`.

  The weekly steps run when the stamp `/tmp/mise-updateall-weekly.stamp` is
  missing or older than one week, and `--force` runs them now. They come
  after the upgrades of steps 1 to 7, because they ask for sudo and restart
  Finder, the Dock and Control Center. macOS clears `/tmp` at a restart, and
  each macOS update restarts the Mac, so the first `updateall` after an update
  applies them again. A step of an opt-in tool (`upgrade:brew`,
  `upgrade:npm`, `upgrade:ai`, `cleanupDS`, `tailscale:daemon`) runs only
  when its task exists, else `updateall` prints one skip line. The
  `upgrade:*` tasks download and install the macOS updates and apply the
  power profile. The aliases are `updateall`, `osx-upgrade`,
  `cancelPrintJobs` and `listServices`.
- **CLI:** `[mise.tools]` installs `github:virajp/bootstrap`, the CLI
  `bootstrap`. mise reads the release tags `bootstrap@X.Y.Z` as the
  versions `X.Y.Z`. `minimum_release_age_excludes` names
  `github:virajp/bootstrap`, so mise installs a new release of the CLI with no
  wait. mise accepts no owner pattern such as `github:virajp/*`.
- **Trust paths:** `MISE_TRUSTED_CONFIG_PATHS` holds `$HOME/.config/mise`,
  `<projects_dir>/github.com` and `<projects_dir>/gitlab.com`. mise loads a
  config below these paths with no prompt, so do not clone an untrusted
  repository below the two project folders.
- **Self-update:** `self_update.auto = true` lets mise install a new release
  before an interactive command. `self_update.minimum_release_age = "0h"` means
  that mise does not wait after a release.

### `network`

the ips:* tasks and the IP, DNS and routing aliases. Each network package is a
tool of its own.

- **Packs:** `base` (required)
- **Depends:** `doggo` (the task `ips:remoteip` runs it)
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/ips/gateway.py`
  - `~/.config/mise/tasks/ips/ii.py`
  - `~/.config/mise/tasks/ips/ips.py`
  - `~/.config/mise/tasks/ips/ipv4.py`
  - `~/.config/mise/tasks/ips/ipv6.py`
  - `~/.config/mise/tasks/ips/remoteip.py`
- **Mise sections:** `[mise.shell_alias]`

### `node`

Node.js (LTS), and the user config of the aube that mise embeds. The tool has no
pnpm and no `PNPM_HOME`, because a repository selects its own package manager.
mise installs the `npm:` tools with its embedded aube (`npm.package_manager`).
That aube reads `~/.config/aube/config.toml`, not the `[env]` of mise, so the
file holds `minimumPackageAge`: 600 minutes (10 hours) for a new package name.
`upgrade:npm` sets `fund=false` and upgrades the global npm packages, and it
does nothing when npm is not on PATH.

- **Packs:** `base` (required)
- **Files:**
  - `~/.config/aube/config.toml` — copy
- **Tasks:**
  - `~/.config/mise/tasks/upgrade/npm.py`
- **Contributes to `visual-studio-code`:** `extensions`
- **Mise sections:** `[mise.settings]`, `[mise.tools]`, `[mise.env]`

### `notion`

App to write, plan, collaborate, and get organised.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `numbers`

Numbers, from the Mac App Store.

- **Packs:** `iwork`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `obsidian`

Knowledge base that works on top Markdown files.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `opencode`

OpenCode, the AI coding agent, and its alias.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `openwhispr`

Voice-to-text dictation with AI agents, and its hotkey. OpenWhispr writes
`.env` again at each start, so the file is a seed: the sync writes it once, and
then the app owns it. The app reads only `DICTATION_KEY` and `ACTIVATION_MODE`
from it, on a new install.

- **Packs:** none
- **Files:**
  - `~/Library/Application Support/open-whispr/.env` — seed, with the hotkey
    `Alt+S` and the `push` activation mode
- **Mise sections:** `[mise.bootstrap.packages]`
- **Manual step:** OpenWhispr keeps the transcription mode and the model in its
  own storage, not in a file. On the first start, select **Local**, then
  **Whisper** with the **turbo** model. The app downloads the model to
  `~/.cache/openwhispr/`.

### `orbstack`

OrbStack (Docker and Kubernetes), its env, the docker and kubectl aliases, its
fish and zsh init and its ssh Include. The Include `~/.orbstack/ssh/config`
gives the `orb` SSH host, and the `ssh` tool puts it at the top of
`~/.ssh/config`, before each Host block. `DOCKER_DEFAULT_PLATFORM` is
`linux/amd64`, and the aliases `dpull` and `drun` also force that platform.

- **Packs:** `dev`
- **Files:**
  - `~/.config/fish/conf.d/90-orbstack.fish` — copy
  - `~/.config/zsh/90-orbstack.zsh` — copy
- **Contributes to `ssh`:** `includes`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.env]`,
  `[mise.shell_alias]`

### `orca`

Orchestrating AI coding agents.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `pages`

Pages, from the Mac App Store.

- **Packs:** `iwork`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pdf-protector`

PDF Protector, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pdf-squeezer`

PDF Squeezer, from the Mac App Store.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pre-commit`

pre-commit through mise, with a shared cache in `~/.cache/pre-commit`
(`PRE_COMMIT_HOME`).

- **Packs:** `base` (required)
- **Depends:** `python`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.env]`

### `prettyping`

prettyping, an enhanced ping, and the ping alias.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `pstree`

pstree, the process tree, with the func:myps task and its myps alias.

- **Packs:** `base` (required)
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/func/myps.py`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `python`

Python through mise. uv is the tool `uv`. mise installs Python 3.14, the
version of the Mac tasks, and does not compile it (`python.compile = false`).

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.settings]`, `[mise.tools]`

### `ripgrep`

ripgrep, the fast search tool.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `ruby`

Ruby through mise, and the RubyGems config (`~/.gemrc`): no documentation on
install and update. mise uses Ruby to run the hooks of some casks and to build
a formula without a bottle. gem writes `~/.gemrc` again (the gem sources), so
the file is a merge: the spec adds only its keys, and sets the format `yaml`
because the file has no extension.

- **Packs:** `base` (required)
- **Files:**
  - `~/.gemrc` — merge
- **Mise sections:** `[mise.tools]`

### `setapp`

Collection of apps available by subscription.

- **Packs:** `essentials`
- **Files:** none
- **Contributes to `mise`:** the line `setapp` of `skill.lines` (the skill
  `mac-setup`)
- **Mise sections:** `[mise.bootstrap.packages]`

### `sill`

Sill menu bar app.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `ssh`

the ssh client config: Host * defaults, the default user, the includes and the
agent that other tools contribute. An Include works only at the top of
`~/.ssh/config`, before each Host block, so the template puts the `includes` of
other tools there, for example the one of `orbstack`. `~/.ssh/config` also
includes `~/.ssh/config.local`, the hosts and keys of this Mac only, which no
tool syncs. ssh ignores the Include when that file is missing. In `Host *`, it
includes `~/.ssh/config.identity`, which holds `User`. The `identity_agent` of
another tool, for example `1password`, goes inside `Host *`.
`StrictHostKeyChecking` is `accept-new`. ssh refuses a `~/.ssh` folder that
other users can read, so `dir_perm` keeps it at mode 700.

- **Packs:** `base` (required)
- **Files:**
  - `~/.ssh/config` — template, format `text`, `dir_perm` `700`
  - `~/.ssh/config.identity` — template, `dir_perm` `700`
- **Contributes to `git`:** the line `ssh` of `skill.lines` (the skill
  `git-identity`)
- **Runtime:** `~/.ssh/known_hosts`, `~/.ssh/known_hosts.old`. Thus the backup
  clone links the synced files of `~/.ssh` one by one.
- **Question `user`:**
  [`macos/tools/ssh/tools.toml`](../../macos/tools/ssh/tools.toml)

### `starship`

starship prompt, its config and the fish and zsh init drop-ins. Each init runs
only in an interactive shell, and only when `starship` is on PATH. bash gets no
starship init.

- **Packs:** `base` (required)
- **Files:**
  - `~/.config/fish/conf.d/51-prompt.fish` — copy
  - `~/.config/starship.toml` — copy
  - `~/.config/zsh/02b-starship.zsh` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `sublime-text`

Text editor for code, markup and prose, and the edit alias.

- **Packs:** `dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `superfile`

superfile, the terminal file manager, and its config. superfile rewrites its
config, so the merge spec adds only the keys of bootstrap, and superfile owns
each other key. The spec owns the list `sidebar_sections` as a block, so its
order holds. The config previews code with bat and uses zoxide, so the tool
depends on `bat` and `zoxide`.

- **Packs:** `dev`
- **Depends:** `bat`, `zoxide`
- **Files:**
  - `~/Library/Application Support/superfile/config.toml` — merge
- **Mise sections:** `[mise.bootstrap.packages]`

### `tailscale`

Tailscale (the formula; tailscaled runs as a root brew service) and its daemon
and upgrade tasks. mise installs the formula, and only the daemon goes through
`brew services`, because tailscaled needs root for the utun interface and mise
has no root services on macOS. The App Store app is sandboxed and cannot run
the Tailscale SSH server, so the tool uses the formula. `updateall` runs
`tailscale:daemon`, which starts the daemon with
`sudo brew services start tailscale` only when tailscaled does not run,
because a start of a running daemon restarts it and drops the tunnel.
`upgrade:brew` runs `upgrade:tailscale` first. It stops the service, gives the
root-owned keg back to the user, so that `brew cleanup` can remove it,
upgrades the formula with `mise bootstrap packages upgrade` and starts the
service again.

- **Packs:** `essentials`
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/tailscale/daemon.py`
  - `~/.config/mise/tasks/upgrade/tailscale.py`
- **Mise sections:** `[mise.bootstrap.packages]`

### `taskexplorer`

Explore running tasks (processes).

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `telegram`

Telegram, from the Mac App Store. The task `cleanupTelegram` (also the shell
alias `cleanupTelegram`) quits Telegram, deletes the `telegram-*` cache files
larger than 1 MB in its App Store group container, and starts Telegram again.

- **Packs:** none
- **Depends:** `mas`
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/cleanupTelegram.py`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `testflight`

TestFlight, from the Mac App Store.

- **Packs:** `mobile-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `the-unarchiver`

The Unarchiver, from the Mac App Store.

- **Packs:** `essentials`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `touchid-sudo`

Touch ID for sudo (/etc/pam.d/sudo_local). Each macOS update writes
`/etc/pam.d/sudo` again, but `sudo_local` is the drop-in that it includes, and
it stays. The target is outside `$HOME`, so it is a `[mise.bootstrap.files]`
entry (mode `0444`, owner `root`, group `wheel`) that `mise bootstrap`
applies.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.files]`

### `transporter`

Transporter, from the Mac App Store.

- **Packs:** `mobile-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `tree`

tree, the directory tree.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `uv`

uv through mise, with uv as the pipx backend. Each Mac task is a uv script, so
each Mac must have uv. Each `pipx:` tool installs through uv (`pipx.uvx`).
`UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS = "never"` stop uv from
downloading a Python of its own.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.settings]`, `[mise.tools]`, `[mise.env]`

### `visual-studio-code`

Open-source code editor.

- **Packs:** `base`
- **Files:** none
- **Defaults:** `extensions`, the VS Code extensions of
  `~/.config/brewfile` (homebrew), which the task `brew:casks` installs. Each
  selected tool adds its own in `[contributes.visual-studio-code]`, and
  `extensions` of `[tools.visual-studio-code]` in `macos.toml` adds more.
- **Mise sections:** `[mise.bootstrap.packages]`

### `vlc`

Multimedia player.

- **Packs:** `media`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `vorssaint`

Menu bar toolkit with keep-awake, system monitor and volume mixer. The defaults
leave out the audio device priorities, the state of one Mac (mic mute devices,
update checks, window sizes), the onboarding and migration markers, and the
binary values (`radialMenuProfiles`, `shelfItems`) that mise cannot write.
Restart the app after `mise bootstrap` applies the defaults.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.bootstrap.macos]`
- **Compare:** in the bootstrap repository,
  `mise run vorssaint:diff <backup.plist>` compares a Vorssaint backup
  (Settings > Export) with the defaults of this tool. Add `--all` to also show
  the keys that only the backup has.

### `warp`

Rust-based terminal.

- **Packs:** `terminal`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `watch`

watch, which runs a command again and again.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `wget`

wget, the network downloader.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `whatsapp`

Native desktop client for WhatsApp.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `worktrunk`

worktrunk, a manager of git worktrees.

- **Packs:** none
- **Depends:** `git`
- **Files:** none
- **Mise sections:** `[mise.tools]`

### `xcode`

Xcode, from the Mac App Store.

- **Packs:** `mobile-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `yq`

yq, the YAML processor.

- **Packs:** `base` (required)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zed`

Multiplayer code editor.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zed-delta`

Multiplayer environment for coding with agents.

- **Packs:** none
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zoxide`

zoxide, the fast directory jumper, its shell init and the cd alias. The alias
makes `cd` run `z`. The zsh init runs only in an interactive shell, and loads
before `02b-starship.zsh` of `starship`.

- **Packs:** `base` (required)
- **Files:**
  - `~/.config/fish/conf.d/52-zoxide.fish` — copy
  - `~/.config/zsh/02a-zoxide.zsh` — copy
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `zsh`

zsh from Homebrew with zinit, its rc files and its drop-in folder
~/.config/zsh/. `~/.zshenv` and `~/.zshrc` each source each
`~/.config/zsh/*.zsh` file in name order, so an interactive shell sources each
drop-in two times. Other tools add their drop-ins to this folder (`homebrew`,
`mise`, `zoxide`, `starship`, `orbstack`), and `~/.zprofile` belongs to the
tool `mise`. `00-bootstrap.zsh` holds only the bootstrap variables, because
each other variable is in the `[env]` of mise; `GPG_TTY` is here because mise
evaluates its `[env]` in a subprocess, which has no terminal.
`00-bootstrap.zsh` adds `~/.local/bin` at the end of `PATH`, in the shell
config and not in `[env] _.path` of mise. The fzf-tab plugin of `98-zinit.zsh`
needs fzf, so the tool depends on `fzf`. `/etc/shells` is outside `$HOME`, so
`[mise.dotfiles]` adds `/opt/homebrew/bin/zsh` to it.

- **Packs:** `base` (required)
- **Stage:** `migrate`
- **Depends:** `fzf`, `homebrew`, `mise`
- **Files:**
  - `~/.config/zsh/00-bootstrap.zsh` — copy
  - `~/.config/zsh/98-zinit.zsh` — copy
  - `~/.config/zsh/99-zsh.zsh` — copy
  - `~/.zshenv` — copy
  - `~/.zshrc` — copy
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.dotfiles]`
