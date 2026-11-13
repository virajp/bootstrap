# The texts of plan 2a

The user accepted these texts on 2026-11-13. U2 copies them as they are.

## The pack descriptions

| Pack         | description                                                                                                                                                             |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| repo `base`  | The tools of each repository: mise, git, pre-commit, dprint, the linter and the security scanners, which a repository cannot remove, and VS Code.                       |
| Mac `base`   | The tools of each Mac: mise, Homebrew, the shells, git and its tools, ssh, the fonts, Python and the command-line tools, which a Mac cannot remove, and VS Code and uv. |
| `claude`     | Claude: the desktop app, Claude Code, its status line, and codeburn, the cost report of the AI tools.                                                                   |
| `essentials` | The daily apps: Alfred, EasyFind, The Unarchiver, Brave and Chrome, GitHub Desktop, Tailscale, TaskExplorer, Battery Health and Setapp.                                 |
| `1password`  | 1Password: the app, its CLI and its Safari extension.                                                                                                                   |
| `dev`        | More development tools: lcov, OrbStack, superfile and Sublime Text.                                                                                                     |
| `terminal`   | The terminal apps Ghostty and Warp.                                                                                                                                     |
| `mobile-dev` | The tools for Apple and Android apps: Xcode, Developer, TestFlight, Transporter and Android Studio.                                                                     |
| `iwork`      | The iWork apps: Keynote, Numbers and Pages.                                                                                                                             |
| `ms-office`  | The Microsoft Office apps: Excel, PowerPoint and Word.                                                                                                                  |
| `disk`       | The disk tools: DaisyDisk, Disk Doctor, Disk Map and Duplicate Detective.                                                                                               |
| `media`      | The media apps VLC and HandBrake.                                                                                                                                       |

## The comment at the top of `repo/packs.toml`

```toml
# The packs of the repository. A pack has no files of its own; it is a shortcut
# of `bootstrap add pack`, and the values file keeps no record of it. Each key
# of its table is a member, `<tool> = {}`, and each member is a default; a
# nested pack is `"pack:<name>" = {}`. In `base` only, `required = true` marks
# a member that a repository cannot remove. The key `description` is the text
# of the pack. See docs/design/packs.md.
```

## The comment at the top of `macos/packs.toml`

```toml
# The packs of the Mac tools in macos/tools/. A pack has no files of its own; it is a shortcut
# of `bootstrap mac add pack`, and the values file keeps no record of it. Each key
# of its table is a member, `<tool> = {}`, and each member is a default; a
# nested pack is `"pack:<name>" = {}`. In `base` only, `required = true` marks
# a member that a Mac cannot remove. The key `description` is the text
# of the pack. See docs/design/packs.md.
```
