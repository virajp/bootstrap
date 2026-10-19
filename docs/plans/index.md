# Plans

The product's plans as a set — the one file every vwf command reads to find
a plan without walking the member repos, and the queue `/vwf:execute next`
and `/vwf:execute all` read to pick the next runnable plan.

## Plans

| Folder                                           | Kind   | Plan                                                               | Target repo | Priority | Status   | Gaps | Requires                            | Backlog          |
| ------------------------------------------------ | ------ | ------------------------------------------------------------------ | ----------- | -------- | -------- | ---- | ----------------------------------- | ---------------- |
| `docs/plans/2026-10-19-android-swift-java-tools` | change | Add the java, android and swift tools and docs/sync/vwf.md         | —           | 10       | APPROVED | -    | —                                   | B11, B03 (piece) |
| `docs/plans/2026-10-19-mobile-toolchain-tasks`   | change | Ship the Gradle, SwiftPM and Android SDK tasks of the mobile tools | —           | 20       | APPROVED | -    | 2026-10-19-android-swift-java-tools | B03              |
