# U4 — Docs

- **Read first:** the skill file `vwf:docs-sync`, then the passages below.
- **Lazy-load:** `docs/design/mise.md`, `docs/setup/migrate-dotfiles.md`.

## Ruling

Run `vwf:docs-sync` over the branch delta. Apply its findings, each
`DOCS FALSIFIED:` line of the earlier units, and the list below.

D1: "A repository at github.com/<anyone>/<repo> still loads its mise config with
no prompt, so the warning not to clone an untrusted repository there stays."

D6: "When setup_sync handles a target below .ssh/, it makes sure once in each
run
that the folder $_SS_HOME/.ssh has mode 700, on a new folder and on an existing
folder."

D8: "Accept these, with the reason of each in the table of 04-docs.md."

D9: "U4 closes only the security entries, each with its ruling in its text. A1
recounts the 9 rows after landing."

D13: "Accept all 12, with no code change. The user chose this. The entries of
the
first three folders record rules that the runs already applied. The
sync-engine-cache entries need a user who points SYNC_ENGINE_CACHE at the wrong
folder."

D11: "In a separate plan of the macos-setup repository,
docs/plans/2026-11-06-dotfiles-security. One run works in one repository."

D2, D3, D4, D5 and D7 are quoted in `01-mac-tool-values.md`.

## Edits

1. **`docs/setup/tools.md`, `claude-cli` section** — Change
   `merge, template, remove `deniedMcpServers`` (:183) to name
   `env.CONTEXT7_RUNNER` too.
2. **`docs/setup/tools.md`, `ssh` section** (:753) — Add one line: the sync
   keeps
   `~/.ssh` at mode 700.
3. **`docs/setup/tools.md`, `mise` section** — Add one line that states the
   three trust paths, and this warning: do not clone an untrusted repository
   below the two project folders. The doc does not name them today.
4. **`docs/design/setup-sync.md`** — Before "The run order", add a short section
   "The mode of `~/.ssh`" with the rule of D6. In step 7 of "Stages and exit 4"
   (:227), add the change of the mode of `~/.ssh` to the list of stage changes.
5. **The gaps of the nine earlier plans** — For each row of the table, run these
   commands, with `F=docs/plans/<folder>/index.json`, with `id`, `decision` and
   `row` from the row, and with `plan.mjs` the file
   `skills/plan-management/scripts/plan.mjs` of the vwf plugin of this run. Skip
   a row whose gap already has state `closed`, so a second run does not add the
   text again:

   ```text
   t=$(node plan.mjs get "$F" "gaps.$id.text" --raw)
   node plan.mjs set "$F" "gaps.$id.text" "$t RULING (2026-11-06-security-gaps $decision): $row"
   node plan.mjs set "$F" "gaps.$id.state" closed
   ```

