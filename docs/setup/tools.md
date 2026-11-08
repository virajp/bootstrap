# The Mac tools

This file is the reference of the Mac tools of `bootstrap`. An AI agent reads
it to select the tools of a Mac (see [`index.md`](index.md)).

## How a tool works

Each tool is a folder `macos/tools/<tool>/`. It holds the files of the tool at
their path below `$HOME`, and its `tools.toml`. The sync
(`mise run updateall:configs`) renders the selected tools into `$HOME`, then
runs `mise bootstrap`. The schema of `tools.toml` is in
[`docs/design/tools.md`](../design/tools.md).

`macos.toml`, at the root of the private `macos-setup` clone, selects the tools:

- `[packs.<name>]`: an empty table. It adds each default member of the pack.
- `[tools.<name>]`: a table. It adds the tool. A table for a tool of the pack
  `base` only gives values.

The sync also adds the pack `base` and each tool in `depends`, transitively.

A key of `[tools.<name>]` sets a value of that tool. It answers a question of
the tool, or it replaces a key of `[defaults]` of its `tools.toml`.

Each section below gives these items for one tool:

- **Packs:** the packs that hold the tool, and `(default)` when the tool is a
  default member of that pack.
- **Stage:** `init` for a tool of the init stage. A tool without this item has
  no stage of its own. See
  [`docs/design/setup-sync.md`](../design/setup-sync.md#stages-and-exit-4).
- **Depends:** the tools that the sync adds with this tool.
- **Files:** each target, and its mode. `generated` is written on each run.
  `seed` is written only when it is missing. `merge` is merged into the live
  file. `template` is a `.tmpl` source. The `format`, `requires` and `remove`
  keys of the `[files]` table follow.
- **Contributes to:** the keys that the tool adds to the values of another
  tool.
- **Mise sections:** the `[mise.*]` tables of `tools.toml`. They become
  `~/.config/mise/conf.d/<tool>/mise.toml`, which the sync generates.
- **Question:** the key of a value that the owner of the Mac gives. The text of
  the question is only in the `[questions.<key>]` table of the `tools.toml` that
  the item links to.

## The packs

`macos/packs.toml` holds the 14 packs. The pack `base` is always added. A tool
can be in more than one pack.

| Pack         | Default members                                                                                                                                                                        | Other members                                                                                                                    |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `base`       | `macos-setup`, `mise`, `homebrew`, `bash`, `zsh`, `fish`, `git`, `ssh`, `cli`, `fnox`, `node`, `touchid-sudo`, `starship`, `network`, `macos-defaults`, `gh`, `python`, `fonts`, `mas` | —                                                                                                                                |
| `ai`         | `claude-desktop`                                                                                                                                                                       | `claude-cli`, `claude-status`, `codeburn`, `opencode`, `openwhispr`, `orca`, `zed-delta`                                         |
| `essentials` | `alfred`, `vorssaint`, `easyfind`, `the-unarchiver`, `brave-browser`, `github-desktop`                                                                                                 | `google-chrome`, `tailscale`, `magnet`, `sill`, `taskexplorer`, `battery-health`, `setapp`, `diff-so-fancy`, `git-tools`, `glab` |
| `security`   | `1password`                                                                                                                                                                            | `1password-cli`, `1password-safari`, `bitwarden`                                                                                 |
| `dev`        | `devtools`, `dprint`, `visual-studio-code`                                                                                                                                             | `lcov`, `orbstack`, `gcloud`, `gem`, `superfile`, `zed`, `sublime-text`                                                          |
| `terminal`   | `ghostty`                                                                                                                                                                              | `warp`                                                                                                                           |
| `apple-dev`  | `xcode`                                                                                                                                                                                | `developer`, `testflight`, `transporter`, `android-studio`                                                                       |
| `iwork`      | `keynote`, `numbers`, `pages`                                                                                                                                                          | —                                                                                                                                |
| `ms-office`  | `microsoft-excel`, `microsoft-powerpoint`, `microsoft-word`                                                                                                                            | —                                                                                                                                |
| `disk`       | `daisydisk`                                                                                                                                                                            | `disk-doctor`, `disk-map`, `duplicate-detective`                                                                                 |
| `media`      | `vlc`                                                                                                                                                                                  | `final-cut-pro`, `handbrake`                                                                                                     |
| `chat`       | `telegram`, `whatsapp`                                                                                                                                                                 | —                                                                                                                                |
| `notes`      | `notion`, `obsidian`                                                                                                                                                                   | `kindle`, `cardhop`, `pdf-protector`, `pdf-squeezer`                                                                             |
| `cli-extras` | `btop`, `dasel`, `imagemagick`, `pstree`, `tree`, `watch`, `wget`, `yq`                                                                                                                | —                                                                                                                                |

## The tools

### `1password`

1Password (the app) as SSH agent and git commit signer.

- **Packs:** `security` (default)
- **Depends:** `ssh`, `git`
- **Files:**
  - `~/.config/1Password/ssh/agent.toml` — generated, template
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
- **Depends:** `homebrew`, `mise`, `macos-setup`
- **Files:**
  - `~/.bash_profile` — generated
  - `~/.bashrc` — generated
  - `~/.config/bash/00-bootstrap.bash` — generated
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
  - `~/.claude/output-styles/ste100.md` — generated
  - `~/.claude/output-styles/terse.md` — generated
  - `~/.claude/settings.json` — merge, template, remove `deniedMcpServers` and
    `env.CONTEXT7_RUNNER`
  - `~/.config/mise/tasks/func/claude` — generated
  - `~/.config/mise/tasks/upgrade/ai` — generated
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `claude-desktop`

Claude Desktop.

- **Packs:** `ai` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `claude-status`

The status line of Claude Code (from the virajp/tap tap).

- **Packs:** `ai`
- **Depends:** `claude-cli`
- **Files:** none
- **Mise sections:** `[mise.bootstrap.brew]`, `[mise.bootstrap.packages]`

### `cli`

command-line utilities (bat, eza, fzf, jq, ripgrep, zoxide, …), the projects
folder and the generic aliases.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/fish/conf.d/52-zoxide.fish` — generated
  - `~/.config/mise/tasks/cleanupDS` — generated
  - `~/.config/mise/tasks/func/backup` — generated
  - `~/.config/mise/tasks/func/code` — generated
  - `~/.config/mise/tasks/func/listStartupItems` — generated
  - `~/.config/mise/tasks/func/zip` — generated
  - `~/.config/mise/tasks/keychain/delete` — generated
  - `~/.config/zsh/02a-zoxide.zsh` — generated
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.env]`,
  `[mise.shell_alias]`

### `codeburn`

codeburn, the usage and cost report of the AI tools, and its aistats alias.

- **Packs:** `ai`
- **Depends:** `claude-cli`
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
  - `~/.config/fish/conf.d/53-dasel.fish` — generated
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
  - `~/.config/dprint/dprint.jsonc` — generated, format `text`
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

fish from Homebrew, its config, drop-in folder conf.d/ and functions, and
the login shell.

- **Packs:** `base` (default)
- **Stage:** `init`
- **Depends:** `homebrew`, `mise`, `macos-setup`
- **Files:**
  - `~/.config/fish/conf.d/01-env.fish` — generated
  - `~/.config/fish/conf.d/02-path.fish` — generated
  - `~/.config/fish/conf.d/fish_frozen_key_bindings.fish` — seed
  - `~/.config/fish/conf.d/fish_frozen_theme.fish` — seed
  - `~/.config/fish/config.fish` — generated
  - `~/.config/fish/functions/fish_title.fish` — seed
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.bootstrap.user]`,
  `[mise.dotfiles]`

### `fnox`

fnox and age: secrets from the macOS Keychain, with a profile per consumer.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/fnox/config.toml` — generated
- **Mise sections:** `[mise.bootstrap.packages]`

