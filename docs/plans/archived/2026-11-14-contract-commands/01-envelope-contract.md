# U1 — The envelope, contract, and --json on questions and config

- **Read first:** `cli/internal/cli/cli.go` (`Usage`, the flag-allow table,
  `parse`, the `switch cmd`), `cli/internal/cli/init.go` (the `questions`
  command and its `--json`, about line 380 before R1a), the `config` verb that
  R1a adds (`cli/internal/cli/config.go` or the file R1a names), and
  `cli/internal/fail/fail.go`.
- **Lazy-load:** `cli/cmd/bootstrap/main.go` (`main.version`, set by
  ldflags), `cli/internal/oldform/` (the pattern for a format integer),
  `cli/cmd/bootstrap/main_test.go` (usage strings), `docs/design/cli.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** every `--json` command prints one object on stdout:
  `{"contract": 1, "command", "ok", "code", "data", "errors", "warnings"}`.
  All human lines go to stderr. `data` is the per-command payload.
- **D7:** the contract version is one integer, `1` in this plan. A new field,
  command or enum value keeps it. A removed or renamed field, or a changed type
  or meaning, raises it.
- **D11:** each item of `errors` and `warnings` is
  `{"code", "message", "path"}`: `code` is the `fail.Code` (a warning has 0),
  `message` is the text a person sees, and `path` is the file when one is known
  and is left out otherwise.
- **D12:** with `--json` the process exit code equals the envelope `code`, and
  `ok` is true only when `code` is 0. A usage error (an unknown flag, a missing
  argument, `--json` and `--format` together) also prints the envelope, with
  code 2 and `command` set to the command name, or `""` when no command was
  parsed.
- **D13:** `bootstrap contract --json` prints `data` =
  `{"versions": [1], "cli_version", "commands", "ports"}`: `cli_version` is
  `main.version` (D25; a local build reports `"dev"`), `commands` is the names
  of every command that takes `--json`, and `ports` is the D3 names only, with
  no `present` flag (D24). `contract`
  takes no repo and works outside one. Without `--json`, `contract` prints the
  same facts as human lines.
- **D3:** one Go table of port names: `code:format`, `code:lint`, `code:sec`,
  `code:precommit`, `git:merge:<branch>`, `git:worktree:create`,
  `git:worktree:setup`, `git:worktree:delete`, `code:lsp:check`, `setup`,
  `test`, `test:e2e`, `test:golden`, `p:<id>:deploy`. This unit adds the table;
  U4 reports `present` for each port in `describe`.
- **D2, D20:** `questions --json` and `mac questions --json` (the same
  `printQuestions`) print the D1 envelope with the question array in `data`.
  The bare array goes away. The flag table is shared with the Mac side:
  `--json` on the `mac` write commands is refused with code 2.
- **D26:** the envelope is compact: one line, then a newline.
- **D31:** `mac check --json` is refused with code 2 too.
- **D37:** a list in the envelope is always `[]`, never `null` (write a nil
  slice as `[]`). When a command fails before it has a result, `data` is
  `null`.
- **D38:** before the parse, the engine looks for a literal `--json` anywhere
  in the arguments. If it finds one, every usage error prints the envelope
  with code 2, with `command` `""` when no command is known. `--json` on a
  command that does not take it (`list`, `tui`) is such a usage error.
- **D46:** on the Mac side only `mac questions` accepts `--json` (D20). Every
  other `mac` command with `--json` exits 2 with the envelope (D38).
  `catalog`, `describe` and `contract` are repo-side commands only; under
  `mac` each is an unknown command, a usage error with code 2.
- **D33:** `command` is the full command words joined by one space:
  `"config get"`, `"mac questions"`. `contract` lists the same full forms in
  `commands`.
- **D9:** `config get`, `show`, `set` and `unset` take `--json`, which prints
  the D1 envelope with R1a's output object in `data`. `--format` without
  `--json` stays as R1a made it. `--json` and `--format` together is a usage
  error (code 2).
- **D15:** no new dependency; use `encoding/json`.

## Edits

1. **A new package `cli/internal/contract`** — the `Version` constant (1), the
   `Supported` list (`[1]`), the envelope type with the D1 fields in that
   order, the item type of D11, the port table of D3, and one function that
   writes an envelope to an `io.Writer` and returns the exit code (D12). Write
   its Go tests first: field order, `ok` against `code`, a `path` left out when
   empty.
2. **`cli/internal/cli/cli.go`** — add the `contract` command and its usage
   line; allow `--json` on `contract`, `questions` and `config`; make a usage
   error under `--json` print the envelope (D12). Keep every human line on
   stderr when `--json` is set.
3. **`cli/internal/cli/init.go`** (`printQuestions`, or the file that owns it
   after R1a) — wrap the `questions --json` array in the envelope, for the
   repo side and the Mac side (D2, D20).
4. **The `config` verb of R1a** — add `--json` to `get`, `show`, `set` and
   `unset` (D9); refuse `--json` with `--format` (code 2).
5. **The Go tests beside each changed file, and the usage strings in
   `cli/cmd/bootstrap/main_test.go`** — add cases for `contract --json`,
   `questions --json`, `config get --json`, a usage error under `--json`, and
   `mac add --json` (code 2). The bash cases of the bare array are U5's (D26).

## Verification

- `mise x -- mise run test:go` passes.
- `go run ./cmd/bootstrap contract --json` in `cli/` prints one JSON object
  whose `contract` is 1 and whose `data.ports` holds the 14 names of D3.
- `grep -rn '"contract"' cli/internal/contract` finds the field.

## Guardrails

- Touch nothing outside `cli/`. Docs are U7's.
- Do not change the output of a command without `--json`, except `contract`,
  which is new.
- Do not bump `cli/VERSION`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the JSON envelope, bootstrap contract, and --json on questions and
config`
