# Plans

The product's plans as a set — the one file every vwf command reads to find
a plan without walking the member repos, and the queue `/vwf:execute next`
and `/vwf:execute all` read to pick the next runnable plan.

## Plans

| Folder                                       | Kind   | Plan                                         | Target repo | Priority | Status   | Gaps                                    | Requires                                                               | Backlog |
| -------------------------------------------- | ------ | -------------------------------------------- | ----------- | -------- | -------- | --------------------------------------- | ---------------------------------------------------------------------- | ------- |
| `docs/plans/2026-10-20-mobile-gap-fixes`     | change | Close the open gaps of the mobile tool plans | —           | 30       | COMPLETE | 9                                       | 2026-10-19-android-swift-java-tools, 2026-10-19-mobile-toolchain-tasks | —       |
| `docs/plans/2026-10-21-packs-and-base-split` | change | Build the packs and split the base tool      | —           | 10       | COMPLETE | 0 — moved to 2026-10-21-packs-gap-fixes | —                                                                      | —       |
| `docs/plans/2026-10-21-packs-gap-fixes`      | change | Close the gaps of the packs and base split   | —           | 10       | RUNNING  | -                                       | —                                                                      | —       |
