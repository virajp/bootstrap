# U4 — Tests

- **Read first:** `e2e/internal/repocase/` (the case runner), the existing
  cases of `e2e/repo/files/merge_test.go` and `conditions_test.go`, the backup
  cases of `e2e/mac/sync_test.go` (:1344-1410),
  `e2e/macostools/macostools_test.go`.

## Ruling

> D2: Skills ship only when the claude tool is selected; each skill entry has
> when = "claude.skills"; skills = false stops all skills. The CLAUDE.md
> section ships whenever claude is selected.

> D3: The repo claude tool owns CLAUDE.md as a merge file with one section in
> three parts (Tasks, Dependencies, Gates), filled from the contributions of
> the selected tools.

> D9: The npm hook changes npm only when pnpm is on PATH.

> D13 (c): U4 tests the real Mac tools in e2e/macostools: a case with a copy
> of macos/examples/macos.toml plus a backup key set to a temporary clone
> checks ~/.claude/skills/secrets/, the PreToolUse item, and
> .claude/skills/mac-setup/ and git-identity/ in the clone.

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

1. **`e2e/repo/files/`** — add cases:
   - a consumer with `claude` and `node`: `CLAUDE.md` gets the section at the
     top above the existing text, the Dependencies part has no pnpm line, and
     `.claude/skills/typescript/SKILL.md` exists with `---` on line 1 and the
     header on lines 2-3;
   - the same with `pnpm`: the pnpm dependency line and
     `.claude/skills/pnpm/` exist;
   - `[tools.claude] skills = false`: no skill file, the section stays;
   - no `claude`: no `CLAUDE.md` change and no skill;
   - a second sync keeps a line that the user wrote after `bootstrap:end`, and
     keeps an edited `references/local.md`.
2. **`e2e/macostools/`** — add one case (D13 c): a copy of
   `macos/examples/macos.toml` with a `backup` key set to a temporary git
   clone; after `bootstrap mac sync`, `~/.claude/skills/secrets/SKILL.md`
   exists, `~/.claude/settings.json` has the `PreToolUse` item with the hook
   command, and `.claude/skills/mac-setup/SKILL.md` and
   `.claude/skills/git-identity/SKILL.md` are real files in the clone. Add a
   second case without `backup`: the two backup skills are skipped. Change any
   other check only where it fails because of the new files or the new `jq`
   dependency.

## Verification

- `env CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins mise
  x -- go test ./e2e/...`
  from the root is green, save the golden of `cli/` (U7).
- The unit gate line `mise x -- mise run code:format --all` is green.

## Guardrails

- Do not change a file of `repo/` or `macos/`: if a case shows a defect there,
  report it as a QUESTION.

## Commit

`test: the CLAUDE.md section, the skills and the Mac hook`
