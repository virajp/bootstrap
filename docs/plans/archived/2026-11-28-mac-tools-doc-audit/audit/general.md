# Audit — the general sections and the checklist of the old Notes

## Findings

### How a tool works

- G1 wrong — doc: "`macos.toml`, at the root of the private `macos-setup`
  clone, selects the tools:" (line 24) — evidence: `docs/design/mac.md:15-21`,
  `docs/design/mac.md:142-158`, `docs/setup/tools.md:878` (the `mise` section
  already gives the fixed place) — new:

  ```markdown
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
  ```

- G2 wrong — doc: "**Stage:** `migrate` for a tool of the init stage: the first
  setup of `bootstrap mac migrate` writes and installs it." (line 45) —
  evidence: `docs/design/tools.md:57`, `macos/tools/bash/tools.toml:8`,
  `macos/tools/fish/tools.toml:13`, `macos/tools/zsh/tools.toml:10` — new:
  "- **Stage:** `migrate` for a tool with `stage = "init"`, a tool of the init
  stage: the first setup of `bootstrap mac migrate` writes and installs it. A
  tool without this item has no stage of its own. See
  [`docs/design/mac.md`](../design/mac.md)."
- G3 wrong — doc: "The `format` and `dir_perm` keys of the `[files]` table
  follow." (lines 52-53) — evidence: `docs/design/tools.md:222-226`,
  `macos/tools/claude/tools.toml:38` (`perm`), `macos/tools/fnox/tools.toml:26`
  (`when`), `macos/tools/mise/tools.toml:74` (`root`),
  `macos/tools/ssh/tools.toml:27-28` (`dir_perm`, `format`) — new:
  "The keys `format`, `perm`, `dir_perm`, `when` and `root` of the `[files]`
  table follow. `<backup>/` starts the target of a `root = "backup"` entry: a
  real file in the backup clone, the key `backup` of `macos.toml`."
- G4 missing — doc: "- **Contributes to:** the keys that the tool adds to the
  values of another tool." (lines 58-59) — evidence: 17 sections use the form
  with a tool name, for example `docs/setup/tools.md:100`,
  `docs/setup/tools.md:669`; `docs/design/tools.md:453-470`;
  `macos/tools/1password/tools.toml:16-24` — new (in place of the old line):
  "- **Contributes to `<tool>`:** the keys that the tool adds to the values of
  the tool `<tool>` (`[contributes.<tool>]`), for example `extensions` of
  `visual-studio-code`, or its line of `skill.lines`, a line of the skill of
  `<tool>`. The section has one item for each tool."
- G5 missing — doc: "- **Question:** the key of a value that the owner of the
  Mac gives. The text of the question is only in the `[questions.<key>]` table
  of the `tools.toml` that the item links to. `bootstrap mac questions` shows
  it." (lines 64-66) — evidence: 8 sections use the form with a key, for
  example `docs/setup/tools.md:104`, `docs/setup/tools.md:582-589`;
  `macos/tools/git/tools.toml:36-39` — new (in place of the old line):
  "- **Question `<key>`:** a value that the owner of the Mac gives, with a link
  to the `tools.toml` that holds its `[questions.<key>]` table. The text of the
  question is only in that table. `bootstrap mac questions` shows it. The
  section has one item for each question."
