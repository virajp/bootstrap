# U5 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (its case helpers, its
  scratch repos and how a case runs the engine; the bare-array cases at about
  lines 3198-3199 and 3836-3844), `.config/mise/tasks/test/mac` (the
  bare-array cases at about lines 1992 and 2428), and the commands that U1 to
  U4 add.
- **Lazy-load:** `cli/internal/contract/` (the envelope fields).

## Ruling

D1, D2, D4, D8, D9 and D12 are quoted in the unit files of U1 to U3; D5, D10
and D14 in `04-catalog-describe.md`. This unit tests the built engine end to
end, the way the other cases of `test/repo` do.

- **D26:** the envelope is compact: one line, then a newline. Rewrite the
  bare-array cases of `test/repo` (about 3198-3199, 3836-3844) and
  `test/mac` (about 1992, 2428) for the envelope (D2, D20).

## Edits

1. **`.config/mise/tasks/test/mac`** — rewrite the two `questions --json`
   cases for the envelope; add three cases: `mac add tool <name> --json`,
   `mac check --json` and `mac describe --json` each exit 2 (D46).
2. **`.config/mise/tasks/test/repo`** — rewrite the two bare-array cases, and
   add cases, each in the style of the
   cases beside it, with `jq` only if the file already uses it, else with
   `grep` on the compact JSON:
   1. `contract --json` outside a repo: exit 0, `"contract":1`, the 14 ports.
   2. `questions --json`: the envelope with `"command":"questions"`.
   3. `config get repo.repo_name --json`: the envelope; `config get ...
      --json --format toml`: exit 2.
   4. `check --json` on a scratch repo with an unknown key in
      `.config/bootstrap/repo.toml`: exit 2, a `fail` stage, a `skipped`
      stage.
   5. `add tool <name> --json` with neither flag: exit 2 and no change in git;
      with `--dry-run`: exit 0, a plan, no outcomes (D34). Commit the scratch
      repo after its setup, so "no change" is against that commit.
   6. `catalog --json --clone <clone>`: exit 0, `"fetched_at":""`.
   7. `describe --json` on a scratch repo: `repo`, `tools`, `members` and
      `ports` keys.
   8. An unknown flag with `--json`: exit 2 and the envelope.

## Verification

- `mise x -- mise run test:repo` and `mise x -- mise run test:mac` pass.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo` and
  `.config/mise/tasks/test/mac`.
- No case needs the network: use `--clone` with the worktree.
- Delete with `rm`, never `git rm`.

## Commit

`test: bash cases for the JSON contract commands`
