# Audit part B — duplicate-detective to mas

## Findings

### `fish`

- F1 missing prose — doc: none — evidence:
  macos/tools/fish/files/.config/fish/conf.d/01-env.fish:8,14-17,
  macos/tools/fish/files/.config/fish/conf.d/02-path.fish:14-22,
  macos/tools/mise/tools.toml:52, macos/tools/fish/tools.toml:27-29 — new:
  "fish loads `conf.d/` in name order: `01-env` and `02-path` set the XDG
  folders, `HOMEBREW_PREFIX` and the PATH before mise activates at
  `06-mise.fish`, so the tools of mise win on PATH. `/etc/shells` is outside
  `$HOME`, so the `[mise.dotfiles]` table adds Homebrew's fish to it."
- F2 order Contributes to — doc: "- **Defaults:** `login_shell`" followed by
  "- **Contributes to `visual-studio-code`:** `extensions`" — evidence:
  docs/setup/tools.md:58-61 (the legend puts Contributes to before Defaults),
  macos/tools/fish/tools.toml:16-22 — new: the two items in this order, text
  unchanged:

  ```markdown
  - **Contributes to `visual-studio-code`:** `extensions`
  - **Defaults:** `login_shell`
  ```

### `fnox`

- F3 missing prose — doc: none — evidence:
  macos/tools/fnox/files/.config/fnox/config.toml:5-6,10-13,19-32 — new: "The
  config has one provider, `global` (the macOS Keychain), and `env = false`,
  so fnox exports no secret to the shell: only the command of
  `fnox exec -P <profile>` gets it. The profile `brew` gives
  `GITHUB_API_TOKEN` and `HOMEBREW_GITHUB_API_TOKEN` to brew, and the profile
  `fnox` gives `AGE_KEY`."

### `fonts`

- F4 order section — doc: the section of this tool (docs/setup/tools.md:507)
  comes after the section of the next tool by sort order
  (docs/setup/tools.md:499) — evidence: `LC_ALL=C command ls macos/tools` lists
  this folder first ("fo" sorts before "fz") — new: move the whole section of
  this tool, text unchanged, up to the place right after the section of
  `fnox`.

### `gcloud`

- F5 missing prose — doc: none (the doc has only "`CLOUDSDK_PYTHON` names the
  Python 3.14 of the `python` tool.") — evidence:
  macos/tools/gcloud/tools.toml:14-20, macos/tools/python/tools.toml:12 —
  new, after the sentence that names `CLOUDSDK_PYTHON`: "Without it, gcloud
  runs the system `python3` (3.9 on macOS) and fails. The path names the minor
  version 3.14, not `latest`, because `latest` moves to each newer Python, and
  the gcloud virtual environment in `~/.config/gcloud/virtenv` then runs
  without its packages."

### `gh`

- F6 missing prose — doc: none — evidence: macos/tools/gh/tools.toml:21-22,
  macos/tools/gh/files/.config/gh/config.yml.merge.toml:1-3,8-19 — new: "`gh
  config set` rewrites `config.yml`, so the file is a merge target: the spec
  adds only `version`, `git_protocol`, `editor`, `pager` and the alias `co`,
  and gh owns each other key."

### `ghostty`

- F7 missing prose — doc: none — evidence:
  macos/tools/ghostty/tools.toml:1-2,5,
  macos/tools/ghostty/files/.config/ghostty/config:5 — new: "The config uses
  the font MesloLGS Nerd Font, so the tool depends on `fonts`."

### `git`

- F8 missing prose — doc: none — evidence:
  macos/tools/git/files/.config/git/identity.tmpl:5,11-29,
  macos/tools/git/tools.toml:2-3,7,
  macos/tools/git/files/.config/git/config.tmpl:77-78,89-99,
  macos/tools/1password/tools.toml:20 — new: "`~/.config/git/identity`
  includes `github.config` for a GitHub remote or a folder below
  `<projects_dir>/github.com/`, and `gitlab.config` in the same way for
  GitLab; `projects_dir` is a value of the `cli` tool, so `git` depends on
  `cli`. The config signs each commit and each annotated tag with SSH, and the
  `program` line of `[gpg "ssh"]` renders only when `1password` gives
  `gpg.ssh.program`."

### `homebrew`

- F9 missing prose — doc: none — evidence: macos/tools/homebrew/tools.toml:3,
  15-17,31,
  macos/tools/homebrew/files/.config/fish/conf.d/05-homebrew.fish:16-17,
  macos/tools/mise/tools.toml:48,52,62,
  macos/tools/homebrew/files/.config/brewfile:3-9 — new: "The `brew` alias
  runs brew through the fnox profile `brew`, so the tool depends on `fnox`.
  The shellenv files (`01a-`, `05-`) run before the mise drop-ins (`01b-`,
  `06-`), so the tools of mise win on PATH, and the Brewfile holds only the
  VS Code extensions of `visual-studio-code`, with no line when that tool is
  not selected."
- F10 missing prose — doc: none — evidence:
  macos/tools/homebrew/tasks/upgrade-brew.py:49-68,
  macos/tools/homebrew/tasks/brew-casks.py:13,43-47 — new, after the text of
  F9: "`upgrade:brew` updates Homebrew, runs `upgrade:tailscale` when that
  task exists, then prunes, upgrades and applies the packages of
  `[mise.bootstrap.packages]`. `brew:casks` runs by hand only: it installs the
  Brewfile, and removes each cask and VS Code extension that the Brewfile does
  not hold."

### `macos-defaults`

- F11 missing prose — doc: none — evidence:
  macos/tools/macos-defaults/tools.toml:8-9,83,127-137,
  macos/tools/mise/tools.toml:107 — new: "`dock_orientation` must be
  `bottom`, `left` or `right`, else the sync stops. The hook `post-defaults`
  writes `NewWindowTargetPath` with `$HOME`, which a `defaults` value cannot
  hold, shows `~/Library` and restarts Finder, the Dock and the other apps
  that read the settings; the `pmset` and `nvram` settings are in the task
  `upgrade:power` of `mise`."

## Stale comments

- macos/tools/fish/tools.toml:2-3 — says that `cli` adds drop-ins to
  `conf.d/` — `cli` adds none (macos/tools/cli/tools.toml has no fish file);
  `zoxide` and `dasel` add them (macos/tools/zoxide/tools.toml:16,
  macos/tools/dasel/tools.toml:12). Already item 2 of the parked record.
- macos/tools/homebrew/files/.config/fish/conf.d/05-homebrew.fish:16 — says
  that mise activates at "51-" — the mise drop-in is `06-mise.fish`
  (macos/tools/mise/tools.toml:52). Already item 1 of the parked record.
- macos/tools/homebrew/tasks/upgrade-brew.py:63 — the printed subheader names
  `[bootstrap.packages]` — the table is `[mise.bootstrap.packages]`
  (macos/tools/homebrew/tasks/upgrade-brew.py:53-54).
- macos/tools/homebrew/tasks/brew-casks.py:8,13-16 — says that the Brewfile
  holds casks and VS Code extensions — the Brewfile renders only `vscode`
  lines (macos/tools/homebrew/files/.config/brewfile:3-9), so
  `brew bundle cleanup --casks` (brew-casks.py:47) finds no cask in it.

## No change

- `duplicate-detective`, `easyfind`, `eza`, `final-cut-pro`, `fzf`,
  `git-tools`, `github-desktop`, `glab`, `google-chrome`, `gping`,
  `handbrake`, `httping`, `imagemagick`, `ipcalc`, `jq`, `keynote`, `kindle`,
  `lazygit`, `lcov`, `magnet`, `mas`
