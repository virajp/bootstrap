# Plan — Mac tools: app-only (52) (2026-11-04)

## Goal

After this plan lands, `macos/tools/` holds all 92 Mac tools of
`MS/tools/` in the new format. The 52 tools of this plan have one file each,
`tools.toml`, with `description`, `depends` and `[mise.*]` tables.
`macos/packs.toml` has all 14 packs of `MS/tools/packs.toml` with the same
members, and `default = true` exactly where the old `default` lists name a
member (or on every member when a pack has no `default` list).
`macos/examples/macos.toml` selects every tool outside pack `base`, so
`test:macos-tools` renders every Mac tool.

Plan 14 of 16 — requires `2026-11-04-macos-tools-extras`; required by
`2026-11-04-docs-design-split`. No reversal of a standing decision.

## Facts the survey established

The design rulings come from the planning session of 2026-11-04 (artifact
https://claude.ai/artifact/7LKisXPLbSGtrFfVD36ez7). Each unit quotes the
rulings it implements; no unit reads another file for them.

`MS` is `/Users/virajpatel/Projects/github.com/virajp/macos-setup` (read
only; read it by its absolute path). No unit edits `MS`.

- **The 52 tools.** None has a file other than `tools.toml` and one mise
  fragment `MS/tools/<tool>/.config/mise/conf.d/<tool>.toml`. Old
  `tools.toml` keys: `description` on each; `depends = ["mas"]` on each Mac
  App Store tool (battery-health, cardhop, daisydisk, developer,
  disk-doctor, disk-map, duplicate-detective, final-cut-pro, keynote,
  kindle, magnet, microsoft-excel, microsoft-powerpoint, microsoft-word,
  numbers, pages, pdf-protector, pdf-squeezer, testflight, the-unarchiver,
  transporter, xcode); no other key.
- **Fragments.** 46 fragments hold only `[bootstrap.packages]` with one
  entry. Six hold more (D3):
  - `gcloud`: `[tools.gcloud] version = "latest"` and `[env]`
    (`CLOUDSDK_ACTIVE_CONFIG_NAME = "default"`,
    `USE_GKE_GCLOUD_AUTH_PLUGIN = { tools = true, value = "true" }`); no
    `[bootstrap.packages]`.
  - `opencode`: `[shell_alias] oc = "opencode --auto"`.
  - `sublime-text`: `[shell_alias] edit = "subl"`.
  - `orca`: `[bootstrap.brew.taps] "stablyai/orca"`.
  - `sill`: `[bootstrap.brew.taps] "saffsanity/sill"`.
  - `vorssaint`: `[bootstrap.macos.defaults."com.vorssaint.utils"]`, about
    185 keys with comment lines (195 lines in all).
- **Old packs** (`MS/tools/packs.toml`, 14 packs in this order): `base`,
  `ai`, `essentials`, `security`, `dev`, `terminal`, `apple-dev`, `iwork`,
  `ms-office`, `disk`, `media`, `chat`, `notes`, `cli-extras`. `iwork`,
  `ms-office`, `chat` and `cli-extras` have no `default` list (absent:
  every member is a default).
- **After P13** `macos/packs.toml` has `base`, `essentials` (with
  `tailscale`, `diff-so-fancy`, `git-tools`, `glab`), `ai`, `security`,
  `dev`, `terminal`, `chat` and `cli-extras`, with the P11 to P13 members
  only. `macos/examples/macos.toml` selects the 20 tools of P13.
- **Sample selection:** `MS/tests/setup-sync/sample-config/macos-setup.toml`
  selects every tool outside pack `base` (73 tools: the 20 of P13, the 52
  of this plan, and `diff-so-fancy`).
- **Earlier plans of the chain.** The `tools.toml` files of P11 to P13 are
  the model of the key set and the order. `test:macos-tools` renders every
  tool that `macos/examples/macos.toml` selects. The plan gate
  `mise run test` runs `test:repo-sync`, `test:setup-sync` and
  `test:macos-tools`.
- **Commit convention:** types `feat`, `fix`, `perf`, `refactor`, `revert`,
  `test`, `ops`, `docs`; no scopes. No CI.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file — the file a package
manager or a toolchain manager writes to pin what it resolved — with Write,
Edit, `sed` or a script: a lock file changes only as the output of the CLI
that owns it, the command the stack's package-manager conventions name, and
the unit keeps what that CLI wrote. A unit whose manifest change moves a lock
file lists that lock file in its `owns` beside the manifest.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan the unit could proceed past on a stated
assumption; it is recorded and the run continues. An `UNRESOLVED:` is a ruling
the unit could not proceed without; it blocks the unit and its dependents.

## Out of scope

- `docs/setup/tools.md`, the Mac tool reference: P16
  (`2026-11-04-docs-setup`).
- The scratch-`HOME` comparison with `MS/dotfiles/` (phase 4): a hand step
  after the chain, with the user's approval of each difference.
- Any change to `MS`: the `macos-setup` repo is read only here.
- `mise run setup:configs` in this repository and in consumers: a hand step.

## Parked

See P16 Parked (`docs/plans/2026-11-04-docs-setup/index.md`).

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-04-macos-tools-apps

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
