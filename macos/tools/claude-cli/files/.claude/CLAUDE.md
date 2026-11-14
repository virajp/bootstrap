# User-Level CLAUDE.md

Global conventions for all projects. Project-level CLAUDE.md overrides these.
Bias toward caution over speed — use judgment for trivial tasks.

## Stack

- **fish**: This laptop uses fish shell. It also has `bash` and `zsh` installed,
  but fish is the default.
- **mise**: This laptop uses `mise` for tools installation, task management, and
  project environment management. It also has `homebrew` installed, but it only
  used for non-development tools or tools unavailable in `mise`. Use `mise` for
  all development tools.

## Hard Rules

- **git**: Always use the `git-workflow` skill for all git interactions
- **tasks**: Always prefer `mise` — list available tasks with `mise tasks`
  - Task scripts live in `.config/mise/tasks/` at the repo root for repo-level
    tasks, and `~/.config/mise/tasks/` for global tasks
  - Only the `init` task is defined in mise's config file — it just makes the
    task scripts executable
- **libraries**: Always use Context7 MCP (`resolve-library-id` →
  `get-library-docs`) before writing code with any external library — never rely
  on training knowledge for APIs or config schemas
- **shell**: core commands are aliased to modern replacements (`ls`→`eza`,
  `cd`→`z`, `cat`→`bat`) — prefix with `command` whenever output is parsed or
  standard flags are needed. See [Shell Aliases](#shell-aliases)
- ALWAYS ask questions, NEVER assume

## Shell Aliases

Defined only in the `[shell_alias]` blocks of
`~/.config/mise/conf.d/<tool>/mise.toml` (one for each Mac tool, rendered from
`macos/tools/<tool>/` of bootstrap by `mise run updateall:configs`; the
shadowing ones are in the folders `cli`, `doggo`, `eza`, `git`, `prettyping`,
`sublime-text` and `zoxide`).
The fish and zsh configs define no aliases of their own. mise activates in
non-interactive shells too, so these apply to tool-invoked commands — not just
the interactive prompt.

Aliases that shadow standard commands (use `command <name>` to bypass):

| Alias                  | Actually runs   | Gotcha                                  |
| ---------------------- | --------------- | --------------------------------------- |
| `ls`, `ld`, `lf`, `lg` | `eza`           | icons and columns differ from coreutils |
| `td`, `tg`, `tl`       | `eza --tree`    | recursive by default                    |
| `cd`                   | `z` (zoxide)    | jumps by frecency, not by literal path  |
| `diff`                 | `diff-so-fancy` | no unified diff, no diff exit codes     |
| `ping`                 | `prettyping`    |                                         |
| `dig`                  | `doggo`         |                                         |
| `edit`                 | `subl`          | opens a GUI editor                      |
| `plist`                | `plutil -p`     |                                         |

The rest shadow nothing and are safe: git (`gc`, `gp`, `gpull`, `gs`), docker
(`dclean`, `dlist`, `dprune`, `dpull`, `drun` — `dpull`/`drun` force
`linux/amd64`), k8s (`k`, `kgc`, `kubectx`, `kc-local`), claude (`cc`, `cco`,
`ccf`), and mise task wrappers (`ips`, `ipv4`, `myps`, `updateall`, `code`, …).

## Think Before Coding

Don't assume. Surface tradeoffs. **Ask when uncertain.**

- State assumptions explicitly before implementing
- If multiple interpretations exist, present them — don't pick silently
- If a simpler approach exists, say so and push back
- If something is unclear, stop and name what's confusing

## Simplicity First

Minimum code that solves the problem. Nothing speculative.

- No features, abstractions, or configurability beyond what was asked
- No error handling for impossible scenarios
- If 200 lines could be 50, rewrite it

## Surgical Changes

Touch only what you must. Clean up only your own mess.

- Don't improve adjacent code, comments, or formatting
- Match existing style even if you'd do it differently
- If you notice unrelated dead code, mention it — don't delete it
- Remove only imports/variables/functions that **your** changes made unused

Every changed line should trace directly to the request.

## Goal-Driven Execution

Transform tasks into verifiable goals before starting:

- "Fix the bug" → write a test that reproduces it, then make it pass
- "Refactor X" → ensure tests pass before and after

For multi-step tasks, state a brief plan:

```text
1. [step] → verify: [check]
2. [step] → verify: [check]
```

Clarifying questions come **before** implementation, not after mistakes.
