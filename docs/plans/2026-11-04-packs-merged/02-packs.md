# U2 — packs.toml in the new shape and the scanner dependencies

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `tools/_security_base/tools.toml` (read only).

## Ruling

From the design rulings of 2026-11-04:

> - One merged packs model for both sides (user's proposal, 2026-11-04):
>   - `packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
>     `<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
>     nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
>     key is `default`; any other stops with exit 2. No pack-level keys (no
>     `depends`, `tools`, `default = [...]`, `required`, `description`).
>   - `required` goes: a common tool (`_security_base`) becomes a `depends` of
>     each member tool (gitleaks, grype, osv-scanner), so a single-tool
>     selection is correct. Repo consumers get a `migrate.md` item for it.

From this plan:

D9: "In no pack. gitleaks, grype and osv-scanner each get depends =
[\"_security_base\"], so a pack selection and a single [tools.<scanner>] entry
both bring it, with the environments of the scanners that need it."

D11: "In file order, so the tool order and the pre-commit hook order of today
stay. An awk scan of each [pack.<name>] section reads the keys in file order
(one <key> = {...} line per member; a member that spans lines stops with code
2). gomplate still validates the shape. tools/packs.toml lists the base
members in the order of today: mise, repo-sync, git, pre-commit, dprint,
linter, vscode, then \"pack:security\"; security lists gitleaks, grype,
osv-scanner. A test proves that the tool order and the pre-commit hook order
do not change."

D12: "The dprint TOML plugin sorts keys (sortKeys: true in the global
config), which breaks D11. The root dprint.jsonc of this repository adds
tools/packs.toml to its excludes, and U2 owns dprint.jsonc. Plan 6 moves the
exclude with the file to repo/packs.toml."

## Edits

1. **`tools/packs.toml`** — keep the header comment and change its second
   sentence to the new shape (a member is `<tool> = { default = true }` or
   `<tool> = {}`; a nested pack is `"pack:<name>"`). Then:

   ```toml
   [pack.base]
   dprint = { default = true }
   git = { default = true }
   linter = { default = true }
   mise = { default = true }
   "pack:security" = { default = true }
   pre-commit = { default = true }
   repo-sync = { default = true }
   vscode = { default = true }

   [pack.security]
   gitleaks = { default = true }
   grype = { default = true }
   osv-scanner = { default = true }
   ```

   The member order is the order of the tools and of the pre-commit hooks
   (D11): keep it exactly. Each member is on one line.
2. **`dprint.jsonc`** (the root file of this repository) — add
   `"tools/packs.toml"` to `excludes`, with a comment: the member order of
   the packs is the sync order, and the TOML plugin sorts keys (D12).
3. **`tools/gitleaks/tools.toml`**, **`tools/grype/tools.toml`**,
   **`tools/osv-scanner/tools.toml`** — `depends = [ "_security_base" ]`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -nE 'depends|required|"p:' tools/packs.toml` finds nothing.
- After `code:format`, `grep -n '= {' tools/packs.toml` lists the base
  members in the order of D11.
- `grep -n '_security_base' tools/gitleaks/tools.toml tools/grype/tools.toml
  tools/osv-scanner/tools.toml`
  finds one `depends` line in each.

## Guardrails

- Touch only the five owned files.
- Do not add `_security_base` to a pack (D9).
- Delete with `rm`, never `git rm`.

## Commit

`ops: write packs.toml in the merged shape and make each scanner depend on
_security_base`
— written by the orchestrator after the unit gate and the wave gate, not by
the unit.
