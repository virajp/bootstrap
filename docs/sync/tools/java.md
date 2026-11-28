# The `java` tool

The Java development kit (JDK), for Gradle and the Android SDK.

| Property     | Value                                                                  |
| ------------ | ---------------------------------------------------------------------- |
| Root only    | No                                                                     |
| Depends      | None                                                                   |
| Parent       | None                                                                   |
| Default envs | dev, ci                                                                |
| Seeds        | `.claude/skills/gradle/references/local.md`, only with `claude` skills |

## When the tool is relevant

- The repository builds with Gradle or Maven.
- Another selected tool adds it: `android` has `java` and `kotlin` in its
  `depends`, and `kotlin` has `java`. `java` stays a direct dependency of
  `android`, because the mise `depends` of a tool come from its direct
  `depends` only, and `android-sdk` needs the JDK of `java`.

`bootstrap detect` runs the detect script of the tool. It gives `match` when
the root, or a folder one level below it, has a `*.gradle` or `*.gradle.kts`
file (see "Scripts").

## What the tool ships

| Path in the repository                            | Kind      | Use                                                                                |
| ------------------------------------------------- | --------- | ---------------------------------------------------------------------------------- |
| `.config/mise/conf.d/java/mise.<env>.toml`        | Generated | The mise config of this tool.                                                      |
| `.claude/skills/gradle/SKILL.md`                  | Template  | The Claude Code skill `gradle`. Only with `claude` skills.                         |
| `.claude/skills/gradle/references/conventions.md` | Synced    | The conventions of the skill `gradle`. Only with `claude` skills.                  |
| `.claude/skills/gradle/references/local.md`       | Seed      | The own rules of the repository for the skill `gradle`. Only with `claude` skills. |

The skill files have `when = "claude.skills"`: the sync writes them only
when the repository selects the `claude` tool and its value `skills` is
`true` (the default). `SKILL.md` names `references/conventions.md` and
`references/local.md`. See [`claude-tool.md`](claude-tool.md).

## mise tools

| mise tool    | Version               | Environments | Note                                                                                                                                                             |
| ------------ | --------------------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `java`       | `lts`                 | dev, ci      | mise resolves `lts` to the newest LTS release. The mise setting `java.shorthand_vendor = "temurin"` selects the vendor, so each machine installs the same build. |
| `http:jdtls` | `1.63.0-202610191322` | dev          | The language server, from the Eclipse jdtls milestone downloads. In `[mise_env.dev.tools]`.                                                                      |

The tool declares the language server `jdtls` (`.java`), from the `http:jdtls`
pin. The pin is in `[mise_env.dev.tools]`, so it stays in dev and does not stop
`java` in ci. No version list gives the build timestamp of a jdtls milestone,
so the pin names one milestone: bump its version by hand. Its launcher
`bin/jdtls` is a Python script that runs on the host `python3`. The language
server is in [`docs/design/tools.md`](../../design/tools.md) ("The `[lsp]`
table"). `bootstrap lsp` lists it, and `code:lsp:check` of the `mise` tool tests
that it resolves.

## Tasks

None.

## Scripts

The sync runs the scripts of the tool from the clone, and never copies them.

| Script           | Kind   | Use                                                                        |
| ---------------- | ------ | -------------------------------------------------------------------------- |
| `scripts/detect` | detect | `match` with a Gradle file at the root or one level below it, else `none`. |

## Contributions

None.

## Values

Only `env`. No questions. The keys of the values file are in
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

None.

## Adapt to the repository

1. A tool that adds `java` gives it its own environments (see
   [`docs/design/mise.md`](../../design/mise.md)). `kotlin` alone is in dev
   only, so `java` of `kotlin` is in dev only. To have it in CI, add
   `[tools.java]` to the values file. `android` adds `java` and `kotlin` in the
   environments of `android` (dev and ci), so in an `android` repository
   `bootstrap sync` gives `kotlin` and `java` dev and ci (`kotlin-lsp` stays in
   dev only). `flutter` adds `android`, and `android` adds `java` and
   `kotlin`, in the environments of `flutter`.
2. If the repository pins a different JDK, tell the user. The values file
   can set it with `[overrides.mise.tools.java]`, for example
   `version = "21"`, and a different vendor with `[overrides.mise.settings]`,
   for example `java.shorthand_vendor = "openjdk"`.

## Migration

- Remove `java` and jdtls from the repository-owned mise files.
- Delete the old stackgen copy of `.claude/skills/gradle/` before the next sync
  (M62 of `migrate.md`).
