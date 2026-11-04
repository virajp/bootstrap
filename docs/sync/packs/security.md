# The `security` pack

The scanners: gitleaks for secrets, grype and osv-scanner for the
dependencies, and the `code:sec` task that runs them.

| Property | Value                              |
| -------- | ---------------------------------- |
| Members  | `gitleaks`, `grype`, `osv-scanner` |

## When the pack is relevant

Always. The `base` pack includes it as `"pack:security"`. Do not write
`[packs.security]` in the values file.

## What the pack adds

Each member is `default = true` in `repo/packs.toml`, for example
`gitleaks = { default = true }`.

| Member        | Default | Tool file                                         |
| ------------- | ------- | ------------------------------------------------- |
| `gitleaks`    | Yes     | [`tools/gitleaks.md`](../tools/gitleaks.md)       |
| `grype`       | Yes     | [`tools/grype.md`](../tools/grype.md)             |
| `osv-scanner` | Yes     | [`tools/osv-scanner.md`](../tools/osv-scanner.md) |

[`_security_base`](../tools/_security_base.md) is the common tool of the pack.
It is not a member. Each scanner has it in its `depends`, so each scanner
brings it, also a single scanner without the pack.

## Environments

Each tool of the pack gets its own `default_envs` (dev and ci), limited by
`environments` of the values file. `_security_base` gets the environments of
the scanners that need it.

## Questions

None. The `base` pack always adds this pack. A consumer cannot remove a tool
of the pack.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
