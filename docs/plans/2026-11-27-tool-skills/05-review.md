# U5 — Review: the hook, the templates and the merge spec

## Scope

U1 to U4: the branch delta since the branch base. D11: the hook script
(`macos/tools/claude/files/.claude/hooks/npm-normalize.sh`) and the templates
are runnable, so this review runs `/code-review` and `/security-review` and the
two reviewers. Pay attention to the hook: it runs on each Bash command of
Claude Code, so a quoting defect changes a command, and it must never change a
command when pnpm is not on PATH.
