# Plans

The product's plans as a set — the one file every vwf command reads to find a
plan without walking the member repos, and the queue `/vwf:execute next` reads
to pick the next runnable plan.

## Plans

| Folder                                         | Kind  | Plan                                           | Target repo | Priority | Status   | Requires                         | Backlog |
| ---------------------------------------------- | ----- | ---------------------------------------------- | ----------- | -------- | -------- | -------------------------------- | ------- |
| `docs/plans/2026-10-06-2332-tool-setup-config` | cycle | Tool catalog and setup config                  | —           | 10       | COMPLETE | —                                | —       |
| `docs/plans/2026-10-08-1124-tool-catalog`      | cycle | Tool catalog, tool categories and setup config | —           | 10       | RUNNING  | —                                | —       |
| `docs/plans/2026-10-09-0931-setup-repository`  | cycle | Set up a repository (bootstrap init)           | —           | 20       | APPROVED | 2026-10-08-1124-tool-catalog     | —       |
| `docs/plans/2026-10-09-1027-remove-tool`       | cycle | Remove a tool (bootstrap remove)               | —           | 30       | APPROVED | 2026-10-09-0931-setup-repository | —       |
| `docs/plans/2026-10-09-1146-add-tool`          | cycle | Add a tool (bootstrap add)                     | —           | 40       | APPROVED | 2026-10-09-1027-remove-tool      | —       |
