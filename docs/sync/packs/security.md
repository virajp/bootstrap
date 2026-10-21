# The `security` pack

The scanners: gitleaks for secrets, grype and osv-scanner for the
dependencies, and the `code:sec` task that runs them.

| Property | Value                              |
| -------- | ---------------------------------- |
| Depends  | `gitleaks`, `grype`, `osv-scanner` |
| Required | `_security_base`                   |

## When the pack is relevant

Always. The `base` pack includes it as `p:security`. Do not write
`[packs.security]` in the values file.

## What the pack adds

| Tool             | Tool file                                               |
| ---------------- | ------------------------------------------------------- |
| `gitleaks`       | [`tools/gitleaks.md`](../tools/gitleaks.md)             |
| `grype`          | [`tools/grype.md`](../tools/grype.md)                   |
| `osv-scanner`    | [`tools/osv-scanner.md`](../tools/osv-scanner.md)       |
| `_security_base` | [`tools/_security_base.md`](../tools/_security_base.md) |

`_security_base` is the common tool of the pack. It is in `required`, not in
`depends`: the sync always adds it with the pack.

## Environments

Each tool of the pack gets its own `default_envs` (dev and ci), limited by
`environments` of the values file.

## Questions

None. The `base` pack always adds this pack. A consumer cannot remove a tool
of the pack.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
