# Audit part D — starship to zsh

## Findings

### `starship`

- F1 missing prose — doc: "starship prompt, its config and the fish and zsh
  init drop-ins." — evidence:
  `macos/tools/starship/tools.toml:3`,
  `macos/tools/starship/files/.config/fish/conf.d/51-prompt.fish:4-6`,
  `macos/tools/starship/files/.config/zsh/02b-starship.zsh:4-7` — new:
  "starship prompt, its config and the fish and zsh init drop-ins. Each init
  runs only in an interactive shell, and only when `starship` is on PATH. bash
  gets no starship init."

### `superfile`

- F2 missing prose — doc: "superfile, the terminal file manager, and its
  config." — evidence: `macos/tools/superfile/tools.toml:1-2`,
  `macos/tools/superfile/tools.toml:9-11`,
  `macos/tools/superfile/files/Library/Application
  Support/superfile/config.toml.merge.toml:1-3`,
  `macos/tools/superfile/files/Library/Application
  Support/superfile/config.toml.merge.toml:14`,
  `macos/tools/superfile/files/Library/Application
  Support/superfile/config.toml.merge.toml:20-26`
  — new:
  "superfile, the terminal file manager, and its config. superfile rewrites
  its config, so the merge spec adds only the keys of bootstrap, and superfile
  owns each other key. The spec owns the list `sidebar_sections` as a block,
  so its order holds. The config previews code with bat and uses zoxide, so
  the tool depends on `bat` and `zoxide`."

### `tailscale`

- F3 missing prose — doc: "Tailscale (the formula; tailscaled runs as a root
  brew service) and its daemon and upgrade tasks." — evidence:
  `macos/tools/tailscale/tasks/tailscale-daemon.py:13-17`,
  `macos/tools/tailscale/tasks/tailscale-daemon.py:29-33`,
  `macos/tools/tailscale/tasks/upgrade-tailscale.py:13-16`,
  `macos/tools/tailscale/tasks/upgrade-tailscale.py:77-96`,
  `macos/tools/mise/tasks/updateall.py:163-164`,
  `macos/tools/homebrew/tasks/upgrade-brew.py:57-59` — new:
  "Tailscale (the formula; tailscaled runs as a root brew service) and its
  daemon and upgrade tasks. mise installs the formula, and only the daemon
  goes through `brew services`, because tailscaled needs root for the utun
  interface and mise has no privileged services on macOS. The App Store build
  is sandboxed and cannot run the Tailscale SSH server. `updateall` runs
  `tailscale:daemon`, which starts the daemon only when tailscaled does not
  run, because `sudo brew services start` restarts a running daemon and drops
  the tunnel. `upgrade:brew` runs `upgrade:tailscale` first: it gives the
  root-owned keg back to the user, so that `brew cleanup` can remove it, and
  upgrades the formula with `mise bootstrap packages upgrade`."

### `telegram`

- F4 missing prose — doc: "Telegram, from the Mac App Store." — evidence:
  `macos/tools/telegram/tools.toml:11-12`,
  `macos/tools/telegram/tasks/cleanupTelegram.py:21`,
  `macos/tools/telegram/tasks/cleanupTelegram.py:38-45` — new:
  "Telegram, from the Mac App Store. The task `cleanupTelegram` (also the
  shell alias `cleanupTelegram`) quits Telegram, deletes the `telegram-*`
  cache files larger than 1 MB in its App Store group container, and starts
  Telegram again."

### `touchid-sudo`

- F5 missing prose — doc: "Touch ID for sudo (/etc/pam.d/sudo_local)." —
  evidence: `macos/tools/touchid-sudo/tools.toml:2-3`,
  `macos/tools/touchid-sudo/tools.toml:9-17` — new:
  "Touch ID for sudo (/etc/pam.d/sudo_local). Each macOS update writes
  `/etc/pam.d/sudo` again, but `sudo_local` is the drop-in that it includes,
  and it stays. The target is outside `$HOME`, so it is a
  `[mise.bootstrap.files]` entry (mode `0444`, owner `root`, group `wheel`)
  that `mise bootstrap` applies."

### `uv`

- F6 missing prose — doc: "uv through mise, with uv as the pipx backend. Each
  Mac task is a uv script, so each Mac must have uv." — evidence:
  `macos/tools/uv/tools.toml:7-9`, `macos/tools/uv/tools.toml:14-17` — new:
  "uv through mise, with uv as the pipx backend. Each Mac task is a uv script,
  so each Mac must have uv. Each `pipx:` tool installs through uv
  (`pipx.uvx`). `UV_NO_MANAGED_PYTHON` and `UV_PYTHON_DOWNLOADS = "never"`
  stop uv from downloading a Python of its own."

### `visual-studio-code`

- F7 missing Defaults — doc: "**Defaults:** `extensions`, the VS Code
  extensions of `~/.config/brewfile` (homebrew). Each tool adds its own." —
  evidence: `macos/tools/visual-studio-code/tools.toml:7-11`,
  `macos/tools/homebrew/files/.config/brewfile:3-6`,
  `macos/tools/homebrew/tasks/brew-casks.py:8` — new:
  "**Defaults:** `extensions`, the VS Code extensions of `~/.config/brewfile`
  (homebrew), which the task `brew:casks` installs. Each selected tool adds
  its own in `[contributes.visual-studio-code]`, and `extensions` of
  `[tools.visual-studio-code]` in `macos.toml` adds more."