### `fonts`

The Nerd Fonts casks (Fira Code, Hack, Iosevka, Iosevka Term, Meslo LG).

- **Packs:** `base` (default)
- **Files:** none
- **Mise sections:** `[mise.bootstrap.packages]`

### `gcloud`

The Google Cloud CLI (gcloud) and its env. Its mise entry depends on `python`
and `uv`, and `CLOUDSDK_PYTHON` names the Python of the `python` tool.

- **Packs:** `dev`
- **Files:** none
- **Mise sections:** `[mise.tools]`, `[mise.env]`

### `gem`

The RubyGems config (~/.gemrc): no documentation on install and update.

- **Packs:** `dev`
- **Files:**
  - `~/.gemrc` — merge, format `yaml`

### `gh`

the GitHub CLI (gh) and its config: ssh as git protocol, the editor, the
aliases.

- **Packs:** `base` (default)
- **Depends:** `git`
- **Files:**
  - `~/.config/gh/config.yml` — merge, template
- **Mise sections:** `[mise.bootstrap.packages]`
- **Question `editor`:**
  [`macos/tools/gh/tools.toml`](../../macos/tools/gh/tools.toml)

### `ghostty`

Ghostty, the terminal emulator, and its config.

- **Packs:** `terminal` (default)
- **Files:**
  - `~/.config/ghostty/config` — generated
- **Mise sections:** `[mise.bootstrap.packages]`

### `git`

git, its global config, the commit identity per forge (GitHub, GitLab) and the
git aliases.

- **Packs:** `base` (default)
- **Depends:** `cli`, `diff-so-fancy`
- **Files:**
  - `~/.config/git/allowed_signers` — generated, template
  - `~/.config/git/config` — generated, template
  - `~/.config/git/github.config` — generated, template
  - `~/.config/git/gitlab.config` — generated, template
  - `~/.config/git/identity` — generated, template
  - `~/.config/git/ignore` — generated
  - `~/.github/git-commit-guidelines.md` — generated
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
  - `~/.config/bash/01a-homebrew.bash` — generated
  - `~/.config/brewfile` — generated
  - `~/.config/fish/conf.d/05-homebrew.fish` — generated
  - `~/.config/mise/tasks/brew/casks` — generated
  - `~/.config/mise/tasks/upgrade/brew` — generated
  - `~/.config/zsh/01a-homebrew.zsh` — generated
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

