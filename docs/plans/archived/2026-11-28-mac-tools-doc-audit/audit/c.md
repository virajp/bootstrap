# Audit part C — microsoft-excel to ssh

## Findings

### `mise`

- F1 missing prose — doc: "mise: its user-level config, its activation in
  bash, zsh and fish, ~/.zprofile, the entry tasks of the Mac sync and the
  Claude skill `mac-setup` of the backup clone." — evidence:
  macos/tools/mise/tools.toml:3-7,
  macos/tools/mise/files/.config/bash/01b-mise.bash.tmpl:3,12-22,
  macos/tools/mise/files/.config/zsh/01b-mise.zsh.tmpl:3,8-18,
  macos/tools/mise/files/.config/fish/conf.d/06-mise.fish.tmpl:7,11-22,
  macos/tools/mise/files/.zprofile:3-4,
  macos/tools/mise/files/.config/mise/config.toml:43-45,
  docs/design/mac.md:15,42-43 — new:

  ```markdown
  mise: its user-level config, its activation in bash, zsh and fish,
  ~/.zprofile, the entry tasks of the Mac sync and the Claude skill `mac-setup`
  of the backup clone. The activation files read `projects_dir` of the `cli`
  tool, so the tool depends on `cli`. The entry tasks are Python uv scripts, so
  the tool also depends on `python` and `uv`: both are then in the init stage,
  and the first setup installs them before its exit 4. `~/.zprofile` puts
  `~/.local/bin`, the folder of mise, on PATH before the activation files run.
  Each activation file exports `MISE_ENV=dev`. An interactive shell gets the
  full `mise activate`, and a non-interactive shell gets
  `mise activate --shims`. The fish file sets `MISE_FISH_AUTO_ACTIVATE` to 0,
  because the Homebrew formula of mise ships
  `vendor_conf.d/mise-activate.fish`, which sorts after it and runs a full
  `mise activate` again. `MISE_BREW_CASK_OPT_APPDIR` puts each cask app in
  `~/Applications`. The place of `macos.toml` is fixed, at
  `~/.config/bootstrap/macos.toml`, so the tool has no question `config_path`:
  `mac migrate` moves an old answer to the key `backup`. See
  [mac.md](../design/mac.md#the-place-of-macostoml).
  ```

