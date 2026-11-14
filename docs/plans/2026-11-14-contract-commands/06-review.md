# U6 — Review: the engine and the tests

## Scope

U1 to U5, the whole branch delta since the branch base. The engine is Go code
that ships in the binary, and the bash cases run in CI (D17). Focus: stdout
holds only the envelope under `--json`, and every human line goes to stderr
(D1); the exit code equals the envelope `code` (D12); no write command prompts
or writes a file under `--json` without `--yes` or `--dry-run` (D4); the
output of every command without `--json` does not change, except
`questions --json` (D2); the fetch rule (D5, D6), above all the 7-day limit and
code 5; `describe` and its call to mise (D19).
