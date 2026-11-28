# U1 — Repo tools: the CLAUDE.md section, contributions and skills

- **Read first:** each owned `tools.toml`;
  `macos/tools/claude/files/.claude/CLAUDE.md.merge.toml`
  (the model of a Markdown section spec); `docs/design/tools.md` ("The
  `[files]` table", "The `[defaults]`, `[questions]` and `[contributes]`
  tables"); `docs/design/cli.md` ("Shared files", "The spec", "Templates").
- **Also read** (for the skill texts): the stackgen source of each skill below,
  in
  `$HOME/Projects/github.com/virajp/claude-plugins/plugins/stackgen/tools/<slug>/`
  (`conventions.md`, `skills/<slug>/SKILL.md`, `skills/<slug>/references/*.md`)
  — read `origin/develop` of that clone (`git -C <clone> show
  origin/develop:<path>`),
  because a run can change the working tree there. Read `docs/design/tools.md`
  ("The header of a source file").
  **Lazy-load:** the Context7 docs of each tool (TypeScript, pnpm, Kotlin,
  kotlinx.coroutines, Gradle, Swift, SwiftPM), to examine each rule marked
  (unsure).
- **Lazy-load:** `repo/tools/vscode/` (a tool whose template reads
  contributions), `repo/tools/git/files/gitignore.merge.toml`.

## Ruling

> D1: Shape A: the owner tool ships SKILL.md as a template that reads .all; a
> short contribution of another tool is a value, a long one is a
> references/<tool>.md copy file that the contributing tool ships into the
> skill folder of the owner; local notes go in the seed references/local.md,
> which SKILL.md always names; the conventions go in
> references/conventions.md, which SKILL.md names.

> D2: Skills ship only when the claude tool is selected: the repo and Mac
> claude tools get [defaults] skills = true, and each [files] entry of a skill
> has when = "claude.skills"; skills = false stops all skills. The CLAUDE.md
> section ships whenever claude is selected; skills does not gate it.

> D3: The repo claude tool owns CLAUDE.md as a merge file (mode merge, source
> CLAUDE.md.merge.toml, format text, comment "<!--", one item of kind section,
> rendered as a template). Three parts. Tasks: run a task with mise run and
> list them with mise tasks; the aggregates setup, code, code:format,
> code:lint, code:sec, code:precommit, test, test:e2e, test:golden; the part
> form code:lint:NN-<name>, setup:deps:NN-<name>; add a tool with bootstrap
> add tool, never a bare mise use or a hand edit of a synced file; prefix each
> tool call with mise x --; an own task is a Python uv script that imports
> helpers. Dependencies: install with mise run setup, upgrade with mise run
> setup --upgrade, never the package manager directly; ranges in the manifest
> and the exact version in the lockfile; ask the user before adding a
> dependency (name it and what it is for); resolve a lockfile merge conflict by
> taking one side, then reinstall; then one line for each key of
> [contributes.claude.claude_md.deps] of the selected tools. Gates: one line
> for each key of [contributes.claude.claude_md.gates] (key = task name,
> value = one line). No repo mise skill.

> D4: typescript with the tsconfig rules (owner node), pnpm (owner pnpm),
> kotlin (owner kotlin), gradle (owner java), swift with the SwiftPM rules
> (owner swift).

> D6: eslint, ktlint, detekt, swiftlint, swift-format, ruff and
> analysis-options get no skill: each becomes a gate line of D3. uv and pub
> get no skill: their lines become dependency lines of D3 from python and
> flutter.

> D13 (edge rules): (a) U1 owns the skill texts and their [files] entries
> together; one commit. (b) A bootstrap check of the change runs with the
> worktree as the clone: mise x -- bootstrap check --clone <worktree root>.
> (e) The paths of a merged skill are the union of the two stackgen lists
> (typescript + tsconfig, swift + swiftpm). (f) A template tests that a tool is
> selected with has .all "<tool>", because a tool with no [defaults] is an
> empty table. (i) The CLAUDE.md section uses no Markdown heading: each part
> starts with a bold label (**Tasks**, **Dependencies**, **Gates**), so the
> heading rules of the linter do not fire above the H1 of the file. Run
> mise x -- mise run code:lint:05-linter on a scratch CLAUDE.md with the
> section above a # title and keep it green.
> D5: "If it can be learnt using context7 then it must NOT be in the skill."
> Keep only the owner's opinions, the bootstrap ties (tasks, mise, seeds) and
> traps that the docs do not warn about; drop what a gate or a shipped seed
> enforces. The unit checks each rule marked (unsure) with Context7 and drops
> it when the docs teach it. Never write "uv run ruff" or the .npmrc
> ignore-scripts/fund advice. Version rule: a range in the manifest, the exact
> version in the lockfile.

> D7: The header goes on lines 2 and 3 as YAML comments, inside the
> frontmatter that starts on line 1. A reference file gets the HTML-comment
> header of each Markdown file.

> D14 (b): The scratch CLAUDE.md of D13 i sits at lint-check/CLAUDE.md in the
> worktree root (a new, not ignored path, so gate_files lists it); pass it as
> a file to code:lint:05-linter, then delete the folder with rm. (d) gitleaks
> (code:sec:10-gitleaks), grype (code:sec:20-grype) and osv-scanner
> (code:sec:30-osv-scanner) each contribute one gates line.

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

Write all prose in ASD-STE100. Each new source file that is neither a seed nor
a merge spec starts with the `Synced from bootstrap` header.

1. **`repo/tools/claude/tools.toml`** — add `skills = true` to `[defaults]`
   with a comment (D2). Add `[files."CLAUDE.md"]` with `mode = "merge"` and
   `source = "CLAUDE.md.merge.toml"`. Add empty `[defaults.claude_md.deps]`
   and `[defaults.claude_md.gates]` tables, so a template can read them with
   no contribution.
2. **`repo/tools/claude/files/CLAUDE.md.merge.toml`** — the spec of D3: a
   comment block in the form of the Mac spec, `format = "text"`,
   `comment = "<!--"`, one `[[item]]` of `kind = "section"`. The section value
   renders the three parts of D3 as Markdown with no heading: each part starts
   with a bold label (`**Tasks**`, `**Dependencies**`, `**Gates**`) (D13 i). The
   Dependencies part writes
   the common rules, then one bullet for each key of `.config.claude_md.deps`
   in key order. The Gates part writes one bullet for each key of
   `.config.claude_md.gates` in key order, as `` `<task>`: <line> ``. A part
   with no key writes no empty label.
3. **Gate lines** — in each tool below, add
   `[contributes.claude.claude_md.gates]` with one key for each gate task the
   tool ships (the task name as installed, for example
   `"code:lint:30-detekt"`), and a value of one line: what the gate checks, its
   seed when it has one, and the rule "fix the code, not the seed" where a seed
   exists. Tools and tasks: `dprint` (05-dprint), `linter` (05-linter,
   06-shellcheck), `kotlin` (30-ktlint, 30-detekt), `swift` (50-swift-format,
   80-swiftlint: also "never run swiftlint --fix; swift-format owns the
   layout"), `flutter` (10-dart, 20-dart), `python` (50-pyrefly, 70-ruff,
   90-uv), `go` (40-go, 45-gomod), `android` (10-android), `mise`
   (60-bootstrap), `gitleaks` (code:sec:10-gitleaks), `grype`
   (code:sec:20-grype), `osv-scanner` (code:sec:30-osv-scanner) (D14 d). Read
   each task file for what it checks; do not invent.
4. **Dependency lines** — add `[contributes.claude.claude_md.deps]` with one
   key named for the tool and a value of one line:
   - `pnpm`: "pnpm 11 ignores onlyBuiltDependencies and its related keys;
     allow a build with an allowBuilds entry in pnpm-workspace.yaml, and never
     weaken strictDepBuilds or minimumReleaseAge."
   - `python` (uv): "Python comes from mise: never run uv python install and
     never pin a Python version in uv; use uv sync --locked through mise run
     setup, never --frozen by hand."
   - `flutter` (pub): "mise run setup --upgrade deletes pubspec.lock and runs
     pub upgrade --major-versions: it re-resolves every package. After an edit
     of pubspec.yaml, run mise run setup and commit both files."
   - `kotlin` (Gradle): "Each project needs dependencyLocking {
     lockAllConfigurations() }; ./gradlew dependencies --write-locks locks only
     the root and passes on a violation, so write the locks with mise run setup
     --upgrade, and commit the catalog edit with the lockfiles."
   - `swift` (SwiftPM): "The Package.resolved inside .xcodeproj belongs to
     Xcode: change it only with mise run setup or mise run setup --upgrade.
     SwiftPM has no audit command: read the revision diff of Package.resolved
     by hand."