| Folder `2026-11-04-`      | Gap | Decision | Text of the row                                                                                          |
| ------------------------- | --- | -------- | -------------------------------------------------------------------------------------------------------- |
| `macos-tools-base-core`   | G5  | D2       | Fixed: the Mac config no longer sets uv_venv_auto; uv creates the venv at uv run.                        |
| `macos-tools-base-core`   | G6  | D1       | Fixed: only ~/.config/mise and the github.com and gitlab.com folders are trusted.                        |
| `macos-tools-base-core`   | G7  | D1       | Accepted: a spec gap of the import; D1 now gives the requirement.                                        |
| `macos-tools-base-core`   | G8  | D3       | Partly fixed: the custom command runs git status only below the two folders; git_status stays, accepted. |
| `macos-tools-base-core`   | G9  | D4       | Fixed: the brew alias and updateall pin the user fnox config; CONTEXT7_RUNNER is deleted (D5).           |
| `macos-tools-base-core`   | G10 | D3       | Accepted: a spec gap of the import; D3 now gives the requirement.                                        |
| `macos-tools-base-shells` | G8  | D8       | Accepted: accept-new is the normal tradeoff for a single user Mac.                                       |
| `macos-tools-base-shells` | G9  | D8       | Accepted: the pin rule exempts dev tooling. The zinit ver ice stays available.                           |
| `macos-tools-base-shells` | G10 | D8       | Accepted: the attacker must already run as the user.                                                     |
| `macos-tools-base-shells` | G11 | D8       | Accepted: the attacker must already run as the user.                                                     |
| `macos-tools-base-shells` | G12 | D8       | Accepted: GitHub still shows the mode change, and a change shows mode differences in all repositories.   |
| `macos-tools-base-shells` | G13 | D6       | Fixed: the sync sets ~/.ssh to mode 700, on a new and on an existing folder.                             |
| `macos-tools-base-shells` | G14 | D6       | Fixed: D6 gives the requirement, and test:setup-sync covers it.                                          |
| `macos-tools-base-shells` | G15 | D8       | Accepted: the attacker must already run as the user.                                                     |
| `setup-sync-front-end`    | G5  | D8       | Accepted: the values file is trusted, and there is no privilege gain.                                    |
| `setup-sync-front-end`    | G7  | D8       | Accepted: the values file is trusted, and there is no privilege gain.                                    |
| `setup-sync-front-end`    | G8  | D8       | Accepted: the values file is trusted, and docs/design/repo-sync.md holds the rule for these templates.   |
| `macos-tools-extras`      | G7  | D8       | Accepted: the user uses skip-permissions on purpose.                                                     |
| `macos-tools-extras`      | G8  | D8       | Accepted: tailscale must run as root.                                                                    |
| `macos-tools-extras`      | G9  | D8       | Accepted: skip-permissions sessions do not use the allow list.                                           |
| `macos-tools-extras`      | G10 | D1       | Fixed: only ~/.config/mise and the github.com and gitlab.com folders are trusted.                        |
| `macos-tools-extras`      | G11 | D7       | Fixed: the task uses readlink -f, with no path prefix.                                                   |
| `macos-tools-extras`      | G12 | D8       | Accepted: the values file is trusted, and docs/design/setup-sync.md lists the exception.                 |
| `macos-tools-extras`      | G13 | D8       | Accepted: the values file is trusted, and docs/design/setup-sync.md lists the exception.                 |
| `macos-tools-extras`      | G14 | D8       | Accepted: the bootstrap source owns the key, and the live file lacks it.                                 |
| `macos-tools-apps`        | G2  | D8       | Accepted: the keep-awake assertion is deliberate.                                                        |
| `macos-tools-apps`        | G3  | D8       | Accepted: the user chooses auto-approve on purpose, as with skip-permissions.                            |
| `macos-tools-apps`        | G4  | D8       | Accepted: an opt-in tool from a tap that the user chose.                                                 |
| `macos-tools-apps`        | G5  | D8       | Accepted: an opt-in tool from a tap that the user chose.                                                 |
| `mise-generation-shared`  | G5  | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `mise-generation-shared`  | G6  | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-files-table`        | G12 | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-files-table`        | G15 | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-files-table`        | G16 | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-manifest`           | G8  | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-manifest`           | G9  | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-manifest`           | G10 | D13      | Accepted: the run already applied this rule. No code change.                                             |
| `repo-manifest`           | G11 | D13      | Accepted: the run already applied this label. No code change.                                            |
| `sync-engine-cache`       | G11 | D13      | Accepted: the user must point SYNC_ENGINE_CACHE at the wrong folder.                                     |
| `sync-engine-cache`       | G12 | D13      | Accepted: the user must point SYNC_ENGINE_CACHE at the wrong folder.                                     |
| `sync-engine-cache`       | G13 | D13      | Accepted: the user must point SYNC_ENGINE_CACHE at the wrong folder.                                     |

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `node plan.mjs validate "$F"` passes for each of the nine files.
- A count of the `state` values shows 41 fewer `open` entries over the nine
  files: core 6, shells 8, front-end 3, extras 8, apps 4, mise-generation-shared
  2, repo-files-table 3, repo-manifest 4, sync-engine-cache 3.
- Each of the 41 entries ends with its `RULING` text.
- `grep -n 'CONTEXT7_RUNNER' docs/setup/tools.md` shows the `remove` passage
  only.

## Guardrails

- Change only the passages that the change falsifies.
- Close only the 41 entries of the table. Leave each other gap of those folders
  open.
- Do not edit `docs/plans/index.json`. A1 recounts the rows after landing.
- Do not edit a gap of `2026-11-04-macos-entry-tool`. G5 and G6 are closed.
- Do not edit the nine `index.json` files with a text editor. Use only
  `plan.mjs`.
- Write in STE100.

## Commit

`docs: record the rulings on the security gaps`
