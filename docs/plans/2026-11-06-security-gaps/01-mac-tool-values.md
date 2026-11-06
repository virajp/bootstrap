# U1 — Mac tool values

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `docs/design/mise.md` and `docs/design/tools.md`, for the rules
  of a `tools.toml`. `docs/design/setup-sync.md`, for the `macos-setup` task.

## Ruling

D1: "The value is $HOME/.config/mise, <projects_dir>/github.com and
<projects_dir>/gitlab.com. The user chose these paths. No new macos.toml key.
Only the sibling folders of projects_dir (backup, copilot, tmp) lose trust. A
repository at github.com/<anyone>/<repo> still loads its mise config with no
prompt, so the warning not to clone an untrusted repository there stays."

D2: "Delete python.uv_venv_auto = \"create|source\" from
macos/tools/python/tools.toml. uv creates the project environment at uv run. The
user accepted this, because the user-level config does not create the venv.
repo/tools/python/tools.toml keeps the setting, so a repository that adopts the
python tool still opts in. This protects only outside the trusted paths of D1."

D3: "Only when the top folder of the repository is below
$PROJECTS_DIR/github.com or $PROJECTS_DIR/gitlab.com. Add case \"$top/\" in
\"$PROJECTS_DIR\"/github.com/*|\"$PROJECTS_DIR\"/gitlab.com/*) ;; *) exit 1 ;;
esac && before git status. Outside those folders, for example in an extracted
archive in Downloads, no git status runs and no mark shows. With PROJECTS_DIR
unset, no git status runs. An archive extracted inside the two folders can still
run code: accepted. git prints a real path, so a projects_dir with a symlink in
its path loses the mark: accepted. This guard covers only
custom.submodule_status. The git_status module of starship stays, and it can
still run the filter driver of a repository in any folder, as git status by hand
does: accepted. A projects_dir with a trailing slash also loses the mark:
accepted."

D4: "Pin the user config: brew = \"fnox -c $HOME/.config/fnox/config.toml exec
-P brew -- brew\", and fnox -c \"$HOME/.config/fnox/config.toml\" exec --profile
brew -- mise run upgrade:brew in updateall/_default. mise writes the alias in
single quotes in zsh, bash and fish, so $HOME expands when the alias runs. K2
tests this. The alias needs no template."

