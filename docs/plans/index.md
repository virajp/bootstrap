# Plans

The product's plans as a set — the one file every vwf command reads to find
a plan without walking the member repos, and the queue `/vwf:execute next`
and `/vwf:execute all` read to pick the next runnable plan.

## Plans

| Folder                                   | Kind   | Plan                                         | Target repo | Priority | Status   | Gaps | Requires                                                               | Backlog |
| ---------------------------------------- | ------ | -------------------------------------------- | ----------- | -------- | -------- | ---- | ---------------------------------------------------------------------- | ------- |
| `docs/plans/2026-10-20-mobile-gap-fixes` | change | Close the open gaps of the mobile tool plans | —           | 30       | COMPLETE | 9    | 2026-10-19-android-swift-java-tools, 2026-10-19-mobile-toolchain-tasks | —       |
