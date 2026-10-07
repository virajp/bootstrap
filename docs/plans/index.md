# Plans

The product's plans as a set — the one file every vwf command reads to find a
plan without walking the member repos, and the queue `/vwf:execute next` reads
to pick the next runnable plan.

## Plans

| Folder                                         | Kind  | Plan                                 | Target repo | Priority | Status   | Requires                          | Backlog |
| ---------------------------------------------- | ----- | ------------------------------------ | ----------- | -------- | -------- | --------------------------------- | ------- |
| `docs/plans/2026-10-06-2332-tool-setup-config` | cycle | Tool catalog and setup config        | —           | 10       | COMPLETE | —                                 | —       |
| `docs/plans/2026-10-06-2333-setup-repository`  | cycle | Set up a repository (bootstrap init) | —           | 20       | APPROVED | 2026-10-06-2332-tool-setup-config | —       |
