---
name: npm-package-manifest
version: 0.1.0
category: development
description: >-
  Writing the publish fields of the CLI's package manifest. The files
  allowlist decides what ships, there is one bin entry with a node shebang,
  and publishConfig carries public access. Treat tarball hygiene as a
  correctness concern, because a published file can never be recalled.
  Auto-applies when editing the CLI's package.json or an npmignore file.
license: MIT
user-invocable: false
allowed-tools: Read Grep Glob Edit Write Bash
paths:
  - "cli/package.json"
  - "cli/**/.npmignore"
---

# The package manifest's publish fields

There are two subjects here, and the second one is the one that is easy to
treat as optional.

| Doing | Read |
| --- | --- |
| Shaping what the tarball is: `bin`, `files`, `engines`, dependencies | [The artifact](references/artifact.md) |
| Checking what must never be inside it | [Hygiene](references/hygiene.md) |

**One rule applies before any reference:** `files` is an allowlist, and
nothing reaches the registry unless it is on it. Read the pack listing
before every release, because the only time to catch a stray file is before
it is published.

Scope: this skill covers the **publish fields only**. Dependencies, scripts
and the TypeScript build belong to the language and workspace bundles.
Choosing this target, releasing it, and runtime configuration belong to
the sibling `npm-registry` skill.
