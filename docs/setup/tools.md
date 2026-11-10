# The Mac tools

This file is the reference of the Mac tools of `bootstrap`. An AI agent reads
it to select the tools of a Mac (see [`index.md`](index.md)).

## How a tool works

Each tool is a folder `macos/tools/<tool>/` with a `tools.toml`. Its `[files]`
table gives each target below `$HOME` and its source in `files/`. Its `[tasks]`
table gives each task below `~/.config/mise/tasks/` and its source in `tasks/`.
A tool can also have scripts in `scripts/`. The sync runs a script, and never
copies it. The schema of `tools.toml` is in
[`docs/design/tools.md`](../design/tools.md). The rules of the Mac sync are in
[`docs/design/mac.md`](../design/mac.md).

`bootstrap mac sync` renders the selected tools into `$HOME`, then runs
`mise bootstrap`. The task `mise run updateall:configs` of the tool `mise` runs
that command.

`macos.toml`, at the root of the private `macos-setup` clone, selects the tools:

- `[packs.<name>]`: an empty table. It adds each default member of the pack.
- `[tools.<name>]`: a table. It adds the tool. A table for a tool of the pack
  `base` only gives values.

The sync also adds the pack `base` and each tool in `depends`, transitively.
`bootstrap mac add`, `bootstrap mac remove`, `bootstrap mac replace` and
`bootstrap mac detect` change the `[tools.<name>]` tables. No command of the CLI
changes a `[packs.<name>]` table.

A key of `[tools.<name>]` sets a value of that tool. It answers a question of
the tool, or it replaces a key of `[defaults]` of its `tools.toml`.

Each section below gives these items for one tool:

- **Packs:** the packs that hold the tool. `(default)` marks a default member.
  `(required)` marks a member of `base` that each Mac must have.
- **Stage:** `init` for a tool of the init stage. A tool without this item has
  no stage of its own. See [`docs/design/mac.md`](../design/mac.md).
- **Depends:** the tools that the sync adds with this tool.
- **Files:** each target, and its mode. `copy` is copied on each run.
  `template` is rendered on each run. `seed` is written only when it is
  missing. `merge` adds the items of a merge spec (`files/<name>.merge.toml`)
  to the live file, and the other items of the live file stay. The `format` and
  `dir_perm` keys of the `[files]` table follow.
- **Tasks:** each task file below `~/.config/mise/tasks/`. `library` marks a
  library of `_scripts/`, which is not a task.
- **Scripts:** the scripts in `scripts/` of the tool. The sync runs them, and
  does not copy them.
- **Contributes to:** the keys that the tool adds to the values of another
  tool.
- **Defaults:** the keys of `[defaults]`. A key of `[tools.<name>]` in
  `macos.toml` replaces it.
- **Mise sections:** the `[mise.*]` tables of `tools.toml`. They become
  `~/.config/mise/conf.d/<tool>/mise.toml`, which the sync generates.
- **Question:** the key of a value that the owner of the Mac gives. The text of
  the question is only in the `[questions.<key>]` table of the `tools.toml` that
  the item links to. `bootstrap mac questions` shows it.

## The packs

`macos/packs.toml` holds the 14 packs. The pack `base` is always added. A tool
can be in more than one pack. The member `mise` of `base` has `required = true`.

