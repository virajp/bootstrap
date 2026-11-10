# The `container-build` pack

The tools that build, sign and move container images: apko, melange, cosign
and crane.

| Property | Value                                |
| -------- | ------------------------------------ |
| Members  | `apko`, `melange`, `cosign`, `crane` |

## When the pack is relevant

- The repository builds OCI images with apko, or apk packages with melange.
- The repository signs images with cosign, or copies them with crane.

A repository that builds images only with a `Dockerfile` does not need this
pack.

## What the pack adds

Each member is `default = true` in `repo/packs.toml`, for example
`apko = { default = true }`.

| Member    | Default | Tool file                                 |
| --------- | ------- | ----------------------------------------- |
| `apko`    | Yes     | [`tools/apko.md`](../tools/apko.md)       |
| `melange` | Yes     | [`tools/melange.md`](../tools/melange.md) |
| `cosign`  | Yes     | [`tools/cosign.md`](../tools/cosign.md)   |
| `crane`   | Yes     | [`tools/crane.md`](../tools/crane.md)     |

To add the tools, write `[packs.container-build]` in the values file.

## Environments

Each tool of the pack gets its own `default_envs` (dev and ci), limited by
`environments` of the values file. An `env` key of `[packs.container-build]`
replaces them.

## Questions

Does the repository need all four tools? If it needs only some of them, do not
select the pack. Select each tool with `bootstrap add <tool>`, or name it in
`bootstrap init <tool…>`.

## Adapt to the repository

Do the items in "Adapt to the repository" and "Migration" of each tool file
of the pack.
