# Audit part A — 1password to dprint

## Findings

### `1password`

- F1 missing prose — doc: "1Password (the app) as SSH agent and git commit
  signer." — evidence: macos/tools/1password/tools.toml:2,
  macos/tools/1password/tools.toml:12, macos/tools/1password/tools.toml:20,
  macos/tools/1password/tools.toml:27,
  macos/tools/1password/files/.config/1Password/ssh/agent.toml.tmpl:12-13,
  macos/tools/mise/files/.config/mise/config.toml:45 — new:

  ```markdown
  1Password (the app) as SSH agent and git commit signer. The SSH agent offers
  the keys of the vault `ssh_vault`. The cask installs into `/Applications`, not
  into `~/Applications` like the other casks, and the git signer
  `gpg.ssh.program` is `/Applications/1Password.app/Contents/MacOS/op-ssh-sign`.
  Only the detect of the question `ssh_vault` runs `op`, the command of the tool
  `1password-cli`.
  ```

### `bash`

- F2 missing prose — doc: "bash from Homebrew, its rc files and its drop-in
  folder ~/.config/bash/." — evidence: macos/tools/bash/files/.bash_profile:7,
  macos/tools/bash/files/.bashrc:9-14,
  macos/tools/bash/files/.config/bash/00-bootstrap.bash:7-8,
  macos/tools/bash/files/.config/bash/00-bootstrap.bash:26-27,
  macos/tools/bash/tools.toml:18-20 — new:

  ```markdown
  bash from Homebrew, its rc files and its drop-in folder ~/.config/bash/.
  `~/.bash_profile` loads `~/.bashrc`, and `~/.bashrc` loads each
  `~/.config/bash/*.bash` file in name order, so the drop-ins of other tools
  (`01a-homebrew.bash`, `01b-mise.bash`) run after `00-bootstrap.bash`.
  `00-bootstrap.bash` sets the XDG folders, `GPG_TTY` (mise cannot set it,
  because it needs the live terminal) and `FNOX_CONFIG_DIR`, and it adds
  `~/.local/bin` at the end of `PATH`. `/etc/shells` is outside `$HOME`, so the
  line of Homebrew's bash in it stays a `[mise.dotfiles]` declaration.
  ```

### `brave-browser`

- F3 wrong Mise sections — doc: "- **Mise sections:**
  `[mise.bootstrap.packages]`" — evidence:
  macos/tools/brave-browser/tools.toml:7-11 — new: "- **Mise sections:**
  `[mise.bootstrap.packages]`, `[mise.env]`"
- F4 missing prose — doc: "Brave Browser." — evidence:
  macos/tools/brave-browser/tools.toml:10-11 — new: "Brave Browser. Its
  `[mise.env]` sets `CHROME_EXECUTABLE` to the Brave app in `~/Applications`,
  so a tool that reads that variable starts Brave."

### `claude`

- F5 wrong Settings — doc: "- **Settings:** The merge spec adds only `$schema`,
  the install keys of `env` and, with `claude-status`, its status lines and
  hook, and the `PreToolUse` item of the npm hook (matcher `Bash`). The own keys
  of the user are in `settings` of `[tools.claude]`. `lists` maps a dotted path,
  for example `"permissions.allow"`, to the items that the sync adds to that
  list." — evidence:
  macos/tools/claude/files/.claude/settings.json.merge.toml:1-16,
  macos/tools/claude/files/.claude/settings.json.merge.toml:38-85 — new:

  ```markdown
  - **Settings:** Claude Code rewrites `~/.claude/settings.json`, so the merge
    spec adds only `$schema`, the install keys of `env`
    (`CLAUDE_INSTALL_METHOD`, `DISABLE_AUTOUPDATER`: Homebrew installs Claude
    Code, so it does not update itself), with `claude-status` its status lines
    and hook, and the `PreToolUse` item of the npm hook (matcher `Bash`). It
    removes `deniedMcpServers` and `env.CONTEXT7_RUNNER`, which bootstrap set
    before. A key that the user or Claude Code changes keeps its value. The own
    keys of the user are in `settings` of `[tools.claude]`. `lists` maps a
    dotted path, for example `"permissions.allow"`, to the items that the sync
    adds to that list, and an item that the user adds in Claude Code stays.
  ```

- F6 wrong upgrade:ai — doc: "- **upgrade:ai:** it calls the `claude` command
  of `claude-cli`. It updates each plugin of the user scope, but not a plugin of
  the `builtin` marketplace (`<name>@builtin`): that plugin ships inside Claude
  Code, and `claude plugin update` cannot find it." — evidence:
  macos/tools/claude/tasks/upgrade-ai.py:9-10,
  macos/tools/claude/tasks/upgrade-ai.py:58-82,
  macos/tools/mise/tasks/updateall.py:141 — new:

  ```markdown
  - **upgrade:ai:** a hidden task that `updateall` runs when it exists. It
    calls the `claude` command of `claude-cli`. It adds the marketplaces
    `anthropics/claude-plugins-official` and `virajp/claude-plugins` when they
    are missing, and updates each marketplace. It updates each plugin of the
    user scope, but not a plugin of the `builtin` marketplace
    (`<name>@builtin`): that plugin ships inside Claude Code, and
    `claude plugin update` cannot find it. Then it removes the unused plugins
    of the user scope (`claude plugin autoremove`).
  ```

### `claude-cli`

- F7 missing prose — doc: "Claude Code: its output styles, the func:claude task
  and the cc* aliases." — evidence: macos/tools/claude-cli/tools.toml:11-20,
  macos/tools/claude-cli/tasks/func-claude.py:13-17,
  macos/tools/claude-cli/tasks/func-claude.py:51-85 — new:

  ```markdown
  Claude Code: its output styles, the func:claude task and the cc* aliases.
  `func:claude` starts `claude` with a session named after the current folder,
  or `<folder>-<name>` with the argument `<name>`, and resumes that session
  unless `--new` is given. It adds `--remote-control` and the flags that skip
  the permission prompts. `cc` uses the default model, and `ccf`, `cch`, `cco`
  and `ccs` give `--model` with Fable, Haiku, Opus and Sonnet.
  ```

### `cli`

- F8 wrong prose — doc: "the projects folder, the generic aliases and the
  func:*, cleanupDS and keychain:delete tasks. Each command-line package is a
  tool of its own." — evidence: macos/tools/cli/tools.toml:1-5,
  macos/tools/cli/tools.toml:10-19,
  macos/tools/git/files/.config/git/identity.tmpl:5,
  macos/tools/mise/files/.config/fish/conf.d/06-mise.fish.tmpl:3 — new:

  ```markdown
  The projects folder, the generic aliases and the func:*, cleanupDS and
  keychain:delete tasks. Each command-line package is a tool of its own.
  `projects_dir` sets `PROJECTS_DIR`, with `~` expanded, because a value of
  `[mise.env]` does not expand it. The tools `git` and `mise` also read
  `projects_dir`. `CLOUD_PATH` is the iCloud Drive folder.
  ```

### `dasel`

- F9 missing prose — doc: "dasel, the query tool for JSON, YAML and TOML, and
  its fish completion." — evidence:
  macos/tools/dasel/files/.config/fish/conf.d/53-dasel.fish:4-6 — new: "dasel,
  the query tool for JSON, YAML and TOML, and its fish completion. The fish file
  loads the completion only when `dasel` is on PATH."

### `diff-so-fancy`

- F10 wrong prose — doc: "diff-so-fancy, the git diff formatter (the pager of
  the git config)." — evidence: macos/tools/diff-so-fancy/tools.toml:4,
  macos/tools/git/tools.toml:7, macos/tools/git/tools.toml:45-46,
  macos/tools/git/files/.config/git/config.tmpl (no `pager` key) — new:
  "diff-so-fancy, the git diff formatter. The tool `git` depends on it, and its
  alias `diff` runs `diff-so-fancy`. The git config sets no pager."

### `dprint`

- F11 order Contributes to — doc: "- **Defaults:** `settings`, `blocks`,
  `excludes`, `plugins`" before "- **Contributes to `visual-studio-code`:**
  `extensions`" — evidence: docs/setup/tools.md:41-66 (the legend puts
  Contributes to before Defaults) — new: put "- **Contributes to
  `visual-studio-code`:** `extensions`" before "- **Defaults:** `settings`,
  `blocks`, `excludes`, `plugins`".
- F12 wrong Plugins — doc: "- **Plugins:** The global config extends no other
  config. The ruff plugin (`npm:@dprint/ruff`, the block `ruff`) formats each
  `.py` file, also each Python task below `~/.config/mise/tasks/`. The shuck
  block formats the other task files, and skips `*.tmpl` and `*.py`. The sync
  finds a plugin by its name, so it keeps the version in the file. `updateall`
  runs `dprint config update` on the file, and that step updates the versions."
  — evidence: macos/tools/dprint/tools.toml:7-12,
  macos/tools/dprint/tools.toml:59-74, macos/tools/dprint/tools.toml:84-85,
  macos/tools/dprint/files/.config/dprint/dprint.jsonc.merge.toml:22-31,
  macos/tools/mise/tasks/updateall.py:101-115 — new:

  ```markdown
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
  ```

## Stale comments

- macos/tools/claude/files/.claude/CLAUDE.md.merge.toml:44-46 — says the
  aliases are in the `[shell_alias]` blocks of `~/.config/mise/conf.d/*.toml`,
  and the shadowing ones in `cli.toml`, `network.toml`, `git.toml` and
  `devtools.toml` — the sync writes `conf.d/<tool>/mise.toml`
  (macos/tools/bash/tools.toml:4), there is no `devtools` tool, and the
  shadowing aliases are in `eza` (macos/tools/eza/tools.toml:13-20), `zoxide`
  (macos/tools/zoxide/tools.toml:13), `git` (macos/tools/git/tools.toml:46),
  `prettyping` (macos/tools/prettyping/tools.toml:11), `doggo`
  (macos/tools/doggo/tools.toml:11), `sublime-text`
  (macos/tools/sublime-text/tools.toml:11) and `cli`
  (macos/tools/cli/tools.toml:28).
- macos/tools/claude/files/.claude/CLAUDE.md.merge.toml:66 — lists the claude
  aliases as `cc`, `cco`, `ccf` — `claude-cli` also has `cch` and `ccs`
  (macos/tools/claude-cli/tools.toml:16-20).
- macos/tools/claude-cli/tasks/func-claude.py:20 — the usage line gives
  `[claude args...]` — the task passes no other argument to `claude`; it builds
  the argument list from `<name>`, `--new` and `--model` only
  (macos/tools/claude-cli/tasks/func-claude.py:67-85).
- macos/tools/claude-cli/tasks/func-claude.py:52 — "Claude Code is the claude
  tool's, which is opt-in" — the tool `claude-cli` installs Claude Code
  (macos/tools/claude-cli/tools.toml:11-12).
- macos/tools/diff-so-fancy/tools.toml:4 — the description says "the pager of
  the git config" — the git config sets no pager; the `diff` alias of `git` runs
  diff-so-fancy (macos/tools/git/tools.toml:46).
- macos/tools/bash/files/.config/bash/00-bootstrap.bash:4-5 — says that fish
  and zsh pick up the `[env]` of mise — bash also activates mise, with the
  drop-in `01b-mise.bash` of the tool `mise`
  (macos/tools/mise/files/.config/bash/01b-mise.bash.tmpl:3).

## No change

- `1password-cli`, `1password-safari`, `age`, `alfred`, `android-studio`,
  `bat`, `battery-health`, `bitwarden`, `btop`, `cardhop`, `claude-desktop`,
  `claude-status`, `codeburn`, `coreutils`, `curl`, `daisydisk`, `developer`,
  `disk-doctor`, `disk-map`, `doggo`
