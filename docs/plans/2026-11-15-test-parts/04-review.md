# U4 — Review: the engine, the parts and the tests

## Scope

U1, U2 and U3, the whole branch delta since the branch base. Focus: `init`
removes only a file whose hash matches (any of the released versions, D13) and
whose tool is selected (D5); the
removal runs before R1c's move (D5); a part that does not apply exits 0 with
no output (D6); no part passes an environment value or a path to a shell
unquoted; the parts declare only `--record` (D3).