- G6 missing — doc: "none" — evidence: the item `Runtime:` in 4 sections
  (`docs/setup/tools.md:231`, `:471`, `:534`, `:1090`); `runtime` of
  `macos/tools/claude/tools.toml:13`, `fish/tools.toml:12`,
  `gh/tools.toml:9`, `ssh/tools.toml:11`; `docs/design/tools.md:58`;
  `docs/design/mac.md:383-397` — new (a legend line after Mise sections):
  "- **Runtime:** the `runtime` paths of `tools.toml`: the files and folders
  below `$HOME` that the program of the tool writes. They must not go into
  git, so the backup clone does not link a folder that holds one: it links the
  synced items of that folder one by one. See
  [`docs/design/mac.md`](../design/mac.md#the-backup-clone)."
- G7 missing — doc: "none" — evidence: the item `Skill <skill>:` in 3 sections
  (`docs/setup/tools.md:493`, `:575`, `:829`); `docs/design/tools.md:563-600`;
  `macos/tools/fnox/tools.toml:26`, `macos/tools/git/tools.toml:84-97`,
  `macos/tools/mise/tools.toml:72-87` — new (a legend line after Runtime,
  before Question `<key>`):
  "- **Skill `<skill>`:** the Claude Code skill that the tool owns, in short:
  its rules, and where the sync writes it. Each `[files]` entry of the skill
  has `when = "claude.skills"`. See
  [`docs/design/tools.md`](../design/tools.md#skills)."
- G8 order — doc: the legend order "Packs, Stage, Depends, Files, Tasks,
  Scripts, Contributes to, Defaults, Mise sections, Question" (lines 43-66) —
  evidence: D16 of `index.json`; G4 to G7 — new: the legend order is "Packs,
  Stage, Depends, Files, Tasks, Scripts, Contributes to `<tool>`, Defaults,
  Mise sections, Runtime, Skill `<skill>`, Question `<key>`". Today these
  sections give their items in another order: `claude` (Runtime after Tasks),
  `dprint` (Defaults before Contributes to), `fish` (Runtime after Files, and
  Defaults before Contributes to), `fnox`
  (Skill before Mise sections), `gh` (Runtime before Contributes to), `git`
  (Skill before Mise sections), `mise` (Defaults and Skill before Contributes
  to) and `ssh` (Runtime before Contributes to). The part units own these
  sections; U6 applies the one order.
- G9 missing — doc: "none" — evidence: C1, C2, C5, C6 and C7 below — new: one
  paragraph before "Each section below gives these items for one tool:" (D3),
  with the texts of C1, C2, C5, C6 and C7 in that order.

### The packs

none

## Checklist

- C1 stated — section: How a tool works — evidence:
  `macos/tools/developer/tools.toml:8-9`,
  `macos/tools/claude-cli/tools.toml:11-12`,
  `macos/tools/mise/tools.toml:26-28`, `docs/design/cli.md:33-35` — new:
  "`[mise.bootstrap.packages]` of a `tools.toml` lists the Homebrew formulae
  (`brew:<name>`), the casks (`brew-cask:<name>`) and the Mac App Store apps
  (`mas:<adam id>`). A dev tool that mise installs goes in `[mise.tools]`, for
  example `usage` of `mise`."
- C2 stated — section: How a tool works — evidence:
  `macos/tools/homebrew/tasks/upgrade-brew.py:53-55,63-68` — new: "mise, not a
  Brewfile, installs these packages. `upgrade:brew` upgrades them with
  `mise bootstrap packages upgrade` and `mise bootstrap packages apply`. The
  Brewfile holds only the VS Code extensions." — dropped part: "into
  `/opt/homebrew`; `brew list` and `brew upgrade` still see them" — no file of
  `macos/tools/` states it.
- C3 stated — section: `mise`, `homebrew`, `1password` — evidence:
  `macos/tools/mise/files/.config/mise/config.toml:43-45`,
  `macos/tools/homebrew/tools.toml:19`, `macos/tools/1password/tools.toml:20,27`
  — new: in `mise`: "`MISE_BREW_CASK_OPT_APPDIR` of `~/.config/mise/config.toml`
  puts each cask that mise installs into `~/Applications`." In `homebrew`:
  "`HOMEBREW_CASK_OPTS` gives `brew` the same folder, `~/Applications`." In
  `1password`: "The cask goes to `/Applications` (`appdir` of its package),
  and the git signer is
  `/Applications/1Password.app/Contents/MacOS/op-ssh-sign`."
- C4 stated — section: `claude-status` — evidence:
  `macos/tools/claude-status/tools.toml:9-13` — new: "The tool declares the
  tap `virajp/tap` in `[mise.bootstrap.brew.taps]`, and its package is
  `brew:virajp/tap/claude-status`." — dropped part: "the formula of a tap
  installs only when the tap publishes `api/formula/<name>.json`" — no file of
  `macos/tools/` states it.
- C5 stated — section: How a tool works — evidence: each tool with a `mas:`
  package (24 tools) has `depends = [ "mas" ]`, for example
  `macos/tools/developer/tools.toml:4,9` — new: "Each tool with a `mas:`
  package depends on `mas`."
- C6 stated — section: How a tool works — evidence:
  `macos/tools/fonts/tools.toml:7-13`, `macos/tools/zsh/tools.toml:13-15`,
  `macos/tools/mise/tools.toml:26-37`; each other tool has at most one entry in
  `[mise.bootstrap.packages]` and `[mise.tools]` — new: "A tool holds one app
  or one package. The exceptions are `fonts` (the Nerd Fonts casks), `zsh`
  (zsh and zinit) and `mise` (`usage`, `github:virajp/bootstrap` and
  `ca-certificates`)." — dropped part: "unless one depends on the other" — no
  file of `macos/tools/` states that rule, and no tool with two packages has a
  package that depends on the other.
- C7 stated — section: How a tool works — evidence:
  `macos/tools/bash/files/.bashrc:7-14`, `macos/tools/zsh/files/.zshrc:10-14`,
  `macos/tools/fish/files/.config/fish/conf.d/02-path.fish:8` — new: "The
  shells load the drop-in files of `~/.config/bash/`, `~/.config/zsh/` and
  `~/.config/fish/conf.d/` in the order of their names. A tool adds its own
  drop-in, for example `01a-homebrew.zsh` of `homebrew` and `01b-mise.zsh` of
  `mise`."
- C8 stated — section: `mise` — evidence:
  `macos/tools/mise/files/.config/zsh/01b-mise.zsh.tmpl:9`,
  `macos/tools/mise/files/.config/bash/01b-mise.bash.tmpl:13`,
  `macos/tools/homebrew/tools.toml:34-45`, `macos/tools/mise/tools.toml:48-64`,
  `macos/tools/fish/files/.config/fish/conf.d/02-path.fish:14-18`,
  `macos/tools/homebrew/files/.config/fish/conf.d/05-homebrew.fish:16-17` (its
  name `51-` is stale: the file is `06-mise.fish`) — new: "The activation of
  mise runs after `brew shellenv` and the `PATH` exports: `01b-mise` sorts
  after `01a-homebrew` in bash and zsh, and `06-mise.fish` after
  `02-path.fish` and `05-homebrew.fish` in fish. mise prepends its tool paths
  to `PATH`, so its tools win, and a later change of `PATH` takes that place
  back." — dropped part: "a tool of both (`jq`, `yq`) resolves to the
  Homebrew copy" — no file of `macos/tools/` names such a tool.
- C9 stated — section: `mise` — evidence:
  `macos/tools/mise/files/.config/fish/conf.d/06-mise.fish.tmpl:17-23`,
  `01b-mise.zsh.tmpl:10-19`, `01b-mise.bash.tmpl:14-23` — new: "An
  interactive shell gets a full `mise activate`. Another shell gets
  `mise activate --shims`."
- C10 stated — section: `mise` — evidence:
  `macos/tools/mise/files/.config/fish/conf.d/06-mise.fish.tmpl:11-15` — new:
  "`06-mise.fish` sets `MISE_FISH_AUTO_ACTIVATE` to `0`. Thus the
  `vendor_conf.d` file of the Homebrew mise formula does not activate mise
  again after it."
- C11 stated — section: `fish`, `zsh`, `bash` — evidence:
  `macos/tools/fish/files/.config/fish/conf.d/02-path.fish:14-18,29-30`,
  `macos/tools/zsh/files/.config/zsh/00-bootstrap.zsh:25-26`,
  `macos/tools/bash/files/.config/bash/00-bootstrap.bash:26-27`,
  `macos/tools/mise/files/.zprofile:3-4` — new: in `fish`: "`02-path.fish` puts
  `~/.local/bin` on `PATH`, not `[env] _.path` of mise. mise puts `_.path`
  before its own tool paths, so the standalone copies in `~/.local/bin` (`uv`
  and `uvx`) then come before the tools of mise." In `zsh`: "`00-bootstrap.zsh`
  adds `~/.local/bin` at the end of `PATH`, in the shell config and not in
  `[env] _.path` of mise." In `bash`: "`00-bootstrap.bash` adds `~/.local/bin`
  at the end of `PATH`, in the shell config and not in `[env] _.path` of
  mise."
- C12 stated — section: `fish` — evidence:
  `macos/tools/fish/files/.config/fish/conf.d/02-path.fish:20-22,29-30` — new:
  "`02-path.fish` uses `fish_add_path --global`. A universal path stays in
  `fish_variables` after its line is removed, and the global scope is built
  again at each start." — dropped part: "`set --erase --universal
  fish_user_paths` clears an old one" — no file of `macos/tools/` has this
  command.
- C13 stated — section: `fish`, `touchid-sudo` — evidence:
  `macos/tools/fish/tools.toml:27-32`, `macos/tools/bash/tools.toml:18-20`,
  `macos/tools/zsh/tools.toml:17-19`, `macos/tools/touchid-sudo/tools.toml:9-17`
  — new: in `fish`: "`/etc/shells` is outside `$HOME`, so the line of fish is a
  `[mise.dotfiles]` entry, as the lines of `bash` and `zsh`. The login shell is
  `login_shell` of `[mise.bootstrap.user]`. `mise bootstrap` applies both." In
  `touchid-sudo`: "`/etc/pam.d/sudo_local` is outside `$HOME`, so it is a
  `[mise.bootstrap.files]` entry, which `mise bootstrap` applies. Each macOS
  update writes `/etc/pam.d/sudo` again, but `sudo_local` stays."
- C14 stated — already in `docs/design/mac.md` "The first setup and exit 4"
  (lines 271-317), and the Stage item of the legend links it.
- C15 stated — section: `fnox` — evidence:
  `macos/tools/fnox/files/.config/fnox/config.toml:5-13` — new: "fnox reads
  each secret from the macOS Keychain (the provider `global`, service `fnox`,
  prefix `global/`). `env = false` exports no secret to the shell, and
  `if_missing = "error"` stops a command when a secret is missing."
- C16 stated — section: `homebrew`, `fnox` — evidence:
  `macos/tools/homebrew/tools.toml:30-31`,
  `macos/tools/fnox/files/.config/fnox/config.toml:19-27` — new: in `homebrew`:
  "The alias `brew` runs `fnox -c $HOME/.config/fnox/config.toml exec -P brew --
  brew`, so `brew` gets the GitHub token of the profile `brew` of fnox. Thus
  the tool depends on `fnox`." In `fnox`: "The profile `brew` gives
  `GITHUB_API_TOKEN` and `HOMEBREW_GITHUB_API_TOKEN` to the command, from the
  Keychain item `GITHUB_API_TOKEN`."
- C17 stated — section: `homebrew` — evidence:
  `macos/tools/mise/tasks/updateall.py:117-129` — new: "`updateall` runs
  `upgrade:brew` under `fnox exec --profile brew` when fnox is on `PATH`, else
  without the token, with a warning. Its other `mise bootstrap` step gets no
  token from fnox."
- C18 stated — section: `fnox` — evidence:
  `macos/tools/fnox/files/.config/fnox/config.toml:29-32`,
  `macos/tools/fnox/tools.toml:1-2` — new: "The second profile is `fnox`. It
  gives `AGE_KEY`."
- C19 stated — section: `cli`, `git` — evidence:
  `macos/tools/cli/tools.toml:10-24`,
  `macos/tools/git/files/.config/git/identity.tmpl:5,18,28`;
  `docs/setup/tools.md:871-874` (the trust paths, already in the `mise`
  section, Trust paths) — new: in `cli`: "`projects_dir` (default
  `~/Projects`) sets `PROJECTS_DIR` and the alias `cdgh`. `git` and `mise` read
  it too." In `git`: "The `gitdir:` rules of `identity` use
  `<projects_dir>/github.com/` and `<projects_dir>/gitlab.com/`, so `git`
  needs `projects_dir` of `cli`."
- C20 stated — section: `git` — evidence:
  `macos/tools/git/files/.config/git/config.tmpl:3-6`,
  `macos/tools/git/files/.config/git/identity.tmpl:6-29` — new: "`config`
  includes `identity`. `identity` holds `[user]`, and includes `github.config`
  or `gitlab.config` by the URL of a remote (`hasconfig:remote.*.url:`) and by
  the folder of the repository (`gitdir:`)."
- C21 stated — section: `git` — evidence:
  `macos/tools/git/files/.config/git/config.tmpl:18-19,75`,
  `macos/tools/git/tools.toml:48` — new: "`push.default` is `current`, so the
  alias `gp` and the git alias `p` push only the current branch."
- C22 stated — section: `ssh` — evidence:
  `macos/tools/ssh/files/.ssh/config.tmpl:13-19`,
  `macos/tools/ssh/files/.ssh/config.identity.tmpl:3-5` — new: "`~/.ssh/config`
  includes `~/.ssh/config.local`, the hosts of one Mac, which no tool syncs,
  and, in `Host *`, `~/.ssh/config.identity`, which holds `User`. A missing
  `config.local` is valid. `StrictHostKeyChecking` is `accept-new`."
- C23 stated — section: `claude` — evidence:
  `macos/tools/claude/files/.claude/settings.json.merge.toml:76-85` — new (one
  sentence at the end of the Settings item): "It also removes
  `deniedMcpServers` and `env.CONTEXT7_RUNNER`, two keys that bootstrap set
  before."
- C24 stated — section: `claude` — evidence:
  `macos/tools/claude/files/.claude/settings.json.merge.toml:1-3`; the first
  part is already in the Settings item of `claude`
  (`docs/setup/tools.md:242-246`)
  — new (in the Settings item): "A key that the spec does not name stays as
  Claude Code or the owner wrote it." — dropped part: "on a new Mac, set it
  again in Claude Code" — with `backup`, the backup clone holds the synced
  `~/.claude/settings.json` (`docs/design/mac.md:362-400`), so the claim is
  not true in general, and no file of `macos/tools/` states it.
- C25 stated — section: `claude-cli` — evidence:
  `macos/tools/claude-cli/tools.toml:14-20`,
  `macos/tools/claude-cli/tasks/func-claude.py:13-17,57-85` — new: "The
  aliases `cc`, `ccf`, `cco`, `ccs` and `cch` run `func:claude` with the
  default model, Fable, Opus, Sonnet and Haiku (each with `[1m]`).
  `func:claude` takes an optional `name`, `-n/--new` and `-m/--model`. It
  starts `claude` with `--name <folder>[-<name>]`, `--resume` with the same
  name unless `--new` is set, and `--remote-control`,
  `--dangerously-skip-permissions` and `--allow-dangerously-skip-permissions`."
- C26 stated — section: `mise` — evidence:
  `macos/tools/mise/tasks/updateall.py:56-166`; the Entry of the Mac sync item
  of `mise` (`docs/setup/tools.md:841-847`) already holds that `updateall` runs
  `updateall:configs` and updates the dprint plugins — new (in the Entry of the
  Mac sync item, in place of its sentences about `updateall`):

  ```markdown
  `updateall` does these steps, in this order:

  1. `mise self-update`.
  2. `updateall:configs`. On exit 4 it stops with exit 4, and on another
     failure it stops with that code.
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
  ```

- C27 stated — section: `mise` — evidence:
  `macos/tools/mise/tasks/updateall.py:14,27-32,57-69,143-155` — new: "The
  weekly steps run when the stamp `/tmp/mise-updateall-weekly.stamp` is
  missing or older than one week, and `--force` runs them now. They come last,
  because they ask for sudo and restart Finder, the Dock and Control Center.
  macOS clears `/tmp` at a restart, and each macOS update restarts the Mac, so
  the first `updateall` after an update applies them again."
- C28 stated — section: `mise` — evidence:
  `macos/tools/mise/tasks/helpers.py:181-204`,
  `macos/tools/mise/tasks/updateall.py:119-120,139-164` — new: "A step of an
  opt-in tool (`upgrade:brew`, `upgrade:npm`, `upgrade:ai`, `cleanupDS`,
  `tailscale:daemon`) runs only when its task exists, else `updateall` prints
  one skip line."
- C29 stated — section: `tailscale` — evidence:
  `macos/tools/tailscale/tasks/tailscale-daemon.py:13-33`,
  `macos/tools/tailscale/tasks/upgrade-tailscale.py:13-16,77-90`,
  `macos/tools/homebrew/tasks/upgrade-brew.py:57-61` — new: "mise installs the
  formula. `tailscaled` needs root for its network interface, and mise has no
  root services, so `tailscale:daemon` starts it with
  `sudo brew services start tailscale`, only when it does not run. A start of
  a running daemon drops the tunnel. `upgrade:brew` calls `upgrade:tailscale`,
  which stops the service, gives the root-owned keg back to the user, upgrades
  the formula with mise and starts the service again. The App Store app cannot
  run the Tailscale SSH server, so the tool uses the formula."
- C30 stated — section: `macos-defaults` — evidence:
  `macos/tools/macos-defaults/tools.toml:1-4,58-68` — new: "A `defaults` value
  is not a template, so the hook `post-defaults` writes `NewWindowTargetPath`
  of Finder with `$HOME`. It also runs `chflags nohidden ~/Library`, and
  restarts `cfprefsd`, Finder, the Dock, `SystemUIServer` and Control Center,
  so they read the new values." — dropped part: "mise never restarts an app" —
  no file of `macos/tools/` states it.
- C31 stated — section: `macos-defaults` — evidence:
  `macos/tools/macos-defaults/tools.toml:21-26,83` — new: "`dock_orientation`
  is `bottom`, `left` or `right`. Its detect proposes the current position of
  the Dock, else `bottom`. Another value stops the render of the file."
- C32 stated — section: `homebrew` — evidence:
  `macos/tools/homebrew/tasks/upgrade-brew.py:63-68` — new: "`upgrade:brew`
  runs `mise bootstrap packages prune`, which removes each package that no
  tool and no `[mise.bootstrap.packages]` of `macos.toml` declares, then
  upgrades and applies the declared packages."

## Stale comments

- `macos/tools/mise/tasks/updateall.py:74-75` — says that mise "manages its own
  formula ([bootstrap.packages])" — mise is the standalone install at
  `~/.local/bin/mise` (`docs/design/mac.md:129-135`), and the `mise` tool
  declares only `brew:ca-certificates` (`macos/tools/mise/tools.toml:31-32`).
- `macos/tools/mise/tasks/upgrade-power.py:13-14` — names the file
  `conf.d/macos-defaults.toml` and the table `[bootstrap.macos.defaults]` — the
  generated file is `~/.config/mise/conf.d/macos-defaults/mise.toml`
  (`macos/tools/macos-defaults/tools.toml:4-5`), from
  `[mise.bootstrap.macos.defaults]`
  (`macos/tools/macos-defaults/tools.toml:29`).
- `macos/tools/touchid-sudo/tools.toml:2` — says that the target is a
  `[bootstrap.files]` entry — the table of the file is
  `[mise.bootstrap.files]` (`macos/tools/touchid-sudo/tools.toml:11`).
- `macos/tools/claude-cli/tasks/func-claude.py:20` — the usage line names
  `[claude args...]` — the task passes no other argument to `claude`
  (`macos/tools/claude-cli/tasks/func-claude.py:66-85`).
