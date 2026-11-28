# U1 — Tools and tests

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `docs/design/tools.md` ("The `tools.toml` file", the `depends`
  key), `e2e/internal/repocase/` (the case runner),
  `repo/tools/kotlin/tools.toml`.

## Ruling

> D1: Drop the bun tool. User, 2026-11-27: "none of my projects are using bun
> as of now, we can drop support for bun".

> D2: android depends on java and kotlin: depends = [ "java", "kotlin" ].
> Android is Kotlin-first; java stays a direct dependency, because the mise
> depends of a tool come from its direct depends only
> (cli/internal/misegen/extend.go:115-160), and android-sdk needs the JDK of
> java. flutter gets kotlin through android: the ktlint and detekt gates, the
> seeds .config/detekt.yml and .config/ktlint.editorconfig,
> setup:deps:30-kotlin and the Kotlin LSP. User, 2026-11-27: "Yes, flutter gets
> kotlin through android".

> D9: kotlin comes in the environments of android (dev and ci), so the CI of
> each Android and Flutter repository installs ktlint and detekt and runs the
> two gates. User, 2026-11-27: "Yes, run the Kotlin gates in CI". kotlin keeps
> default_envs = [ "dev" ] for a repository that selects it alone. Do not
> change repo/tools/kotlin/tools.toml.

> D8: U1 Verification and the plan gate set
> CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins, so the
> cases that add android run.

> D5: No review unit. The one runnable change deletes the bun skip (three
> lines) of repo/tools/node/tasks/setup-deps-node.py.

> D11: U1 Edit 3 runs bootstrap sync only in a scratch consumer outside the
> worktree, as a check; it writes nothing in the tree.

> D12: The extra depends of android-sdk (the mise tools of kotlin) are
> accepted. Record the computed depends of a generated
> conf.d/android/mise.<env>.toml in DECIDED:, for U2.

> D13: The migrate case checkGoldenPartsQuiet (android + swift, real mise run
> test:golden, no auto install) now also gets the kotlin mise tools. If it
> fails only because of the kotlin tools, report a QUESTION with the output
> and the options (change the values of the case, or accept), and change
> nothing else.

## Edits

1. **`repo/tools/bun/`** — delete the whole folder (`tools.toml`,
   `files/bunfig.toml`, `scripts/detect`) with `rm -r`.
2. **`repo/tools/node/tasks/setup-deps-node.py`** — delete the comment line
   "The bun tool owns the dependencies of a Bun repository, and it has no
   task." and the `if not has_pnpm and not has_aube and os.path.isdir(...bun)`
   block with its `sys.exit(0)` (about lines 33-35). Keep `has_pnpm` and
   `has_aube` only if a later line still reads them; if no line reads them
   after the delete, delete their assignments too, and the `conf_d` line if it
   is then unused.
3. **`repo/tools/android/tools.toml`** — change `depends = [ "java" ]` to
   `depends = [ "java", "kotlin" ]` (D2). Change the `description` text "It
   needs the java tool." to "It needs the java and kotlin tools." Then run the
   migrate case that adds android
   (`e2e/repo/migrate/migrate_test.go`, about line 510) with `HARNESS_KEEP=all`
   and `CLAUDE_PLUGINS_CLONE` set, and examine its generated
   `.config/mise/conf.d/android/mise.*.toml`: the `android-sdk` entry still
   depends on the JDK of java. If no case writes that file, run
   `bootstrap sync` in a scratch consumer that selects android, and report the
   result in `DECIDED:`.
4. **`e2e/repo/detect/detect_test.go`** — delete the line
   `detectCase(t, "bun", 1, "bunfig.toml=[install]")` (about line 123). Run the
   case "detect: a tool with the evidence of a selected tool" (about line 95):
   it selects java and gives kotlin evidence; keep it as it is when it stays
   green, else report the result as a QUESTION.
5. **`e2e/repo/detect/selection_test.go`** —
   - In the case "add writes the dependencies" (about line 55), keep the two
     existing checks as they are, and add one check named
     "add: the kotlin dependency of android" with
     `c.TomlQ(`.tools | has("kotlin")`, "")`.
   - In the case "remove: the dependencies that no tool needs go" (about line
     109), add `[tools.kotlin]` to the values next to `[tools.android]`, and add
     the check "remove: kotlin goes" with `notGrep(c.ToolsValues, "kotlin")`.
   - Add one case after "remove: a tool that another selected tool needs":
     values `devValues + "\n[tools.android]\n[tools.kotlin]"`, call
     `removeTool("kotlin")`, code 2, message
     `kotlin is needed by android - remove android first` (copy the exact form
     of the android message at about line 118).
6. **`e2e/repo/files/packs_test.go`** — no change is expected: the mobile
   pack case with `--exclude android` keeps its message. Run the package; if a
   case fails because of the new closure, fix the case to the new closure and
   say so in `DECIDED:`.
7. **`e2e/repo/migrate/migrate_test.go`** — run the package. A case that adds
   android (about lines 443 and 461-536) can now log or write kotlin. Change
   only a check that fails because of the new closure, to expect kotlin.

## Verification

- `env CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins mise
  x -- go test ./e2e/repo/detect/ ./e2e/repo/files/ ./e2e/repo/migrate/`
  is green (from the repository root; the golden of `cli/` is U3's), and the
  migrate cases that add android ran (not skipped).
- The unit gate line `mise x -- mise run code:format --all` is green.
- `grep -rn '\bbun\b' repo/ e2e/` gives no hit.
- `grep -n depends repo/tools/android/tools.toml` shows `kotlin`.

## Guardrails

- Do not touch `cli/internal/source/testdata/toml-yq.golden`: U3 regenerates it.
  `go test ./cli/...` fails on it until then; that is expected.
- Do not touch any doc. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not touch `.config/mise/tasks/setup/deps/40-node.py`: it is the synced
  copy, and the after-landing sync writes it.
- Delete with `rm`, never `git rm`.

## Commit

`refactor: drop the bun tool, and android depends on kotlin`