| Pack         | Default members                                                                                                                                                         | Other members                                                                                                                    |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `base`       | `mise`, `homebrew`, `bash`, `zsh`, `fish`, `git`, `ssh`, `cli`, `fnox`, `node`, `touchid-sudo`, `starship`, `network`, `macos-defaults`, `gh`, `python`, `fonts`, `mas` | —                                                                                                                                |
| `ai`         | `claude-desktop`                                                                                                                                                        | `claude-cli`, `claude-status`, `codeburn`, `opencode`, `openwhispr`, `orca`, `zed-delta`                                         |
| `essentials` | `alfred`, `vorssaint`, `easyfind`, `the-unarchiver`, `brave-browser`, `github-desktop`                                                                                  | `google-chrome`, `tailscale`, `magnet`, `sill`, `taskexplorer`, `battery-health`, `setapp`, `diff-so-fancy`, `git-tools`, `glab` |
| `security`   | `1password`                                                                                                                                                             | `1password-cli`, `1password-safari`, `bitwarden`                                                                                 |
| `dev`        | `devtools`, `dprint`, `visual-studio-code`                                                                                                                              | `lcov`, `orbstack`, `gcloud`, `gem`, `superfile`, `zed`, `sublime-text`                                                          |
| `terminal`   | `ghostty`                                                                                                                                                               | `warp`                                                                                                                           |
| `apple-dev`  | `xcode`                                                                                                                                                                 | `developer`, `testflight`, `transporter`, `android-studio`                                                                       |
| `iwork`      | `keynote`, `numbers`, `pages`                                                                                                                                           | —                                                                                                                                |
| `ms-office`  | `microsoft-excel`, `microsoft-powerpoint`, `microsoft-word`                                                                                                             | —                                                                                                                                |
| `disk`       | `daisydisk`                                                                                                                                                             | `disk-doctor`, `disk-map`, `duplicate-detective`                                                                                 |
| `media`      | `vlc`                                                                                                                                                                   | `final-cut-pro`, `handbrake`                                                                                                     |
| `chat`       | `telegram`, `whatsapp`                                                                                                                                                  | —                                                                                                                                |
| `notes`      | `notion`, `obsidian`                                                                                                                                                    | `kindle`, `cardhop`, `pdf-protector`, `pdf-squeezer`                                                                             |
| `cli-extras` | `btop`, `dasel`, `imagemagick`, `pstree`, `tree`, `watch`, `wget`, `yq`                                                                                                 | —                                                                                                                                |

## The tools

### `1password`

1Password (the app) as SSH agent and git commit signer.

- **Packs:** `security` (default)
- **Depends:** `ssh`, `git`
- **Files:**
  - `~/.config/1Password/ssh/agent.toml` — template
- **Contributes to `ssh`:** `identity_agent`
- **Contributes to `git`:** `gpg.ssh.program`
- **Mise sections:** `[mise.bootstrap.packages]`
- **Question `ssh_vault`:**
  [`macos/tools/1password/tools.toml`](../../macos/tools/1password/tools.toml)

### `1password-cli`

The 1Password CLI (op).

- **Packs:** `security`
- **Depends:** `1password`
- **Files:** none
- **Mise sections:** `[mise.tools]`

### `1password-safari`

1Password for Safari, from the Mac App Store.

- **Packs:** `security`
- **Depends:** `mas`, `1password`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `alfred`

Application launcher and productivity software.

- **Packs:** `essentials` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `android-studio`

Tools for building Android applications.

- **Packs:** `apple-dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `bash`

bash from Homebrew, its rc files and its drop-in folder ~/.config/bash/.

- **Packs:** `base` (default)
- **Stage:** `init`
- **Depends:** `homebrew`, `mise`
- **Files:**
  - `~/.bash_profile` — copy
  - `~/.bashrc` — copy
  - `~/.config/bash/00-bootstrap.bash` — copy
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.dotfiles]`

### `battery-health`

Battery Health 2, from the Mac App Store.

- **Packs:** `essentials`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `bitwarden`

Bitwarden, the password manager.

- **Packs:** `security`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `brave-browser`

Brave Browser.

- **Packs:** `essentials` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `btop`

btop, the resource monitor for the terminal.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `cardhop`

Cardhop, from the Mac App Store.

- **Packs:** `notes`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `claude-cli`

Claude Code: its settings, CLAUDE.md, output styles, the func:claude and
upgrade:ai tasks and the cc* aliases.

- **Packs:** `ai`
- **Depends:** `cli`
- **Files:**
  - `~/.claude/CLAUDE.md` — seed
  - `~/.claude/output-styles/ste100.md` — copy
  - `~/.claude/output-styles/terse.md` — copy
  - `~/.claude/settings.json` — merge