- F2 wrong The place of `macos.toml` — doc: "**The place of `macos.toml`:**
  fixed, at `~/.config/bootstrap/macos.toml`. The old question `config_path`
  is gone: `mac migrate` moves its answer to the key `backup`. See
  [mac.md](../design/mac.md#the-place-of-macostoml)." — evidence:
  macos/tools/mise/tools.toml:1-108 (no file, task or key of the tool holds the
  place), docs/design/mac.md:15,42-43 — new: "" (delete the item; F1 moves its
  text into the prose of this section. By D16, the kind occurs in this section
  only and names no part of the tool, so its text moves into the prose as a
  finding of kind wrong).
- F3 wrong Python tasks — doc: "**Python tasks:** each Mac task is a Python uv
  script with a `.py` name, … It runs under `mise run` and directly, with uv on
  PATH." (docs/setup/tools.md:848-865, the whole item) — evidence:
  macos/tools/\*/tasks/\* (each of the 23 tasks has
  `# MISE tools={uv="latest", python="3.14"}`; only
  macos/tools/claude-cli/tasks/func-claude.py,
  macos/tools/cli/tasks/func-code.py, macos/tools/cli/tasks/keychain-delete.py
  and macos/tools/mise/tasks/updateall-configs.py do not import `helpers`),
  macos/tools/mise/files/.config/mise/config.toml:28,
  macos/tools/mise/tasks/updateall-configs.py:19-20 — new: delete the item from
  the `mise` section (F1 keeps the reason for `depends` in the prose). D3
  governs the place, not the prose rule of D16: the item kind occurs in this
  section only, but its text is a rule for the tasks of each Mac tool, not a
  fact of the tool `mise`. D3 puts a rule for many tools in "How a tool works",
  as a paragraph before "Each section below gives these items for one tool:".
  That paragraph is also the place of G9 of audit/general.md, so U6 writes one
  paragraph with the facts of both, with no repeated sentence (D14). The text
  of this finding:

  ````markdown
  Each task of a Mac tool is a Python uv script with a `.py` name, and mise
  removes `.py` from the task name. Each task has
  `# MISE tools={uv="latest", python="3.14"}`, and the Mac config of the tool
  `mise` sets `task.run_auto_install = false`, so `mise` depends on `python`
  and `uv`. A task that uses a function of the library `helpers.py` of the tool
  `mise` (the print functions of the Mac tasks, `task_exists` and
  `run_task_if_exists`) imports it with this preamble:

  ```python
  sys.dont_write_bytecode = True
  sys.path.insert(0, os.path.expanduser("~/.config/mise/tasks/_scripts"))
  import helpers  # noqa: E402
  ```

  `func:claude`, `func:code` and `keychain:delete` use no function of the
  library, so they do not import it. `updateall:configs` imports no library, so
  it runs before the first sync writes `helpers.py`. It runs under `mise run`
  and directly, with uv on PATH.
  ````

- F4 wrong Skill `mac-setup` — doc: "**Skill `mac-setup`:** to install or
  remove a tool or an app, or to change a setting, edit `macos.toml` and run
  `mise run updateall:configs`. Never run `brew install`, `mise use -g` or
  `defaults write` directly. The dev tools come from mise, the others from
  Homebrew. A file with the `Synced from bootstrap` header is changed in
  `bootstrap`. `<backup>` is the backup clone (`macos-setup`). Without
  `backup`, the sync skips the skill." — evidence:
  macos/tools/mise/files/.claude/skills/mac-setup/SKILL.md.tmpl:12-24,
  macos/tools/mise/tools.toml:18-21,69-87, docs/design/mac.md:377-378 — new:

  ```markdown
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
  ```

- F5 wrong Entry of the Mac sync — doc: "**Entry of the Mac sync:**
  `updateall:configs` runs `bootstrap mac sync`, with the flags `--dry-run`,
  `--no-bootstrap` and `--verbose`. `updateall` upgrades the Mac, and runs
  `updateall:configs` as one step. With the `dprint` tool, it also updates the
  plugins of `~/.config/dprint/dprint.jsonc`. The `upgrade:*` tasks download
  and install the macOS updates and apply the power profile. The aliases are
  `updateall` and `osx-upgrade`." — evidence:
  macos/tools/mise/tools.toml:39-45,
  macos/tools/mise/tasks/updateall-configs.py:11-17,155-164,
  macos/tools/mise/tasks/updateall.py:14,27-31,81-88,101-115,143-155 — new:

  ```markdown
  - **Entry of the Mac sync:** `updateall:configs` runs `bootstrap mac sync`,
    with the flags `--dry-run`, `--no-bootstrap` and `--verbose`, and exits
    with its code. `BOOTSTRAP_BIN`, a local build, replaces the `bootstrap` of
    mise. `updateall` upgrades the Mac, and runs `updateall:configs` as one
    step. An exit 4 of that step stops `updateall` with exit 4: open a new
    terminal, then run `updateall` again. With the `dprint` tool, `updateall`
    also updates the plugins of `~/.config/dprint/dprint.jsonc`. Its weekly
    steps (the macOS settings and `upgrade:power`) run last, because they ask
    for sudo and restart Finder, Dock and ControlCenter. They run when the
    stamp `/tmp/mise-updateall-weekly.stamp` is missing or older than a week,
    or with `--force`. macOS clears `/tmp` on a reboot, and a macOS update
    always reboots, so the first `updateall` after an update applies the
    settings again. The `upgrade:*` tasks download and install the macOS
    updates and apply the power profile. The aliases are `updateall`,
    `osx-upgrade`, `cancelPrintJobs` and `listServices`.
  ```

- F6 order Contributes to — doc: "- **Defaults:** `skill.lines`, an empty
  table. `homebrew`, `mas`, `setapp` and `macos-defaults` add their lines."
  before "- **Contributes to `visual-studio-code`:** `extensions`" —
  evidence: docs/setup/tools.md:58-61 (the legend puts Contributes to before
  Defaults), macos/tools/mise/tools.toml:21-24 — new: put "- **Contributes to
  `visual-studio-code`:** `extensions`" before "- **Defaults:**
  `skill.lines`, an empty table. `homebrew`, `mas`, `setapp` and
  `macos-defaults` add their lines.", text unchanged. The item Skill
  `mac-setup` (F4) gets no order finding: the legend does not name its kind
  today, so U6 puts it by the legend of U5 (G8 of audit/general.md).

### `node`

- F7 missing prose — doc: "Node.js (LTS), and the user config of the aube that
  mise embeds. The tool has no pnpm and no `PNPM_HOME`, because a repository
  selects its own package manager." — evidence:
  macos/tools/node/tools.toml:16,
  macos/tools/node/files/.config/aube/config.toml:3-7,
  macos/tools/node/tasks/upgrade-npm.py:37-47 — new: "Node.js (LTS), and the
  user config of the aube that mise embeds. The tool has no pnpm and no
  `PNPM_HOME`, because a repository selects its own package manager. mise
  installs the `npm:` tools with its embedded aube (`npm.package_manager`).
  That aube reads `~/.config/aube/config.toml`, not the `[env]` of mise, so
  the file holds `minimumPackageAge`: 600 minutes (10 hours) for a new package
  name. `upgrade:npm` sets `fund=false` and upgrades the global npm packages,
  and it does nothing when npm is not on PATH."

### `openwhispr`

- F8 missing prose — doc: "Voice-to-text dictation with AI agents, and its
  hotkey." — evidence: macos/tools/openwhispr/tools.toml:7-10,15-17 — new:
  "Voice-to-text dictation with AI agents, and its hotkey. OpenWhispr writes
  `.env` again at each start, so the file is a seed: the sync writes it once,
  and then the app owns it. The app reads only `DICTATION_KEY` and
  `ACTIVATION_MODE` from it, on a new install."

### `orbstack`

- F9 missing prose — doc: "OrbStack (Docker and Kubernetes), its env, the
  docker and kubectl aliases, its fish and zsh init and its ssh Include." —
  evidence: macos/tools/orbstack/tools.toml:7-10,20,31-32 — new: "OrbStack
  (Docker and Kubernetes), its env, the docker and kubectl aliases, its fish
  and zsh init and its ssh Include. The Include `~/.orbstack/ssh/config` gives
  the `orb` SSH host, and the `ssh` tool puts it at the top of
  `~/.ssh/config`, before each Host block. `DOCKER_DEFAULT_PLATFORM` is
  `linux/amd64`, and the aliases `dpull` and `drun` also force that platform."

### `python`

- F10 wrong Mise sections — doc: "**Mise sections:** `[mise.settings]`,
  `[mise.tools]`, `[mise.env]`" — evidence: macos/tools/python/tools.toml:7-12
  (no `[mise.env]` table) — new: "**Mise sections:** `[mise.settings]`,
  `[mise.tools]`"
- F11 missing prose — doc: "Python through mise. uv is the tool `uv`." —
  evidence: macos/tools/python/tools.toml:9,11-12 — new: "Python through mise.
  uv is the tool `uv`. mise installs Python 3.14, the version of the Mac tasks,
  and does not compile it (`python.compile = false`)."

### `ruby`

- F12 missing prose — doc: "Ruby through mise, and the RubyGems config
  (~/.gemrc): no documentation on install and update. mise uses Ruby to run the
  hooks of some casks and to build a formula without a bottle." — evidence:
  macos/tools/ruby/tools.toml:12-19, macos/tools/ruby/files/gemrc.merge.toml:1-8
  — new: "Ruby through mise, and the RubyGems config (`~/.gemrc`): no
  documentation on install and update. mise uses Ruby to run the hooks of some
  casks and to build a formula without a bottle. gem writes `~/.gemrc` again
  (the gem sources), so the file is a merge: the spec adds only its keys, and
  sets the format `yaml` because the file has no extension."

### `ssh`

- F13 missing prose — doc: "the ssh client config: Host \* defaults, the default
  user, the includes and the agent that other tools contribute." — evidence:
  macos/tools/ssh/tools.toml:1-4,24-25,
  macos/tools/ssh/files/.ssh/config.tmpl:3-18,29-31,
  macos/tools/orbstack/tools.toml:9-10, macos/tools/1password/tools.toml:17 —
  new: "the ssh client config: Host \* defaults, the default user, the includes
  and the agent that other tools contribute. An Include works only at the top
  of `~/.ssh/config`, before each Host block, so the template puts the
  `includes` of other tools there, for example the one of `orbstack`. The
  `identity_agent` of another tool, for example `1password`, goes inside
  `Host *`. A host or a key of this Mac only goes in `~/.ssh/config.local`,
  which the sync does not write. ssh ignores the Include when that file is
  missing. ssh refuses a `~/.ssh` folder that other users can read, so
  `dir_perm` keeps it at mode 700."
- F14 wrong Folder mode — doc: "**Folder mode:** `dir_perm` keeps the folder
  `~/.ssh` at mode 700." — evidence: macos/tools/ssh/tools.toml:24-34 (the
  Files item already gives `dir_perm` `700` for both targets) — new: "" (delete
  the item; F13 moves the reason into the prose of this section. By D16, the
  kind occurs in this section only and the Files item already names the key,
  so its text moves into the prose as a finding of kind wrong).