D5: "Delete both. The user removed context7 from vwf and uses the official
plugin
of Claude Desktop. Delete the CONTEXT7_RUNNER line of settings.json.tmpl and the
[profiles.context7.secrets.CONTEXT7_API_KEY] table of the fnox config. Add [
\"env\", \"CONTEXT7_RUNNER\" ] to remove of .claude/settings.json in
claude-cli/tools.toml, because a merge file keeps a deleted key unless remove
names it. Keep the mcp__plugin_context7_context7__* allow lines. The context7
addon of stackgen in claude-plugins keeps ${CONTEXT7_RUNNER:-pnpm dlx}: out of
scope, because no repository uses it today."

D7: "keg=\"$(dirname \"$(dirname \"$(readlink -f \"$(command -v
tailscale)\")\")\")\". readlink -l gives illegal option on this Mac, and
readlink
-f gives an absolute path, so the /opt/homebrew/bin/ prefix goes away. If keg
does not match */Cellar/tailscale/*, the task prints a warning, skips the chown
and continues the upgrade, because the service is already stopped and a chown of
another folder is not safe."

D12: "No. Those rulings applied to the import. A security fix after the import
is
a new change. These value differences from macos-setup are the drift that the
user approved, and the macos-setup plan brings the live dotfiles to the same
values."

## Edits

1. **`macos/tools/mise/.config/zsh/01b-mise.zsh.tmpl`:6** — Change the line to:

   ```text
   export MISE_TRUSTED_CONFIG_PATHS="$HOME/.config/mise:<% $pd %>/github.com:<% $pd %>/gitlab.com"
   ```

2. **`macos/tools/mise/.config/bash/01b-mise.bash.tmpl`:10** — Make the same
   change as edit 1.
3. **`macos/tools/mise/.config/fish/conf.d/06-mise.fish.tmpl`:9** — Change the
   line to:

   ```text
   set --global --export MISE_TRUSTED_CONFIG_PATHS "$HOME/.config/mise:<% $pd %>/github.com:<% $pd %>/gitlab.com"
   ```

4. **`macos/tools/claude-cli/.claude/settings.json.tmpl`** — Delete the line
   `"CONTEXT7_RUNNER": "fnox exec -P context7 -- pnpm dlx",` (:52). Change the
   value of `MISE_TRUSTED_CONFIG_PATHS` (:55) to:

   ```text
   <% .env.HOME %>/.config/mise:<% $projects %>/github.com:<% $projects %>/gitlab.com
   ```

5. **`macos/tools/claude-cli/tools.toml`:16** — Change the line to
   `remove = [ [ "deniedMcpServers" ], [ "env", "CONTEXT7_RUNNER" ] ]`.
6. **`macos/tools/python/tools.toml`:10** — Delete the line
   `python.uv_venv_auto = "create|source"`. The other keys of `[mise.settings]`
   stay.
7. **`macos/tools/starship/.config/starship.toml`:147** — Change the line to:

   ```text
   top=$(git rev-parse --show-toplevel 2>/dev/null) && [ -f "$top/.gitmodules" ] && case "$top/" in "$PROJECTS_DIR"/github.com/*|"$PROJECTS_DIR"/gitlab.com/*) ;; *) exit 1 ;; esac && git -C "$top" status --porcelain=v2 --ignore-submodules=none 2>/dev/null | awk '$1=="1" && $3 ~ /^S/{f=1} END{exit !f}'
   ```

8. **`macos/tools/homebrew/tools.toml`:25** — Change the line to
   `brew = "fnox -c $HOME/.config/fnox/config.toml exec -P brew -- brew"`.
9. **`macos/tools/macos-setup/.config/mise/tasks/updateall/_default`:69** —
   Change the line to:

   ```text
   fnox -c "$HOME/.config/fnox/config.toml" exec --profile brew -- mise run upgrade:brew
   ```

10. **`macos/tools/fnox/.config/fnox/config.toml`:29-32** — Delete the table
    `[profiles.context7.secrets.CONTEXT7_API_KEY]`, its three keys and the empty
    line after them.
11. **`macos/tools/fnox/tools.toml`:2** — Change `brew, context7, fnox` to
    `brew, fnox`. Keep the comment at its line width.
12. **`macos/tools/tailscale/.config/mise/tasks/upgrade/tailscale`:44-46** —
    Replace the three lines from `keg=` through `sudo chown` with:

    ```text
    keg="$(dirname "$(dirname "$(readlink -f "$(command -v tailscale)")")")"
    if [[ $keg == */Cellar/tailscale/* ]]; then
      print_subheader "Restoring ownership of ${keg} ..."
      sudo chown -R "${USER}:staff" "${keg}"
    else
      print_warn "Cannot find the Tailscale keg: skipping the ownership step"
    fi
    ```

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test:macos-tools` passes.
- `grep -rn 'CONTEXT7_RUNNER\|profiles.context7' macos/tools` shows only the
  `remove` line of `macos/tools/claude-cli/tools.toml`.
- `grep -rn 'uv_venv_auto\|readlink -l' macos/tools` shows no line.
- `grep -n 'uv_venv_auto' repo/tools/python/tools.toml` still shows line 21.

## Guardrails

- Do not touch `repo/`, `sync-engine/` or a doc.
- Keep the `mcp__plugin_context7_context7__*` allow lines.
- Do not put `~` in a trust path. mise must get a full path.
- Do not add a `macos.toml` key.
- Keep the `Synced from bootstrap` header of each file.

## Commit

`fix: narrow the trust paths and pin fnox in the Mac tools`