- **Tasks:**
  - `~/.config/mise/tasks/func/claude`
  - `~/.config/mise/tasks/upgrade/ai`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `claude-desktop`

Claude Desktop.

- **Packs:** `ai` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `claude-status`

The status line of Claude Code (from the virajp/tap tap). When this tool is
selected, the merge spec of `~/.claude/settings.json` of `claude-cli` adds
`statusLine`, `subagentStatusLine` and the `PostToolUse` hook of
`claude-status`.

- **Packs:** `ai`
- **Depends:** `claude-cli`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `cli`

command-line utilities (bat, eza, fzf, jq, ripgrep, zoxide, …), the projects
folder and the generic aliases.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/fish/conf.d/52-zoxide.fish` — copy
  - `~/.config/zsh/02a-zoxide.zsh` — copy
- **Tasks:**
  - `~/.config/mise/tasks/cleanupDS`
  - `~/.config/mise/tasks/func/backup`
  - `~/.config/mise/tasks/func/code`
  - `~/.config/mise/tasks/func/listStartupItems`
  - `~/.config/mise/tasks/func/zip`
  - `~/.config/mise/tasks/keychain/delete`
- **Defaults:** `projects_dir`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.env]`,
  `[mise.shell_alias]`

### `codeburn`

codeburn, the usage and cost report of the AI tools, and its aistats alias.

- **Packs:** `ai`
- **Depends:** `claude-cli`, `node`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.shell_alias]`

### `daisydisk`

DaisyDisk, from the Mac App Store.

- **Packs:** `disk` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `dasel`

dasel, the query tool for JSON, YAML and TOML, and its fish completion.

- **Packs:** `cli-extras` (default)
- **Files:**
  - `~/.config/fish/conf.d/53-dasel.fish` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `developer`

Developer, from the Mac App Store.

- **Packs:** `apple-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `devtools`

The development env (browser for tests, telemetry opt-out, pre-commit cache,
runtime env).

- **Packs:** `dev` (default)
- **Files:** none
- **Mise sections:** `[mise.env]`

### `diff-so-fancy`

diff-so-fancy, the git diff formatter (the pager of the git config).

- **Packs:** `essentials`
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

### `dprint`

dprint, the code formatter, and its global config.

- **Packs:** `dev` (default)
- **Files:**
  - `~/.config/dprint/dprint.jsonc` — copy, format `text`
- **Mise sections:** `[mise.tools]`

### `duplicate-detective`

Duplicate Detective, from the Mac App Store.

- **Packs:** `disk`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `easyfind`

Find files, folders, or contents in any file.

- **Packs:** `essentials` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `final-cut-pro`

Final Cut Pro, from the Mac App Store.

- **Packs:** `media`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `fish`

fish from Homebrew, its config, drop-in folder conf.d/ and functions, and the
login shell.

- **Packs:** `base` (default)
- **Stage:** `init`
- **Depends:** `homebrew`, `mise`
- **Files:**
  - `~/.config/fish/conf.d/01-env.fish` — copy
  - `~/.config/fish/conf.d/02-path.fish` — copy
  - `~/.config/fish/conf.d/fish_frozen_key_bindings.fish` — seed
  - `~/.config/fish/conf.d/fish_frozen_theme.fish` — seed
  - `~/.config/fish/config.fish` — copy
  - `~/.config/fish/functions/fish_title.fish` — seed
- **Scripts:** `after`. When the login shell is not the shell of the
  terminal, it asks for a new terminal, and the sync stops with exit 4.
- **Defaults:** `login_shell`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.bootstrap.user]`,
  `[mise.dotfiles]`

### `fnox`

fnox and age: secrets from the macOS Keychain, with a profile per consumer.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/fnox/config.toml` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `fonts`

The Nerd Fonts casks (Fira Code, Hack, Iosevka, Iosevka Term, Meslo LG).

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `gcloud`

The Google Cloud CLI (gcloud) and its env. `CLOUDSDK_PYTHON` names the Python
3.14 of the `python` tool.