5. **Skill entries** — in `node`, `pnpm`, `kotlin`, `java` and `swift`, add one
   `[files."<target>"]` entry for each skill file of the section "The skill
   texts" below: `SKILL.md` with `mode = "template"`, each
   `references/*.md` with `mode = "copy"`, `references/local.md` with
   `mode = "seed"`. Each has `when = "claude.skills"` (D2). Sources are below
   `files/.claude/skills/<skill>/`.

## Verification

- The rendered Markdown of the scratch consumer passes `mise x -- dprint check`
  (D16 b).
- `mise x -- bootstrap check --clone .` from the worktree root gives no new
  error (D13 b).
- `mise x -- mise run code:lint:05-linter` passes on a scratch `CLAUDE.md`
  with the rendered section above a `# Title` line (D13 i).
- `mise x -- go test ./e2e/repo/files/` from the root is green (U4 adds the
  new cases; this checks that nothing older breaks).
- The unit gate line `mise x -- mise run code:format --all` is green.

## Guardrails

- Do not touch `macos/`: U3 owns it.
- Do not edit a doc. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not regenerate `cli/internal/source/testdata/toml-yq.golden`: U7 does.

## The skill texts

For each skill, write three files below
`repo/tools/<owner>/files/.claude/skills/<skill>/`:

