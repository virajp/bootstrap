# U6 — Docs

- **Read first:** `docs/sync/index.md`, the last item of `docs/sync/migrate.md`,
  `docs/sync/tools/claude-tool.md`, the sections of the owned tools in
  `docs/setup/tools.md`, `docs/design/tools.md` ("The header of a source
  file").
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`,
  `${CLAUDE_PLUGIN_ROOT}/assets/memory.md` (the decision record format).

## Ruling

> D1, D2, D3, D4, D8, D9: as quoted in the plan's decisions.

> D10: M62 in docs/sync/migrate.md: the next sync adds the bootstrap section at
> the top of CLAUDE.md and the skills of the selected tools. A consumer with
> stackgen copies (linter, virajp.dev) deletes
> .claude/skills/{eslint,ktlint,detekt,swiftlint,swift-format,analysis-options}/
> where they exist, and .claude/hooks/npm-normalize.sh with its PreToolUse
> entry in .claude/settings.json; bootstrap writes the kept skills over the old
> copies at the same paths; the next stackgen release removes its tool/
> lockfile rows.

> D12: U6 writes docs/memory/decisions/2026-11-28-tool-skills-are-bootstraps.md:
> bootstrap owns tool knowledge (the 2026-11-14 split gave stack knowledge to
> stackgen), and bootstrap ships no repo mise skill; the task text is the
> bootstrap section of CLAUDE.md (replaces item 1 of
> docs/prompts/2026-11-27-mise-skill-and-flutter-test.md and D14 of the
> claude-plugins plan 2026-11-26-stackgen-bootstrap-tasks).

> D13 (d): M62 also deletes the old stackgen folders
> .claude/skills/{tsconfig,swiftpm,ruff,uv,pub,bash,markdown}/, and deletes the
> old folders of the kept skills
> .claude/skills/{typescript,pnpm,kotlin,gradle,swift}/
> before the sync, so no old stackgen reference stays; the sync then writes
> them again. (g) U6 also owns docs/design/bootstrap-config.md and adds the key
> skills of [tools.claude].

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

## Edits

Write all prose in ASD-STE100.

1. Run `/vwf:docs-sync` over the branch delta of the run, and apply its
   findings and every `DOCS FALSIFIED:` line of U1 to U4.
2. **`docs/sync/tools/<tool>.md`** of `claude` (`claude-tool.md`), `node`,
   `pnpm`, `kotlin`, `java`, `swift`, `python`, `flutter`, `dprint`, `linter`,
   `go`, `android`, `mise` — the new files, values and contributions (CLAUDE.md
   section, skills, gate and dependency lines).
3. **`docs/setup/tools.md`** — the sections of `claude` (skills, the hook, the
   `jq` dependency), `fnox`, `1password-cli`, `age`, `mise`, `homebrew`, `mas`,
   `setapp`, `macos-defaults`, `git`, `1password`, `ssh`, `gh`, `glab`.
4. **`docs/design/tools.md`** — a short rule for skills (D1, D2) and the
   header of a `SKILL.md` (D7).
5. **`docs/sync/migrate.md`** — `## M62: The bootstrap section of CLAUDE.md
   and the tool skills`, per D10 and D13 d.
   5a. **`docs/design/bootstrap-config.md`** (about :150-163) — the key `skills`
   of `[tools.claude]` (D13 g).
6. **`docs/prompts/2026-11-27-mise-skill-and-flutter-test.md`** — remove item
   1 and its sources, and add one line that points to D12; keep item 2 (the
   `flutter` test part).
7. **`docs/memory/decisions/2026-11-28-tool-skills-are-bootstraps.md`** — the
   record of D12, in the format of `docs/memory/decisions/`.
8. **`CLAUDE.md`** — "Layout rule": a tool folder can hold
   `files/.claude/skills/<skill>/`; the `claude` tool owns the bootstrap
   section of `CLAUDE.md`.

## Verification

- `mise x -- mise run code:format --all` is green.
- `grep -rn "mise skill" docs/sync docs/design CLAUDE.md` gives no passage
  that says bootstrap ships a mise skill.

## Guardrails

- Do not edit `docs/plans/archived/`, `docs/scratchpad/` or older records of
  `docs/memory/`.
- The bootstrap section that the self-sync adds to this repository's
  `CLAUDE.md` comes after the landing (A1): do not write it by hand.

## Commit

`docs: the tool skills and the bootstrap section of CLAUDE.md`