- **Packs:** `dev`
- **Depends:** `python`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.env]`

### `gem`

The RubyGems config (~/.gemrc): no documentation on install and update.

- **Packs:** `dev`
- **Files:**
  - `~/.gemrc` — merge

### `gh`

the GitHub CLI (gh) and its config: ssh as git protocol, the editor, the
aliases.

- **Packs:** `base` (default)
- **Depends:** `git`
- **Files:**
  - `~/.config/gh/config.yml` — merge
- **Mise sections:** `[mise.bootstrap.packages]`
- **Question `editor`:**
  [`macos/tools/gh/tools.toml`](../../macos/tools/gh/tools.toml)

### `ghostty`

Ghostty, the terminal emulator, and its config.

- **Packs:** `terminal` (default)
- **Files:**
  - `~/.config/ghostty/config` — copy
- **Mise sections:** `[mise.bootstrap.packages]`

### `git`

git, its global config, the commit identity per forge (GitHub, GitLab) and the
git aliases.

- **Packs:** `base` (default)
- **Depends:** `cli`, `diff-so-fancy`
- **Files:**
  - `~/.config/git/allowed_signers` — template
  - `~/.config/git/config` — template
  - `~/.config/git/github.config` — template
  - `~/.config/git/gitlab.config` — template
  - `~/.config/git/identity` — template
  - `~/.config/git/ignore` — copy
  - `~/.github/git-commit-guidelines.md` — copy
- **Defaults:** `core_editor`
- **Mise sections:** `[mise.tools]`, `[mise.bootstrap.packages]`,
  `[mise.shell_alias]`
- **Question `name`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `email_github`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `email_gitlab`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)
- **Question `signing_key`:**
  [`macos/tools/git/tools.toml`](../../macos/tools/git/tools.toml)

### `git-tools`

git-tools, more git commands.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `github-desktop`

GitHub Desktop.

- **Packs:** `essentials` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `glab`

glab, the GitLab CLI.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `google-chrome`

Google Chrome.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `handbrake`

Open-source video transcoder.

- **Packs:** `media`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `homebrew`

Homebrew: its environment, the brew alias, the Brewfile, its upgrade task and
its shellenv in bash, zsh and fish.

- **Packs:** `base` (default)
- **Depends:** `fnox`
- **Files:**
  - `~/.config/bash/01a-homebrew.bash` — copy
  - `~/.config/brewfile` — copy
  - `~/.config/fish/conf.d/05-homebrew.fish` — copy
  - `~/.config/zsh/01a-homebrew.zsh` — copy
- **Tasks:**
  - `~/.config/mise/tasks/brew/casks`
  - `~/.config/mise/tasks/upgrade/brew`
- **Mise sections:** `[mise.env]`, `[mise.shell_alias]`

### `imagemagick`

ImageMagick, the image manipulation tool.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `keynote`

Keynote, from the Mac App Store.

- **Packs:** `iwork` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `kindle`

Kindle, from the Mac App Store.

- **Packs:** `notes`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `lcov`

lcov, the code coverage tool.

- **Packs:** `dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `macos-defaults`

macOS settings: Finder, Dock, Mission Control, window management, menu bar clock
and trackpad.

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.macos]`, `[mise.bootstrap.hooks]`
- **Question `dock_orientation`:**
  [`macos/tools/macos-defaults/tools.toml`](../../macos/tools/macos-defaults/tools.toml)

### `magnet`

Magnet, from the Mac App Store.

- **Packs:** `essentials`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `mas`

Mac App Store command-line interface.

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-excel`

Microsoft Excel, from the Mac App Store.