- `SKILL.md.tmpl` — frontmatter on line 1 (`name`, `description` with the
  trigger, `paths`: the union of the stackgen lists of the merged entries (D13
  e), `user-invocable: false`), the
  header on lines 2-3 (D7), then a short body (at most about 40 lines): the
  rules below, then a line that names `references/conventions.md` and
  `references/local.md`. The template may read `.all` to name a selected tool
  (for example, the `typescript` skill names pnpm only when
  `has .all "pnpm"` is true, D13 f); it never reads a value that may be missing
  without a default.
- `references/conventions.md` — the conventions prose of the tool, cut by D5,
  with the HTML-comment header. stackgen will read it from the consumer.
- `references/local.md` — a seed with no header: one heading and one line that
  says the repository writes its own rules here.

The rules that survive D5 (the audit of 2026-11-27). Keep each one only if
Context7 does not teach it; write each in ASD-STE100:

1. **`typescript`** (owner `node`; from stackgen `typescript` and `tsconfig`):
   errors are values at a boundary (services layer, job handler, route) and
   thrown inside a module, with one mapping home to coded responses; config is
   read from `process.env` once at the composition root, validated, typed,
   fail fast, no default for a secret; telemetry over OTLP only, never a vendor
   SDK, bridge a vendor exporter into OTLP; start telemetry in its own entry
   module before the framework (unsure); bound fan-out concurrency and choose
   fail-fast or collect-all; Vitest with v8 coverage at 100% through
   `include`, never a lower threshold, tests in sequence, test-only code in
   `src/_testUtils/`; run tests through package.json scripts; `@/` alias inside
   a package, never across a package edge, barrels only for public surfaces;
   files named `module.responsibility.ts`, private folders start with `_`; a
   `@/` import that fails at runtime means `tsc-alias` did not run; three
   config layers `tsconfig.base.json` (strict) → `tsconfig.json` (`noEmit`,
   `@/*` → `./src/*`) → `tsconfig.build.json` (emit, no tests), and `check`
   runs against the build config; `noUncheckedIndexedAccess` and the
   `noUnused*` pair on top of `strict`; `allowImportingTsExtensions` off in the
   build variant and the `base-url` replacer of `tsc-alias` off (unsure). Check
   which of naming-convention, consistent-type-imports,
   explicit-function-return-type, no-explicit-any, max-params, func-style and
   no-floating-promises the house linter (`@virajp.dev/linter`) turns on, and
   leave out each rule it enforces.
