# U2 — The test tasks: the cases of the new keys

- **Read first:** `.config/mise/tasks/test/repo` (the helpers at the top, the
  packs cases near `:1780-1809`, the questions cases near `:3001`), and
  `.config/mise/tasks/test/mac` (the helpers, the packs cases near `:330-338`).
- **Lazy-load:** `macos/fixtures/`, `CLAUDE.md` ("Tests").

## Ruling

The decisions D1–D6 and D14 of `index.json`, as the code of U1 implements them.
Each
rule gets at least one case. Each case compares the exit code and the message,
in the form of the existing cases.

## Edits

1. **`.config/mise/tasks/test/repo`** — add cases with a scratch clone, as the
   packs cases do:
   - D1: a `summary` of 48 characters (code 2), a `summary` that is not a
     string, a `selectable` that is not a boolean, a valid `summary` and
     `description` (code 0).
   - D2: a tool with a `category` and no `repo/categories.toml` (code 2); an
     unknown category id (code 2); a `[category.<id>]` with another key
     (code 2); a `[category.<id>]` with no `title` (code 2); a top-level key
     other than `[category]` (code 2); a valid file and category (code 0).
   - D3: `description = "..."` in a `[pack.<name>]` (code 0); a
     `description` that is not a string (code 2). Keep `packs_member_string`
     as it is: it must still give its message.
   - D4: `bootstrap add` of a tool with `hosts` that does not name the host of
     the test (code 2, the message names the tool and the host); the same
     tool in a values file with `bootstrap sync --dry-run` (code 0); a bad
     host name (code 2); `hosts = []` (code 2); `bootstrap replace` with the
     tool as its new tool (code 2).
   - D5: `bootstrap add` and `bootstrap init` of a tool with
     `selectable = false` (code 2, the message names the tools that need it);
     a values file that names it, with `bootstrap sync --dry-run` (code 0) and
     `bootstrap check` (the warning, and the exit code of today);
     `bootstrap replace` with the tool as its new tool (code 2); a
     not-selectable tool that no tool depends on (the short message); `bootstrap
     detect --dry-run` does not propose
     a tool with `selectable = false` or a host that it leaves out; the text of
     the warning of D13 e; `init --yes` again on a values file that answers a
     group (code 0).
   - D6: a group with all answers missing (code 2, "the group <tool>.<group>
     has no answer in <file> - run bootstrap init"); a group with one answer
     (code 0);
     an `optional` question with no answer (code 0); a question with `group`
     and `optional` (code 2); `bootstrap init --yes` with a group and no
     `--answer` (code 2); D14: `bootstrap init --yes` with an `optional`
     question and no flag (code 2, the hint names `--answer` and
     `--skip-answer`); `--skip-answer` on an `optional` question (code 0, `""`
     written; a second `init --yes` with no flag passes); `--answer <q>=` on an
     `optional` question (code 0, `""`
     written); `--skip-answer` and `--answer <q>=` on a question with no
     `group` and no `optional` (code 2); `--skip-answer` on `bootstrap add`;
     `init --yes --skip-answer` on each question of a group (code 2, "the group
     <tool>.<group> has no answer" and the hint of D14 (9) c);
     `--skip-answer <q>` with `--answer <q>=<value>` (code 2, "--answer and
     --skip-answer both name <q>");
     `bootstrap questions --json` shows `group`, `note`
     and `optional`; the text output of `bootstrap questions` for an optional
     question with no answer and an answer `""`; the changed key list of the
     question message at `test/repo:2359` and `:2411`.
2. **`.config/mise/tasks/test/mac`** and **`macos/fixtures/`** — the Mac side:
   `summary`, `category` and `selectable` are accepted (code 0); `hosts` on a
   Mac tool gives code 2 ("hosts is for repo tools only"); the changed key list
   of the unknown-key message; a `macos/categories.toml` case; a pack
   `description`; a group question with no answer (code 2).
3. Change each existing case whose expected message U1 changed (the key list
   of the Mac unknown-key message, the key list of the question message).

The test tasks have no terminal, so they do not test the prompt. The Go tests
of U1 test the field model.

Write the shell for bash 3.2 and bash 5.x (see `CLAUDE.md`, "Task scripts").

## Verification

- `mise x -- mise run test:repo` and `mise x -- mise run test:mac` pass.
- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Do not touch `cli/`, `repo/tools/`, `macos/tools/` or a doc.
- Make the host of a D4 case independent of the machine: give the tool a
  `hosts` list without the host of the test, for example from `uname`.

## Commit

`test: add the cases of the catalog keys`