- **Packs:** `ms-office` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-powerpoint`

Microsoft PowerPoint, from the Mac App Store.

- **Packs:** `ms-office` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `microsoft-word`

Microsoft Word, from the Mac App Store.

- **Packs:** `ms-office` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `mise`

mise: its user-level config, its activation in bash, zsh and fish, ~/.zprofile,
and the entry tasks of the Mac sync.

- **Packs:** `base` (default, required)
- **Depends:** `cli`
- **Files:**
  - `~/.config/bash/01b-mise.bash` — template
  - `~/.config/fish/conf.d/06-mise.fish` — template
  - `~/.config/mise/config.toml` — copy
  - `~/.config/mise/miserc.toml` — copy
  - `~/.config/zsh/01b-mise.zsh` — template
  - `~/.zprofile` — copy
- **Tasks:**
  - `~/.config/mise/tasks/_scripts/helpers` — library
  - `~/.config/mise/tasks/updateall/_default`
  - `~/.config/mise/tasks/updateall/configs/_default`
  - `~/.config/mise/tasks/upgrade/macos/download`
  - `~/.config/mise/tasks/upgrade/macos/upgrade`
  - `~/.config/mise/tasks/upgrade/power`
- **Mise sections:** `[mise.env]`, `[mise.tools]`, `[mise.shell_alias]`
- **Entry of the Mac sync:** `updateall:configs` runs `bootstrap mac sync`, with
  the flags `--dry-run`, `--no-bootstrap` and `--verbose`. `updateall` upgrades
  the Mac, and runs `updateall:configs --no-bootstrap` as one step. The
  `upgrade:*` tasks download and install the macOS updates and apply the power
  profile. The aliases are `updateall` and `osx-upgrade`.
- **CLI:** `[mise.tools]` installs `github:virajp/bootstrap`, the CLI
  `bootstrap`, with `version_prefix = "cli-v"`, because the release tags are
  `cli-vX.Y.Z`.
- **Trust paths:** `MISE_TRUSTED_CONFIG_PATHS` holds `$HOME/.config/mise`,
  `<projects_dir>/github.com` and `<projects_dir>/gitlab.com`. mise loads a
  config below these paths with no prompt, so do not clone an untrusted
  repository below the two project folders.
- **Self-update:** `self_update.auto = true` lets mise install a new release
  before an interactive command. `self_update.minimum_release_age = "0h"` means
  that mise does not wait after a release.
- **Question `config_path`:**
  [`macos/tools/mise/tools.toml`](../../macos/tools/mise/tools.toml). The answer
  is the folder of `macos.toml`. The sync writes it to the mise env as
  `MACOS_SETUP_REPO_PATH`, and `~` at the start becomes `$HOME`. Each
  `bootstrap mac` command except `mac init` finds `macos.toml` through this
  variable.

### `network`

network utilities (doggo, gping, httping, ipcalc, prettyping), the ips:* tasks
and the DNS, ping and routing aliases.

- **Packs:** `base` (default)
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/ips/gateway`
  - `~/.config/mise/tasks/ips/ii`
  - `~/.config/mise/tasks/ips/ips`
  - `~/.config/mise/tasks/ips/ipv4`
  - `~/.config/mise/tasks/ips/ipv6`
  - `~/.config/mise/tasks/ips/remoteip`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `node`

Node.js (LTS), and the user config of the aube that mise embeds. The tool has no
pnpm and no `PNPM_HOME`, because a repository selects its own package manager.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/aube/config.toml` — copy
- **Tasks:**
  - `~/.config/mise/tasks/upgrade/npm`
- **Mise sections:** `[mise.settings]`, `[mise.tools]`, `[mise.env]`

### `notion`

App to write, plan, collaborate, and get organised.

- **Packs:** `notes` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `numbers`

Numbers, from the Mac App Store.

- **Packs:** `iwork` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `obsidian`

Knowledge base that works on top Markdown files.

- **Packs:** `notes` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `opencode`

OpenCode, the AI coding agent, and its alias.

- **Packs:** `ai`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `openwhispr`

Voice-to-text dictation with AI agents, and its hotkey.

- **Packs:** `ai`
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
fish init and its ssh Include.

- **Packs:** `dev`
- **Files:**
  - `~/.config/fish/conf.d/90-orbstack.fish` — copy
- **Contributes to `ssh`:** `includes`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.env]`,
  `[mise.shell_alias]`

