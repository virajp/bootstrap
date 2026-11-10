# The `firebase` tool

The Firebase CLI only. The repository owns its Firebase config files.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | `node`  |
| Parent       | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- A script, a task or a CI job of the repository calls `firebase`.

`bootstrap detect` runs the detect script of the tool (see "Scripts"). It
finds `firebase.json`. A config file is not enough: the repository can use
the Firebase emulators in a container that installs its own CLI. Ask the user
before you add the tool.

## What the tool ships

| Path in the repository                         | Kind      | Use                           |
| ---------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/firebase/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool            | Version  | Environments |
| -------------------- | -------- | ------------ |
| `npm:firebase-tools` | `latest` | dev, ci      |

## Tasks

None.

## Scripts

| Script           | Kind   | Use                                                    |
| ---------------- | ------ | ------------------------------------------------------ |
| `scripts/detect` | detect | `match` with `firebase.json` at the root, else `none`. |

The sync runs the scripts from the clone of `bootstrap`, and never copies them.
See [`docs/design/tools.md`](../../design/tools.md).

## Contributions

None.

## Values

Only `env`. No questions. See
[`docs/design/bootstrap-config.md`](../../design/bootstrap-config.md).

## Files the repository supplies

| File                                    | Used by    |
| --------------------------------------- | ---------- |
| `firebase.json`, `.firebaserc`, others. | `firebase` |

The repository owns these files and their location. `bootstrap` does not ship,
seed or read them.

## Adapt to the repository

None.

## Migration

- Remove `npm:firebase-tools` and any other Firebase CLI entry from the
  repository-owned mise files.