## Stale comments

- macos/tools/mise/tasks/updateall.py:74-75 — "mise manages its own formula
  ([bootstrap.packages])" — no tool of macos/tools/ lists a mise package (a
  grep of `brew:mise` finds none); the step runs `mise self-update`
  (macos/tools/mise/tasks/updateall.py:77), and mise is in `~/.local/bin`
  (macos/tools/mise/files/.zprofile:3).
- macos/tools/mise/tasks/updateall.py:143-146 — "`files` re-checks
  /etc/pam.d/sudo_local (Touch ID for sudo), `defaults` the macOS settings,
  `task` the pmset/nvram power profile" — the weekly step runs only
  `mise bootstrap --only macos-defaults,defaults`
  (macos/tools/mise/tasks/updateall.py:150); the power profile is the task
  `upgrade:power` (macos/tools/mise/tasks/updateall.py:152), and the files
  step runs earlier, in `mise bootstrap --skip macos-defaults,packages`
  (macos/tools/mise/tasks/updateall.py:129).
- macos/tools/mise/tasks/upgrade-power.py:13-14 — "Everything `defaults`-based
  lives in [bootstrap.macos.defaults] (conf.d/macos-defaults.toml)" — the sync
  generates `~/.config/mise/conf.d/macos-defaults/mise.toml`
  (macos/tools/macos-defaults/tools.toml:4-5).

## No change

- `microsoft-excel`, `microsoft-powerpoint`, `microsoft-word`, `network`,
  `notion`, `numbers`, `obsidian`, `opencode`, `orca`, `pages`,
  `pdf-protector`, `pdf-squeezer`, `pre-commit`, `prettyping`, `pstree`,
  `ripgrep`, `setapp`, `sill`
