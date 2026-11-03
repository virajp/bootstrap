# U4 — Docs

- **Read first:** every owned file that the list below names, at the named
  passages, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D4: "The user, by hand, after the landing. It is out of scope, as in the
earlier plans."

D5: "Port the rules of setup-sync as they are. A base or dependency tool is a
tool that pack base adds, also through a nested default = true pack
("pack:security": gitleaks, grype, osv-scanner, and _security_base through
depends), or a tool in the depends closure of a selected tool, a tool that a
mise depends adds included; a <tool> = {} member that the pack does not add
does not count. It is a base or dependency tool also when a [tools.<name>]
table names it. For such a tool, a list in [tools.<name>] appends its new
items, and a value whose type differs from the merged value stops the sync
with code 2. For any other tool, a value of [tools.<name>] replaces the
merged one, and tables merge key by key."

D7: "No. The engine validates only the shape of [questions]. A template that
needs an answer uses required and fails that file with its message. Exit 4
stays Mac-only."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`**: the `tools.toml` keys bullet ("Each tool folder has a
     `tools.toml` with ... keys") adds the optional `[defaults]`,
     `[questions.<key>]` and `[contributes.<tool>]`. "How the sync operates":
     "The only key of an entry is `env`" becomes: `env` selects, and each
     other key of a `[tools.<name>]` table is a value; a `[packs.<name>]`
     entry has only `env`. "Templates": add `.config` and `.all`, the merge
     order, the list rules of D5, and that a missing answer is not an engine
     error (D7).
   - **`docs/design/tools.md`**: the three tables, their keys and their
     checks.
   - **`docs/design/bootstrap-config.md`**: `[tools.<name>]` holds values.
   - **`docs/sync/index.md`** step 3: a tool value goes into `[tools.<name>]`;
     the top-level keys stay.
   - **`README.md`**: each passage that says `env` is the only key.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files.
3. No `docs/sync/migrate.md` item: a values file that works today works
   unchanged.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -l 'contributes' CLAUDE.md docs/design/tools.md` finds both files.
- `grep -rn 'the only key is .env' CLAUDE.md README.md docs/sync docs/design`
  finds no passage about `[tools.<name>]`.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Write identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe tool values, questions and contributions` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
