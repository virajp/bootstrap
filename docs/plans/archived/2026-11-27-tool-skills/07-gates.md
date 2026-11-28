# U7 — Gates

- **Read first:** `cli/internal/source/toml_test.go` (lines 16-20 and 90-113).

## Ruling

> S1: cli/internal/source/testdata/toml-yq.golden is generated from every TOML
> file of repo/ and macos/; this fixed gates unit regenerates it.

> D14 (a): This plan must not land before the user releases the engine of
> 2026-11-27-mac-backup-target. Run
> git tag --contains $(git log --format=%H -1 --grep "mac-backup-target"
> origin/main) --list "bootstrap@*"
> and, when it prints nothing, report a QUESTION (release first, or hold the
> landing).

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

1. **`cli/internal/source/testdata/toml-yq.golden`** — in `cli/`, run
   `UPDATE_GOLDEN=1 mise x -- go test ./internal/source`. Keep only the lines
   of the TOML files this plan changed or added (the `tools.toml` files of U1
   and U3 and `macos/examples/macos.toml`; no `*.merge.toml`, D13 j); report any
   other changed line as a QUESTION.
2. Run the release check of D17 c (it replaces the commands of D14 a and D16 c).
3. Run each wave gate and plan gate line of `index.json`, and report each
   result.

## Verification

- `mise x -- mise run code:lint --all` is green.
- `env CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins mise
  x -- mise run test`
  is green.

## Guardrails

- Bump no version. Edit no file but the golden.

## Commit

`test: regenerate the TOML golden file`
