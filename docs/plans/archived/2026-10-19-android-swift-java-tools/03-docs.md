# U3 — Docs: reconcile the docs with the new tools and vwf.md

- **Wave:** 2
- **Depends on:** U1, U2
- **Owns:** `README.md`, `CLAUDE.md`, `docs/sync/index.md`,
  `docs/sync/migrate.md`, and any other doc that `vwf:docs-sync` finds
  falsified, except the files that U1 and U2 own
- **Model:** opus
- **Kind:** edit
- **Read first:** every owned file you edit, top to bottom; the two unit
  reports' `DOCS FALSIFIED:` lines, which the orchestrator passes in.
- **Lazy-load:** `docs/sync/java.md`, `docs/sync/android.md`,
  `docs/sync/swift.md`, `docs/sync/vwf.md` (read only, to quote them
  correctly).

## Ruling

From index.md, assumed decisions:

> D12 — `docs/sync/<tool>.md` and `docs/sync/vwf.md` are shipped instructions
> that agents run, so the edit units own them. The docs unit owns
> `README.md`, `CLAUDE.md`, `docs/sync/index.md` and `docs/sync/migrate.md`.

> D4 — `tools/flutter/tools.toml` drops `[mise.tools.android-sdk]` and
> `[mise.tools.java]` and sets `depends = ["android"]`.

> D2 — A new `java` tool. `tools/java/tools.toml` defines only
> `[mise.tools.java]`, `version = "lts"`, `envs = ["dev", "ci"]`.

> D3 — A new `android` tool defines `[mise.tools.android-sdk]`,
> `version = "23"`, `envs = ["dev", "ci"]`, mise `depends = ["java"]`.

## Edits

1. Run `vwf:docs-sync` over the run's branch delta (the skill file is
   `/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply its findings that fall inside this unit's Owns, plus every
   `DOCS FALSIFIED:` line from U1 and U2, plus the items below.
2. **`docs/sync/index.md`**:
   - the tool list (:32-36): add `android.md`, `java.md` and `swift.md` in
     alphabetical order;
   - the files table (:24-30): add a row for `vwf.md` — the entry point for
     the vwf and stackgen plugins, which give it `add tools` and
     `remove tools` instructions;
   - any example that shows `android-sdk` or `java` in `flutter`.
3. **`README.md`** — the `tools.toml` example at :44-60 shows `flutter`
   defining `android-sdk` with `depends = ["java"]`. After this plan, that is
   false. Change the example so that it is true of a real tool: either the
   `android` tool (`android-sdk`, `depends = ["java"]`, envs dev and ci) or
   keep the example generic and say that it is an illustration. Keep the
   `[mise_env.ci.env]` illustration only if it still says it is an example.
   Add one line that names `docs/sync/vwf.md` as the entry point for the vwf
   and stackgen plugins, where the README describes `docs/sync/`.
4. **`CLAUDE.md`** — section "Layout rule": add `docs/sync/vwf.md` to the
   list of files of `docs/sync/`: the entry point for the vwf and stackgen
   plugins; a plugin names this path and no other bootstrap path.
5. **`docs/sync/migrate.md`** — only if it names `android-sdk` or `java` as
   part of `flutter`, or lists the tools; make it true. Otherwise do not
   change it.

## Verification

- Wave-gate lines 1 and 2 (index.md, Wave gate) pass.
- `grep -n 'vwf.md' docs/sync/index.md CLAUDE.md` prints at least one line
  for each file.
- `grep -n 'android.md\|java.md\|swift.md' docs/sync/index.md` prints the
  tool list line(s).
- No passage in the owned files says that `flutter` defines `android-sdk` or
  `java`.

## Guardrails

- Do not touch `tools/**`, `docs/sync/vwf.md` or the `docs/sync/<tool>.md`
  files that U1 owns. If one of them is wrong, return an `UNRESOLVED:` line
  that names the file and the passage.
- Do not edit the synced copies at the repository root (`.config/`,
  `.gitignore`).
- Delete with `rm`, never `git rm`.
- STE prose, as in the existing docs.

## Commit

`docs: describe the java, android and swift tools and vwf.md`
