DECISION 2026-11-28 ★4 | repo:bootstrap, repo:claude-plugins, repo/tools/claude,
macos/tools/claude, plugin:stackgen | bootstrap owns the tool knowledge (the
skills keyed by a bootstrap tool) and ships no repo mise skill; the task text,
the dependency rules and the gate lines are the bootstrap section of CLAUDE.md
of the claude tool | alt rejected: stackgen keeps the tool knowledge, a repo
mise skill

# The tool skills belong to bootstrap

The user ruled this in the plan interview of
`docs/plans/2026-11-27-tool-skills` (its decisions D1 to D20), over 2026-11-27
and 2026-11-28. It is plan 3 of the skill delivery work of 2026-11-27.

## Reversals (D12)

- **The tool knowledge moves to bootstrap.** The split of 2026-11-14
  (`2026-11-14-bootstrap-plugins-split.md`) gave the stack knowledge to
  stackgen, and its R5 amendment (B3) kept the tool knowledge in stackgen,
  keyed by the bootstrap tool name. The tool knowledge now belongs to
  bootstrap: the tool that a skill is about ships the skill. The other stack
  knowledge stays in stackgen.
- **No repo mise skill.** Item 1 of the handoff prompt
  `docs/prompts/2026-11-27-mise-skill-and-flutter-test.md` and D14 of the
  claude-plugins plan `2026-11-26-stackgen-bootstrap-tasks` said that
  bootstrap ships `.claude/skills/mise/`. The user ruled: "move both to
  CLAUDE.md and drop the mise skill". The task rules, the dependency rules and
  the gate lines go into one bootstrap section at the top of the `CLAUDE.md`
  of a consumer (D3).

## The rulings

- **The section (D3, D13 i, D17).** The repo `claude` tool owns `CLAUDE.md`
  as a merge file with one `section` item. The section has three parts, each
  with a bold label and no heading: Tasks, Dependencies and Gates. A tool adds
  a line with `[contributes.claude.claude_md.deps]` (key: the tool) or
  `[contributes.claude.claude_md.gates]` (key: the gate task).
- **The delivery (D1, D2).** The owner tool ships `SKILL.md` as a template,
  `references/conventions.md` and the seed `references/local.md`. A short
  contribution of another tool is a value, and a long one is a
  `references/<tool>.md` file. Each skill entry has `when = "claude.skills"`,
  and `[defaults] skills = true` of the `claude` tool of each side.
- **The repo skills (D4, D6).** `typescript` (`node`), `pnpm` (`pnpm`),
  `kotlin` (`kotlin`), `gradle` (`java`) and `swift` (`swift`). The other
  stackgen tool entries get no skill: a gate rule becomes a gate line, a
  `uv` or `pub` rule becomes a dependency line, and `bash` and `markdown` are
  dropped. `ux-gate` stays in stackgen.
- **The content (D5).** "If it can be learnt using context7 then it must NOT
  be in the skill." A skill keeps only the opinions of the owner, the ties to
  bootstrap and the traps that the docs do not warn about.
- **The header (D7).** On lines 2 and 3 of `SKILL.md`, as YAML comments inside
  the frontmatter, because Claude Code reads the frontmatter only on line 1.
- **The Mac skills (D8).** `secrets` (`fnox`) in `~/.claude/skills/`;
  `mac-setup` (`mise`) and `git-identity` (`git`) in the backup clone.
- **The npm hook (D9, D15, D18, D19).** The Mac `claude` tool ships
  `.claude/hooks/npm-normalize.sh` and its `PreToolUse` item. With pnpm on
  PATH, it changes `npm` to `pnpm`, `npx` to `pnpm exec` with the flags kept,
  and `npm ci` to `pnpm install --frozen-lockfile`. It returns only
  `updatedInput`, with no permission decision.
- **The alias (D20).** The shell alias `npx` of the `pnpm` tool is
  `pnpm exec`, not `pnpm dlx`.
- **The consumers (D10, D13 d, D16 i).** M62 of `docs/sync/migrate.md`.

## Parked

The stackgen side (stop landing tool knowledge, the lockfile conversion,
`ux-gate` to a pack) is
`docs/memory/gaps/2026-11-28-stackgen-tool-knowledge-moves-parked.md`.