### `orca`

Orchestrating AI coding agents.

- **Packs:** `ai`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `pages`

Pages, from the Mac App Store.

- **Packs:** `iwork` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pdf-protector`

PDF Protector, from the Mac App Store.

- **Packs:** `notes`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pdf-squeezer`

PDF Squeezer, from the Mac App Store.

- **Packs:** `notes`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `pstree`

pstree, the process tree, with the func:myps task and its myps alias.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/func/myps`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `python`

Python and uv through mise, with uv as the pipx backend.

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.settings]`, `[mise.tools]`, `[mise.env]`

### `setapp`

Collection of apps available by subscription.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `sill`

Sill menu bar app.

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `ssh`

the ssh client config: Host * defaults, the default user, the includes and the
agent that other tools contribute.

- **Packs:** `base` (default)
- **Files:**
  - `~/.ssh/config` — template, format `text`, `dir_perm` `700`
  - `~/.ssh/config.identity` — template, `dir_perm` `700`
- **Folder mode:** `dir_perm` keeps the folder `~/.ssh` at mode 700.
- **Question `user`:**
  [`macos/tools/ssh/tools.toml`](../../macos/tools/ssh/tools.toml)

### `starship`

starship prompt, its config and the fish and zsh init drop-ins.

- **Packs:** `base` (default)
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

superfile, the terminal file manager, and its config.

- **Packs:** `dev`
- **Files:**
  - `~/Library/Application Support/superfile/config.toml` — merge
- **Mise sections:** `[mise.bootstrap.packages]`

### `tailscale`

Tailscale (the formula; tailscaled runs as a root brew service) and its daemon
and upgrade tasks.

- **Packs:** `essentials`
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/tailscale/daemon`
  - `~/.config/mise/tasks/upgrade/tailscale`
- **Mise sections:** `[mise.bootstrap.packages]`

### `taskexplorer`

Explore running tasks (processes).

- **Packs:** `essentials`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `telegram`

Telegram, from the Mac App Store.

- **Packs:** `chat` (default)
- **Depends:** `mas`
- **Files:** none
- **Tasks:**
  - `~/.config/mise/tasks/cleanupTelegram`
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `testflight`

TestFlight, from the Mac App Store.

- **Packs:** `apple-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `the-unarchiver`

The Unarchiver, from the Mac App Store.

- **Packs:** `essentials` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `touchid-sudo`

Touch ID for sudo (/etc/pam.d/sudo_local).

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.files]`

### `transporter`

Transporter, from the Mac App Store.

- **Packs:** `apple-dev`
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `tree`

tree, the directory tree.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `visual-studio-code`

Open-source code editor.

- **Packs:** `dev` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `vlc`

Multimedia player.

- **Packs:** `media` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `vorssaint`

Menu bar toolkit with keep-awake, system monitor and volume mixer.

- **Packs:** `essentials` (default)
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

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `wget`

wget, the network downloader.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `whatsapp`

Native desktop client for WhatsApp.

- **Packs:** `chat` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `xcode`

Xcode, from the Mac App Store.

- **Packs:** `apple-dev` (default)
- **Depends:** `mas`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `yq`

yq, the YAML processor.

- **Packs:** `cli-extras` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zed`

Multiplayer code editor.

- **Packs:** `dev`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zed-delta`

Multiplayer environment for coding with agents.

- **Packs:** `ai`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `zsh`

zsh from Homebrew with zinit, its rc files and its drop-in folder
~/.config/zsh/.

- **Packs:** `base` (default)
- **Stage:** `init`
- **Depends:** `homebrew`, `mise`
- **Files:**
  - `~/.config/zsh/00-bootstrap.zsh` — copy
  - `~/.config/zsh/98-zinit.zsh` — copy
  - `~/.config/zsh/99-zsh.zsh` — copy
  - `~/.zshenv` — copy
  - `~/.zshrc` — copy
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.dotfiles]`