### `macos-setup`

The Mac sync entry: updateall:configs, updateall and the upgrade tasks.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/mise/tasks/_scripts/helpers` — generated
  - `~/.config/mise/tasks/updateall/_default` — generated
  - `~/.config/mise/tasks/updateall/configs/_default` — generated, template
  - `~/.config/mise/tasks/updateall/configs/sync` — generated
  - `~/.config/mise/tasks/upgrade/macos/download` — generated
  - `~/.config/mise/tasks/upgrade/macos/upgrade` — generated
  - `~/.config/mise/tasks/upgrade/power` — generated
- **Mise sections:** `[mise.env]`
- **Question `config_path`:**
  [`macos/tools/macos-setup/tools.toml`](../../macos/tools/macos-setup/tools.toml)

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

mise: its user-level config, its activation in bash, zsh and fish, and
~/.zprofile.

- **Packs:** `base` (default)
- **Depends:** `cli`
- **Files:**
  - `~/.config/bash/01b-mise.bash` — generated, template
  - `~/.config/fish/conf.d/06-mise.fish` — generated, template
  - `~/.config/mise/config.toml` — generated
  - `~/.config/mise/miserc.toml` — generated
  - `~/.config/zsh/01b-mise.zsh` — generated, template
  - `~/.zprofile` — generated
- **Trust paths:** `MISE_TRUSTED_CONFIG_PATHS` holds `$HOME/.config/mise`,
  `<projects_dir>/github.com` and `<projects_dir>/gitlab.com`. mise loads a
  config below these paths with no prompt, so do not clone an untrusted
  repository below the two project folders.
- **Self-update:** `self_update.auto = true` lets mise install a new release
  before an interactive command. `self_update.minimum_release_age = "0h"` means
  that mise does not wait after a release.
- **Mise sections:** `[mise.tools]`, `[mise.shell_alias]`

### `network`

network utilities (doggo, gping, httping, ipcalc, prettyping), the ips:* tasks
and the DNS, ping and routing aliases.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/mise/tasks/ips/gateway` — generated
  - `~/.config/mise/tasks/ips/ii` — generated
  - `~/.config/mise/tasks/ips/ips` — generated
  - `~/.config/mise/tasks/ips/ipv4` — generated
  - `~/.config/mise/tasks/ips/ipv6` — generated
  - `~/.config/mise/tasks/ips/remoteip` — generated
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.shell_alias]`

### `node`

Node.js (LTS), and the user config of the aube that mise embeds. The tool has no
pnpm and no `PNPM_HOME`, because a repository selects its own package manager.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/aube/config.toml` — generated
  - `~/.config/mise/tasks/upgrade/npm` — generated
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
  - `~/.config/fish/conf.d/90-orbstack.fish` — generated
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
- **Files:**
  - `~/.config/mise/tasks/func/myps` — generated
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
  - `~/.ssh/config` — generated, template, format `text`
  - `~/.ssh/config.identity` — generated, template
- **Folder mode:** the sync keeps `~/.ssh` at mode 700.
- **Question `user`:**
  [`macos/tools/ssh/tools.toml`](../../macos/tools/ssh/tools.toml)

### `starship`

starship prompt, its config and the fish and zsh init drop-ins.

- **Packs:** `base` (default)
- **Files:**
  - `~/.config/fish/conf.d/51-prompt.fish` — generated
  - `~/.config/starship.toml` — generated
  - `~/.config/zsh/02b-starship.zsh` — generated
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
  - `~/Library/Application Support/superfile/config.toml` — merge, remove
    `sidebar_sections`
- **Mise sections:** `[mise.bootstrap.packages]`

### `tailscale`

Tailscale (the formula; tailscaled runs as a root brew service) and its daemon
and upgrade tasks.

- **Packs:** `essentials`
- **Files:**
  - `~/.config/mise/tasks/tailscale/daemon` — generated
  - `~/.config/mise/tasks/upgrade/tailscale` — generated
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
- **Files:**
  - `~/.config/mise/tasks/cleanupTelegram` — generated
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
- **Compare:** `mise run vorssaint:diff <backup.plist>` compares a Vorssaint
  backup (Settings > Export) with the defaults of this tool. Add `--all` to
  also show the keys that only the backup has.

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
- **Depends:** `homebrew`, `mise`, `macos-setup`
- **Files:**
  - `~/.config/zsh/00-bootstrap.zsh` — generated
  - `~/.config/zsh/98-zinit.zsh` — generated
  - `~/.config/zsh/99-zsh.zsh` — generated
  - `~/.zshenv` — generated
  - `~/.zshrc` — generated
- **Mise sections:** `[mise.bootstrap.packages]`, `[mise.dotfiles]`
