GAP 2026-11-28 ★4 | repo:claude-plugins, plugin:stackgen,
plan:2026-11-27-tool-skills | stackgen stops landing tool knowledge, and its
skills point to the bootstrap section of CLAUDE.md | parked: a change plan in
claude-plugins after plan 2026-11-27-tool-skills lands

**Source** The design of tool skills, 2026-11-27 and 2026-11-28 (plans
`2026-11-27-tool-selection`, `2026-11-27-mac-backup-target` and
`2026-11-27-tool-skills` of bootstrap). The user ruled: "Later in
claude-plugins", and moved the change of the mise-skill line here.

What bootstrap now owns: the tool skills `typescript`, `pnpm`, `kotlin`,
`gradle` and `swift` in `.claude/skills/<skill>/` of a consumer, each with
`references/conventions.md`; the bootstrap section of the repo `CLAUDE.md`
(Tasks, Dependencies, Gates); the npm hook on the Mac side.

What the change plan in claude-plugins does:

1. stackgen stops landing tool knowledge: delete
   `plugins/stackgen/tools/<slug>/`
   (18 entries) and the pnpm hook (`hooks/npm-normalize.sh`, `hooks.yaml`),
   the test task `.config/mise/tasks/p/plugins/npm-normalize-test.py`, its
   pre-commit hook and its CI step. Edit every file that names tool knowledge
   (survey of 2026-11-27: the stack-template skill and `materializer.md`
   :19-20, :44, :49-50, :238, :398-415; `stackgen-sync` :5, :9, :59-66, :91,
   :96-98; `taxonomy.md`, `pack-format.md` :162-199, `output-tree.md`,
   `artifact-doctrine.md`, `stacks/readme.md`, `stacks/inventory.md`;
   `plugin.json` and `marketplace.json`; vwf `stack-adapter.md`,
   `doctor/references/stack-checks.md`, `setup/references/materialize.md`;
   `scripts/src/inventory.ts`, `check.ts` and their tests; the repo docs and
   the site docs).
2. `ux-gate` (now in `tools/typescript/skills/ux-gate/`) stays in stackgen: it
   moves to a pack, for example `stacks/framework/html/skills/ux-gate/`.
3. The materializer folds the conventions of a tool into a template entry by
   reading `.claude/skills/<skill>/references/conventions.md` of the consumer.
4. `stackgen-sync` gets a row of its one-time conversion: a row whose source is
   `tool/<slug>@<v>` is removed, and the file stays. The bullet at
   `SKILL.md:64-67`, which makes `tool/` rows, is reworded.
5. Replace the line "The repo's mise skill, .claude/skills/mise/, lists every
   task." (ruling D6 of `2026-11-26-stackgen-bootstrap-tasks`) in each stackgen
   skill with a pointer to the bootstrap section of the repo `CLAUDE.md`. Fix
   the record `docs/memory/decisions/2026-11-26-mise-skill-is-bootstraps.md`
   (D14 of that plan): bootstrap ships no mise skill; the task text is in the
   bootstrap section of `CLAUDE.md`.

Order: after `2026-11-26-stackgen-bootstrap-tasks` lands in claude-plugins,
and after `2026-11-27-tool-skills` lands in bootstrap, so the paths are final.