2. **`pnpm`** (owner `pnpm`): `"type": "module"` and `private: true` on each
   unpublished package; in a monorepo `workspace:*` links, a catalog for shared
   dependencies, fan-out from the root; the supply-chain settings live in the
   `pnpm-workspace.yaml` seed; the audit is advisory and the gate is
   `code:sec:30-osv-scanner`; `pnpm outdated` is a report only. (The install,
   upgrade, range, lockfile and `onlyBuiltDependencies` rules are in the
   `CLAUDE.md` section of U1: do not repeat them.)
3. **`kotlin`** (owner `kotlin`): the JDK comes only from the `java` tool
   (mise, Temurin LTS): no `jvmToolchain`, no `java { toolchain }`, no
   foojay; set the target with `jvmTarget`, `-Xjdk-release` and
   `options.release`; Gradle runs only as `./gradlew`, never `kotlinc` or a
   system gradle; the compiler version is the KGP version in the catalog;
   `allWarningsAsErrors = true`; no `!!`, and a platform type gets an explicit
   type at first use (unsure: detekt may enforce `!!`); one failure shape for
   each API (own exception hierarchy or a sealed result), `require`/`check`
   only for programmer error; tests on `kotlin.test` with JUnit 5, `runTest`,
   Kover, hand-written fakes over mocks; the LSP is
   `mise x -- kotlin-lsp --stdio` (dev, macOS). **In a library:** explicit API
   mode; depend only on the SLF4J, Micrometer and OTel APIs; read no
   environment variable or system property.
4. **`gradle`** (owner `java`): Kotlin DSL only; plugins through
   `alias(libs.plugins.x)`; each version in `gradle/libs.versions.toml` as a
   range (D5), none in a build script; ask before a new library, plugin or
   repository, and say that a plugin runs code at build time; repositories in
   settings only, with `FAIL_ON_PROJECT_REPOS`; apply ben-manes
   `com.github.ben-manes.versions` at the root, else the `--upgrade` report is
   skipped; build cache and configuration cache on in `gradle.properties`, but
   the task runs with `--no-configuration-cache`. (The locking rules are in the
   `CLAUDE.md` section of U1.)
5. **`swift`** (owner `swift`; from stackgen `swift` and `swiftpm`): Swift 6
   language mode with strict concurrency, fix the race and never silence it;
   `@unchecked Sendable` and `nonisolated(unsafe)` only on a reviewed
   lock-guarded type with a comment why, `@preconcurrency import` temporary and
   tracked (unsure); `Task.detached` almost never; no `precondition` or
   `fatalError` on input the caller controls, a comment on each `try?`; typed
   throws only for closed error sets; new tests use Swift Testing, not XCTest;
   the swift-format seed turns off NeverForceUnwrap and NeverUseForceTry on
   purpose because SwiftLint `force_unwrapping` covers them: do not turn them
   on; declare a package with `from:`, `exact:` only with a comment, `branch:`
   or `revision:` only for a temporary fork, never on the integration branch;
   the repository URL in a comment above each `.package`; commit
   `Package.resolved` in a library too; the bootstrap `.gitignore` ignores all
   of `.swiftpm/`, so shared Xcode settings cannot live in
   `.swiftpm/configuration/` (unsure which tool ships the line). **In a
   library:** no `@MainActor` on the API of a package with no UI; read no
   environment (`ProcessInfo`, `UserDefaults`, fixed paths); never call
   `LoggingSystem.bootstrap`, take a `Logger` and config values as parameters;
   bare semver tags (`1.4.0`), one minor of deprecation before a major.

### Verification of the skill texts

- Each `SKILL.md.tmpl` has `---` on line 1 and the header on lines 2-3.
- `grep -rn "uv run ruff\|ignore-scripts\|fund=false"
  repo/tools/*/files/.claude/`
  gives no hit.
- The unit gate line `mise x -- mise run code:format --all` is green.

### Guardrails of the skill texts

- Do not copy a stackgen file as it is: cut it by D5.
- Do not write a skill for `eslint`, `ktlint`, `detekt`, `swiftlint`,
  `swift-format`, `ruff`, `analysis-options`, `bash`, `markdown`, `uv`, `pub`
  or `ux-gate`.

## Commit

`feat: the bootstrap section of CLAUDE.md and the tool skills`
