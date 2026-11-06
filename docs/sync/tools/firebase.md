# The `firebase` tool

The Firebase CLI only. The tool ships no Firebase config file.

| Property     | Value   |
| ------------ | ------- |
| Root only    | No      |
| Depends      | None    |
| Default envs | dev, ci |
| Seeds        | None    |

## When the tool is relevant

- A script, a task or a CI job of the repository calls `firebase`.

A config file such as `firebase.json` is not enough. The repository can use the
Firebase emulators in a container that installs its own CLI.

## What the tool ships

| Path in the repository                         | Kind      | Use                           |
| ---------------------------------------------- | --------- | ----------------------------- |
| `.config/mise/conf.d/firebase/mise.<env>.toml` | Generated | The mise config of this tool. |

## mise tools

| mise tool            | Version  | Environments | Note        |
| -------------------- | -------- | ------------ | ----------- |
| `npm:firebase-tools` | `latest` | dev, ci      | Needs node. |

## Tasks

None.

## Pre-commit hooks

None.

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