### `vorssaint`

- F8 missing prose — doc: "Menu bar toolkit with keep-awake, system monitor
  and volume mixer." — evidence: `macos/tools/vorssaint/tools.toml:11-15` —
  new:
  "Menu bar toolkit with keep-awake, system monitor and volume mixer. The
  defaults leave out the audio device priorities, the state of one Mac (mic
  mute devices, update checks, window sizes), the onboarding and migration
  markers, and the binary values (`radialMenuProfiles`, `shelfItems`) that
  mise cannot write. Restart the app after `mise bootstrap` applies the
  defaults."

### `zoxide`

- F9 missing prose — doc: "zoxide, the fast directory jumper, its shell init
  and the cd alias." — evidence: `macos/tools/zoxide/tools.toml:12-13`,
  `macos/tools/zoxide/files/.config/zsh/02a-zoxide.zsh:4-7`,
  `macos/tools/zoxide/files/.config/fish/conf.d/52-zoxide.fish:4-5`,
  `macos/tools/starship/files/.config/zsh/02b-starship.zsh:1` — new:
  "zoxide, the fast directory jumper, its shell init and the cd alias. The
  alias makes `cd` run `z`. The zsh init runs only in an interactive shell,
  and loads before `02b-starship.zsh` of `starship`."

### `zsh`

- F10 missing prose — doc: "zsh from Homebrew with zinit, its rc files and its
  drop-in folder ~/.config/zsh/." — evidence:
  `macos/tools/zsh/tools.toml:3-6`, `macos/tools/zsh/tools.toml:17-19`,
  `macos/tools/zsh/files/.zshenv:10-14`, `macos/tools/zsh/files/.zshrc:10-14`,
  `macos/tools/zsh/files/.config/zsh/00-bootstrap.zsh:3-8`,
  `macos/tools/zsh/files/.config/zsh/98-zinit.zsh:11` — new:
  "zsh from Homebrew with zinit, its rc files and its drop-in folder
  ~/.config/zsh/. `~/.zshenv` and `~/.zshrc` each source each
  `~/.config/zsh/*.zsh` file in name order, so an interactive shell sources
  each drop-in two times. Other tools add their drop-ins to this folder
  (`homebrew`, `mise`, `zoxide`, `starship`, `orbstack`), and `~/.zprofile`
  belongs to the tool `mise`. `00-bootstrap.zsh` holds only the bootstrap
  variables, because each other variable is in the `[env]` of mise;
  `GPG_TTY` is here because mise evaluates its `[env]` in a subprocess, which
  has no terminal. The fzf-tab plugin of `98-zinit.zsh` needs fzf, so the tool
  depends on `fzf`. `/etc/shells` is outside `$HOME`, so `[mise.dotfiles]`
  adds `/opt/homebrew/bin/zsh` to it."

## Stale comments

- `macos/tools/starship/tools.toml:2-3` — says that `02a-zoxide.zsh` is of
  `cli` — the tool `zoxide` ships it (`macos/tools/zoxide/tools.toml:19`).
- `macos/tools/zsh/tools.toml:2-3` — says that `homebrew`, `mise`, `cli` and
  `starship` add drop-ins to `~/.config/zsh/` — `cli` adds none; `zoxide`
  (`macos/tools/zoxide/tools.toml:19`) and `orbstack`
  (`macos/tools/orbstack/files/.config/zsh/90-orbstack.zsh`) add them.
- `macos/tools/zsh/files/.zshenv:4-5` — says to keep the file minimal, with
  only what non-interactive shells need — the file sources each drop-in of
  `~/.config/zsh/` (`macos/tools/zsh/files/.zshenv:10-14`), and `~/.zshrc`
  sources them again (`macos/tools/zsh/files/.zshrc:10-14`).
- `macos/tools/touchid-sudo/tools.toml:2` — names a `[bootstrap.files]` entry
  — the table is `[mise.bootstrap.files]`
  (`macos/tools/touchid-sudo/tools.toml:11`).
- `macos/tools/tailscale/tasks/upgrade-tailscale.py:13` — names
  `[bootstrap.packages]` — the table is `[mise.bootstrap.packages]`
  (`macos/tools/tailscale/tools.toml:7`).
- `macos/tools/superfile/tools.toml:9-11` — the comment of the merge spec is
  above `[mise.bootstrap.packages]` — it is about the `[files]` entry
  (`macos/tools/superfile/tools.toml:16-18`).
- `macos/tools/yq/tools.toml:1` — says "the YAML, JSON and TOML processor" —
  the description says "the YAML processor"
  (`macos/tools/yq/tools.toml:4`).

## No change

- `sublime-text`, `taskexplorer`, `testflight`, `the-unarchiver`,
  `transporter`, `tree`, `vlc`, `warp`, `watch`, `wget`, `whatsapp`,
  `worktrunk`, `xcode`, `yq`, `zed`, `zed-delta`
