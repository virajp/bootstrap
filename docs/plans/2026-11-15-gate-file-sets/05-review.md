# U5 — Review: the helper, the gates and the tests

## Scope

D18: the plan lands runnable code (a Python helper and bash gates that each
consumer runs). U1, U2, U3 and U4, the whole branch delta since the branch base.
Focus: the
helper never lists a path of a submodule or a nested repository (D16); no
gate keeps an own filter, a `git ls-files` call or `--force-exclude` (D7);
every gate passes the list file path quoted and gives it to its tool per D14;
a gate with an empty list prints nothing and exits 0; the helper parses a
`#FILES` header and `config-files.toml` without running their content; the
shell is valid for bash 3.2; the Python uses the standard library only.
