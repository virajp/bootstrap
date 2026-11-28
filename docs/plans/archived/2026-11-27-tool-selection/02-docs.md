# U2 — Docs

- **Read first:** `docs/sync/index.md` ("Files in this folder" and "Files that
  bootstrap no longer ships"), the last item of `docs/sync/migrate.md` (M59),
  and each owned file at the line named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

> D1: Drop the bun tool.

> D3: bunfig.toml is a seed, so the consumer owns it and keeps it. It gets no
> row in "Files that bootstrap no longer ships", because step 4 deletes each
> path of that table and the adoption check fails while one exists. M60 says
> that bunfig.toml stays, and that the user can delete it by hand.

> D4: Two items. M60 (the bun tool is gone): every engine command reads the
> clone, so bootstrap remove tool bun fails with code 3 once bun is gone; the
> steps are: delete the [tools.bun] table (and any [overrides.tools.bun]) from
> .config/bootstrap/tools.toml by hand, and each override that names bun
> ([overrides.mise.tools.bun], [overrides.mise_env.<env>.tools.bun]), select
> node with pnpm or aube when the
> repository still needs a package manager, then run mise run setup:configs;
> bunfig.toml stays (D3). M61 (android brings kotlin): the next sync adds the
> kotlin tool, its gates and seeds to an android or flutter repository; when
> the new gates fail on existing Kotlin code (for example the MainActivity.kt
> of Flutter), run mise run code:format --fix for ktlint, then fix each detekt
> finding in the code; never weaken the seeds. When setup:deps:30-kotlin stops
> because gradle-wrapper.properties has no distributionSha256Sum, add the
> checksum as the message of the task says.

> D2: android depends on java and kotlin: depends = [ "java", "kotlin" ].
> java stays a direct dependency, because the mise depends of a tool come from
> its direct depends only, and android-sdk needs the JDK of java. flutter gets
> kotlin through android: the ktlint and detekt gates, the seeds
> .config/detekt.yml and .config/ktlint.editorconfig, setup:deps:30-kotlin and
> the Kotlin LSP. Write this wording in every doc.

> D9: kotlin comes in the environments of android (dev and ci), so the CI of
> each Android and Flutter repository installs ktlint and detekt and runs the
> two gates. User, 2026-11-27: "Yes, run the Kotlin gates in CI". kotlin keeps
> default_envs = [ "dev" ] for a repository that selects it alone. M61 says
> that CI now runs the Kotlin gates.

> D11: U2 is the fixed docs unit: it edits the docs it owns.

> D12: The Computed depends cell of the android row of docs/design/mise.md:524
> gets the value that the generator writes (U1 reports it in DECIDED:).

## Edits

Write all prose in ASD-STE100.

1. Run `/vwf:docs-sync` over the branch delta of the run, and apply its
   findings and every `DOCS FALSIFIED:` line of U1.
2. **`docs/sync/tools/bun.md`** — delete it with `rm`.
3. **`docs/sync/index.md`** — remove `bun.md` from the tool list (about line
   58). Add no row to "Files that bootstrap no longer ships" (D3).
4. **`docs/sync/migrate.md`** — add `## M60: The bun tool is gone` and
   `## M61: android brings kotlin` after M59, in the form of M59: a short
   paragraph, then numbered steps. Change the sentence of M30 (about line 684)
   "A repository with only the `bun` tool does nothing." only if it now reads
   as a current instruction; it is history, so leave it if it is a past item.
5. **`docs/sync/tools/node.md`** — remove the "`bun`, and no other" row (about
   line 79) and each other bun passage.
6. **`docs/sync/tools/android.md`** (about line 8, Depends), **`flutter.md`**
   (about lines 8 and 35-37) and **`kotlin.md`** (about lines 140-146: the CI
   text, which D9 makes untrue for an android or flutter repository) — state the
   new
   closure: android → java and kotlin, kotlin → java, and flutter → android.
   **`docs/sync/tools/java.md`** (about lines 16-17 and 73-76) — it says that
   android adds java directly and that "`kotlin` is in dev only"; write the
   new closure and the environments that `bootstrap sync` now gives `kotlin`
   in an android repository.
7. **`docs/design/mise.md`** — remove the `bun` row (about line 538); fix the
   closure lines (about lines 524 and 806-807).
8. **`docs/design/tools.md`** (about lines 584-585) — fix the closure text.
9. **`docs/design/cli.md`** (about lines 704-716) — the sample output says
   "kotlin not offered … evidence of android". Write a sample that is true
   with the new closure (run `bootstrap detect` in a scratch consumer with an
   Android Gradle file if a doubt stays).

10. **`README.md`** (about lines 128-136) — the sample `tools.toml` of android
    shows `depends = [ "java" ]`; change it to `depends = [ "java", "kotlin" ]`.

## Verification

- `grep -rn '\bbun\b' docs/design docs/sync | grep -v 'migrate.md'` gives only
  no hit (D3: no row in the no-longer-ships table).
- `mise x -- mise run code:format --all` is green.

## Guardrails

- Do not touch any file outside `owns`. `docs/plans/archived/`,
  `docs/scratchpad/` and `docs/memory/` are history: do not edit them.
- Delete with `rm`, never `git rm`.

## Commit

`docs: drop bun, and android brings kotlin`
