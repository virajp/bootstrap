# U3 — Mac tools: secrets, the npm hook, mac-setup and git-identity

- **Read first:** each owned `tools.toml`;
  `macos/tools/claude/files/.claude/settings.json.merge.toml` (its hooks item
  at :58-63); `docs/design/tools.md` ("The `[files]` table", with the `root`
  key of plan `2026-11-27-mac-backup-target`); `docs/design/mac.md` ("The
  backup clone"); `docs/setup/tools.md` (the sections of the owned tools).
- **Lazy-load:**
  `$HOME/Projects/github.com/virajp/claude-plugins/plugins/stackgen/tools/pnpm/hooks/npm-normalize.sh`
  and `hooks.yaml` (read `origin/develop` of that clone).

## Ruling

> D1: Shape A (see the plan): owner template SKILL.md reading .all, values for
> short contributions, references/<tool>.md copy files for long ones, the seed
> references/local.md, references/conventions.md.

> D2: Skills ship only when the claude tool is selected: the Mac claude tool
> gets [defaults] skills = true, and each [files] entry of a skill has
> when = "claude.skills".

> D5: Nothing that Context7 can teach; keep the owner's opinions, the
> bootstrap ties and the traps.

> D7: The header goes on lines 2 and 3 of SKILL.md as YAML comments.

> D8: secrets in ~/.claude/skills/secrets/ (owner fnox; contributors
> 1password-cli and age as values): secrets come from fnox (the macOS
> Keychain) or op; give a secret to a command with fnox exec or op run; never
> print a secret, never write it to a file, never put it in chat. mac-setup
> (owner mise; contributors homebrew, mas, setapp, macos-defaults) and
> git-identity (owner git; contributors 1password, ssh, gh, glab) as
> root = "backup" entries at .claude/skills/<skill>/ of the backup clone
> (macos-setup). mac-setup: to install or remove a tool or an app, or to
> change a setting, edit macos.toml and run mise run updateall:configs; never
> brew install, mise use -g or defaults write directly; dev tools from mise,
> others from Homebrew; a file with the Synced from bootstrap header is
> changed in bootstrap. git-identity: the identity for each forge comes from
> includeIf, never a local user.email; commits are signed by the 1Password
> SSH signer; if a signature fails, tell the user to unlock 1Password and
> never turn off signing; gh auth status / glab auth status, never print a
> token.

> D9: macos/tools/claude ships .claude/hooks/npm-normalize.sh (bash, mode 755)
> and a PreToolUse item (matcher Bash) in its settings.json merge spec. The
> hook reads the JSON input with jq. If command -v pnpm finds pnpm, it changes
> npm to pnpm, npx to pnpm dlx, and npm ci to pnpm install --frozen-lockfile;
> else it changes nothing. No lockfile walk, no bun. The Mac claude tool gets
> depends = [ "jq" ].

> D13 (k): Each Mac owner (fnox, mise, git) has [defaults.skill.lines] as an
> empty table; a contributor writes [contributes.<owner>.skill.lines] <tool> =
> "<line>"; the template renders one bullet for each key in key order. (f) A
> template tests that a tool is selected with has .all "<tool>".

> D14 (c): The Markdown that U3 ships into the backup clone passes the gates
> of macos-setup: render each SKILL.md and reference of mac-setup and
> git-identity into a scratch folder and run mise x -- dprint check with the
> dprint config of this repository on them; write the text so dprint changes
> nothing.

> D15: The hook returns only hookSpecificOutput with hookEventName PreToolUse
> and updatedInput, with no permissionDecision, so the changed command goes
> through the normal permission rules. With no change it prints {}. User,
> 2026-11-28: "No decision, only updatedInput".

> D16 (edge rules, round 3): (a) U6 also updates docs/sync/tools/gitleaks.md,
> grype.md and osv-scanner.md, and CLAUDE.md:123-124. (b) U1 renders the
> CLAUDE.md section and each SKILL.md and reference in a scratch consumer
> (mise x -- bootstrap sync --clone <worktree root> in a temporary git
> repository that selects claude, node, pnpm, kotlin, java and swift) and runs
> mise x -- dprint check with the dprint config of this repository on the
> written Markdown. (c) The release check tests the last engine change:
> git tag --contains $(git log --format=%H -1 origin/main -- cli/internal/)
> --list "bootstrap@*".
> (d) The gate lines keep mise run <task> --all. (e) The hook writes the whole
> tool_input back with only the command changed (jq: .tool_input | .command =
> $new). (f) When the command holds a newline, the hook changes nothing and
> prints {}. (g) A seed references/local.md left in a folder with no SKILL.md
> is accepted; M62 and the claude tool doc say the user can delete it. (h) U1
> updates the description and summary of repo/tools/claude/tools.toml; U3
> those of the Mac claude, fnox, mise and git tools. (i) M62 tells linter and
> virajp.dev not to run a stackgen sync or materialize until the stackgen
> release of the parked change; if one writes over a bootstrap skill, run mise
> run setup:configs again. (j) U4 owns all of e2e/repo/.

> D17 (edge rules, round 4): (a) The Gates bullets of CLAUDE.md name the task
> only: `<task>`: <line>; a value holds no command and no --all. The scanner
> lines say what the scanner finds. (b) The Tasks part names code:precommit
> only when has .all "pre-commit" is true. (c) The release check tests the
> engine commit of plan 2026-11-27-mac-backup-target:
> git tag --contains $(git log --format=%H -1 --grep "^feat: a Mac \[files\]
> entry can write into the backup clone" origin/main) --list "bootstrap@*".
> (d) The Mac claude tool adds ".claude/skills" and ".claude/hooks" to its
> runtime list, so the link rule never links these two folders; the sync links
> .claude/skills/secrets/ and .claude/hooks/npm-normalize.sh one by one, and
> the other items of ~/.claude/skills/ and ~/.claude/hooks/ stay in place. U4
> tests that a foreign item of ~/.claude/skills/ stays a real item in $HOME
> after a sync with backup.

## Edits

Write all prose and comments in ASD-STE100.

1. **`macos/tools/claude/`** — `tools.toml`: `skills = true` in `[defaults]`;
   `depends` gets `"jq"`; `runtime` gets `".claude/skills"` and
   `".claude/hooks"` (D17 d); a `[files.".claude/hooks/npm-normalize.sh"]` entry
   (copy, `perm = "755"`). `files/.claude/hooks/npm-normalize.sh`: the stackgen
   script cut to D9, with the header after the shebang, BSD `sed` only.
   `files/.claude/settings.json.merge.toml`: one more hooks item, a
   `PreToolUse` block with `matcher = "Bash"` and the command
   `$HOME/.claude/hooks/npm-normalize.sh` (match the form of the existing
   `PostToolUse` item; give it an `id`).
2. **`secrets`** — in `macos/tools/fnox/`: `[files]` entries for
   `.claude/skills/secrets/SKILL.md` (template), `references/conventions.md`
   (copy) and `references/local.md` (seed), each `when = "claude.skills"`, and
   the files below `files/.claude/skills/secrets/`. `[defaults.skill.lines]` is
   an
   empty table (D13 k). `1password-cli` and `age` each add one
   line with `[contributes.fnox.skill.lines]` (`op run` / `op read` for
   1password-cli; encrypt a file with age, never commit the key, for age).
3. **`mac-setup`** — in `macos/tools/mise/`: the three entries with
   `root = "backup"` and `when = "claude.skills"`, targets
   `.claude/skills/mac-setup/SKILL.md` and its two references, and the source
   files. `[defaults.skill.lines]` of `mise` is an empty table (D13 k).
   `homebrew`,
   `mas`, `setapp` and `macos-defaults` each add one line
   with `[contributes.mise.skill.lines]` (for example homebrew: casks and
   non-dev tools go in the Brewfile through macos.toml; mas: App Store apps by
   id in macos.toml; setapp: Setapp apps are installed in Setapp, not by the
   sync; macos-defaults: a setting is a value in macos.toml, never a
   `defaults write`). Read each tool's `tools.toml` and `docs/setup/tools.md`
   section first; write only what is true.
4. **`git-identity`** — in `macos/tools/git/`: the three entries with
   `root = "backup"` and `when = "claude.skills"`, and the source files.
   `[defaults.skill.lines]` of `git` is an empty table (D13 k). `1password`,
   `ssh`, `gh` and `glab` each add one line with
   `[contributes.git.skill.lines]`.
5. **`macos/examples/macos.toml`** — select `jq` if it is not selected, so
   `e2e/macostools` stays green with the new `depends`.

## Verification

- `mise x -- go test ./e2e/macostools/` from the root is green.
- With a command that holds a newline, the hook prints `{}` (D16 f); the
  output keeps `timeout` and `description` of the input (D16 e).
- `bash -n macos/tools/claude/files/.claude/hooks/npm-normalize.sh` passes, and
  `echo '{"tool_input":{"command":"npm ci && npx tsc"}}' | PATH=<a dir with a
  fake pnpm>:$PATH bash <the script>`
  prints the rewritten command in `updatedInput` with no `permissionDecision`
  key (D15); with no pnpm on PATH it prints `{}`.
- The rendered Markdown of mac-setup and git-identity passes
  `mise x -- dprint check` with the dprint config of this repository (D14 c).
- The unit gate line `mise x -- mise run code:format --all` is green.

## Guardrails

- Do not touch `repo/`: U1 owns it.
- Do not edit a doc: report each falsified passage as `DOCS FALSIFIED:`.
- A `root = "backup"` entry cannot use `mode = "merge"`.

## Commit

`feat: the Mac skills and the npm hook`
