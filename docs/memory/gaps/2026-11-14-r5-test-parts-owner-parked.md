GAP 2026-11-14 ★3 | plan:2026-11-14-task-ports, phase:R5, plugin:stackgen,
repo/tools | who ships the test parts of a stack after R5 | parked: decide in
the R5 interview

**Source** The R1c interview of `docs/plans/2026-11-14-task-ports/` (D23). The
user asked whether `test/*` is managed by bootstrap or by claude-plugins, and
asked for this record.

What R1c decides:

- bootstrap owns the dispatchers: the mise tool ships `test/_default`,
  `test/e2e/_default` and `test/golden/_default`, the manifest lists them, and
  `init` and `check` protect their paths.
- A part (`NN-<name>`) belongs to whoever writes it: the repo itself (for
  example `test/50-local` after the `init` move), a stackgen pack, or a
  bootstrap tool. The manifest lists only bootstrap's files, so the sync never
  touches a part from the repo or from stackgen.

What is open: after R5 ("stack without packs"), who ships the test parts of a
stack, for example `go test ./...` for a Go project or the e2e runner of an
Android app? The design record says the packs go in R5, and stackgen keeps the
conventions, the skills and the LSP registration. By the split principle, a
fixed command is deterministic and fits a part of a bootstrap tool (for example
`test/20-go` from the go tool); a command that needs a judgment fits stackgen.
No ruling says this yet.

Related: `docs/memory/gaps/2026-11-14-pack-test-parts-parked.md` (the packs
must ship `NN-` parts until R5 removes them).
