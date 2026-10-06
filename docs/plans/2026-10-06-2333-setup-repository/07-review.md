# U7 — Review: e2e harness, repository context, apply engine, output, options and init

- **Wave:** 5
- **Depends on:** U1, U2, U3, U4, U5, U6
- **Owns:** —
- **Model:** opus
- **Kind:** review

## Scope

Covers U1–U6: every code unit of this plan. Reviews the branch delta since the
branch base. One row per plan (assumed decision 8). Security focus: the hidden
`BOOTSTRAP_TEST_PAUSE_AFTER_WRITE` variable does nothing but pause, path
handling never follows a symlink or leaves the repository root, and no `--json`
output can carry anything but the documented keys.
