# U3 — Docs

- **Read first:** `CLAUDE.md`, `README.md`, `docs/sync/index.md`, every file
  of `docs/design/` as U1 left it, and the `DOCS FALSIFIED:` and
  `DECIDED:` lines of U1 and U2.
- **Lazy-load:** the other files of `docs/sync/`, the diffs of U1 and U2
  (`git log -p` of the branch).

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Contract docs: one place for each rule. `docs/design/` holds the full
contract: [...] More files are fine, but CLAUDE.md and `docs/sync/index.md`
load them lazily ("read X when you change Y"). CLAUDE.md keeps layout,
change steps, task rules, test rule, pointers. [...]"

"Tests: [...] CLAUDE.md rule: `sync-engine/` → `mise run test`; `repo/` →
`test:repo-sync`; `macos/` → `test:setup-sync` + `test:macos-tools`. No test
in hooks or in `code:merge` (a shared task; consumers have no such tests)."

From this plan:

D7: "A section "Where the rules are" with one line for each area: "Read
<doc> before you change <path>." The sections How the sync operates,
Templates, Branch model in the merge tasks and mise config go, after U3
confirms that docs/design holds each of their rules. Purpose, Layout rule
(folders only), Task scripts, Task output, Change a shared file and the test
rule stay. The rules for an edit in this repo also stay: the last_sync.toml
and commit_id steps, the file conventions of docs/sync/ (claude-tool.md),
and "do not edit a synced copy". The header authoring rule moves to
docs/design/tools.md, with a pointer in CLAUDE.md."

D8: "git log <commit_id>..HEAD -- tools repo docs/sync docs/design CLAUDE.md,
with tools kept in front of repo, because the rules moved into docs/design.
U3 writes it in repo/examples/last_sync.toml and in CLAUDE.md, and owns
repo/examples/last_sync.toml for that edit."

D6: "Every instruction that the adopting AI acts on stays in docs/sync/. A
passage that only restates a rule of docs/design becomes a link to it.
docs/sync/index.md gets a short list of pointers to docs/design, in the form
that CLAUDE.md uses."

D3: "sync-engine.md: the rules that both sides share [...]. Each other doc
links to the one place."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-macos-tools-apps."

## Edits

1. **`CLAUDE.md`**
   - For each rule of the sections How the sync operates, Templates, Branch
     model in the merge tasks and mise config, and of the `tools.toml` key
     rules in Layout rule: find it in `docs/design/`. When a rule is
     missing there, add it to the design doc that D3 names, then remove it
     here.
   - Remove those four sections and the key rules. Layout rule keeps the
     folders (`sync-engine/`, `repo/`, `macos/`, `docs/`, `examples`, the
     forwarding file) and what each holds.
   - Keep the rules for an edit in this repo (D7): the `last_sync.toml` and
     `commit_id` steps, the file conventions of `docs/sync/` (for example
     `claude-tool.md`), and "do not edit a synced copy".
   - Move the header authoring rule (the `Synced from bootstrap` header of
     each file that is neither a seed nor a merge source) to
     `docs/design/tools.md` when U1 did not, and leave one pointer line to it
     in `CLAUDE.md`.
   - The example-review command becomes
     `git log <commit_id>..HEAD -- tools repo docs/sync docs/design CLAUDE.md`
     (D8).
   - Add a section `## Where the rules are`, after Layout rule, with one line
     for each area (D7):
     - `sync-engine/` → `docs/design/sync-engine.md`;
     - `sync-engine/repo-sync`, `repo/tools/repo-sync/` →
       `docs/design/repo-sync.md`;
     - `sync-engine/setup-sync`, `macos/tools/macos-setup/` →
       `docs/design/setup-sync.md`;
     - `repo/packs.toml`, `macos/packs.toml` → `docs/design/packs.md`;
     - a `tools.toml` → `docs/design/tools.md` and `docs/design/mise.md`;
     - the keys of `.config/bootstrap.toml` →
       `docs/design/bootstrap-config.md`;
     - fnox → `docs/design/fnox.md`;
     - `docs/sync/` → `docs/sync/index.md`.
   - Keep Task scripts, Task output and Change a shared file; replace in
     them each restated engine rule with a pointer. Keep the test rule of
     the ruling as it is.
2. **`docs/sync/index.md`** — add a short pointer list to `docs/design`, in
   the form of the `CLAUDE.md` list (D6), near "Files in this folder" or
   "Rules". Replace each passage that only restates a design rule with a
   link. Keep every instruction and every step.
3. **Other files of `docs/sync/`** — apply D6 only where docs-sync finds a
   restated rule.
4. **`README.md`** — its sections that restate engine rules (for example
   How a repository syncs, Templates) keep a short summary and link to the
   design doc.
5. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line inside the owned
   files. The `GAP:` lines of U2 (a header rule that a design doc lacks)
   stay as they are; this unit does not edit `sync-engine/`.
6. **`repo/examples/last_sync.toml`** — in its comment, the review command
   becomes `git log <commit_id>..HEAD -- tools repo docs/sync docs/design
   CLAUDE.md`
   (D8). Do not change `commit_id`.
7. **No `docs/sync/migrate.md` item.** No consumer file changes.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.
- `grep -nE '^## (How the sync operates|Templates|Branch model in the merge
  tasks|mise config)' CLAUDE.md`
  finds nothing; `grep -n '^## Where the rules are' CLAUDE.md` finds one
  line.
- Each `docs/design/*.md` path that the `CLAUDE.md` list names exists.
- `grep -c 'docs/design/' docs/sync/index.md` is at least 3.
- `wc -l CLAUDE.md` is smaller than at the branch base.
- `grep -nF 'tools repo docs/sync docs/design CLAUDE.md' CLAUDE.md
  repo/examples/last_sync.toml`
  finds a line in each file.

## Guardrails

- Touch only the owned files. Change `docs/design/**` only to add a rule
  that this unit moves out of `CLAUDE.md` or that docs-sync names as
  missing. Do not touch `sync-engine/`.
- Do not change a rule. Do not remove an instruction from `docs/sync/`.
- Write new prose in ASD-STE100 Simplified Technical English. Write
  identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: load the design docs lazily from CLAUDE.md and docs/sync` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
