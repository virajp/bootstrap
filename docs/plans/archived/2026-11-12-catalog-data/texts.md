# The catalog data of plan 1b

The user accepted these texts and this mapping on 2026-11-12 (D1, D2). The
units copy them as they are. Each summary has 47 characters or fewer.

## The categories files

Each file starts with the comment that its block shows (D10 b), then the tables
in this order.

`repo/categories.toml`:

```toml
# The categories of the repo tools. Each [category.<id>] table has one key,
# title. The file order is the order of the TUI. Each tools.toml of repo/tools/
# names one id in its key category. The sync never copies this file. See
# docs/design/tools.md.
[category.core]
title = "Core"

[category.format-lint]
title = "Format and lint"

[category.editor-ai]
title = "Editor and AI"

[category.languages]
title = "Languages"

[category.package-managers]
title = "Package managers"

[category.security]
title = "Security"

[category.secrets]
title = "Secrets"

[category.containers]
title = "Containers"

[category.cloud]
title = "Cloud and deploy"

[category.forge]
title = "Forge"
```

`macos/categories.toml`:

```toml
# The categories of the Mac tools. Each [category.<id>] table has one key,
# title. The file order is the order of the TUI. Each tools.toml of macos/tools/
# names one id in its key category. The sync never copies this file. See
# docs/design/tools.md.
[category.system]
title = "System"

[category.shell]
title = "Shell and terminal"

[category.cli]
title = "CLI tools"

[category.dev]
title = "Development"

[category.editors]
title = "Editors"

[category.ai]
title = "AI"

[category.browsers]
title = "Browsers"

[category.passwords]
title = "Passwords and secrets"

[category.productivity]
title = "Productivity"

[category.office]
title = "Office"

[category.notes]
title = "Notes and reading"

[category.media]
title = "Media"

[category.chat]
title = "Chat"

[category.disk]
title = "Disk"
```

## The comments of the packs files

Not used (D13).

## Repo tools

| Tool             | category           | summary                                       | description                                                                                                                                                                       |
| ---------------- | ------------------ | --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `_security_base` | `security`         | The code:sec task and the sec hook            | The code:sec task, which runs each scanner of the security pack, and the sec hook of pre-commit. gitleaks, grype and osv-scanner need it, and you cannot select it alone.         |
| `android`        | `languages`        | The Android SDK command-line tools            | The Android SDK command-line tools. It needs the java tool.                                                                                                                       |
| `apko`           | `containers`       | Builds OCI images from apk packages           | apko builds OCI container images from apk packages, from a declarative config.                                                                                                    |
| `aube`           | `package-managers` | The aube package manager for Node             | aube, the package manager of a repository that selects it. Select this tool or the pnpm tool, not both. It needs the node tool.                                                   |
| `bun`            | `languages`        | The Bun runtime and package manager           | The Bun runtime and package manager. The repository owns its bunfig.toml.                                                                                                         |
| `claude`         | `editor-ai`        | Claude Code plugins, with setup:ai            | The setup:ai task, which installs and updates the Claude Code plugins of the repository. In a git submodule, setup:ai does nothing.                                               |
| `cloudflare`     | `cloud`            | The Cloudflare CLI                            | The Cloudflare CLI (npm:cf). It needs the node tool.                                                                                                                              |
| `cosign`         | `containers`       | Signs and verifies container images           | cosign signs and verifies container images and other artifacts in an OCI registry.                                                                                                |
| `crane`          | `containers`       | Operates on container images and registries   | crane operates on container images and registries: it copies, examines and changes images with no Docker daemon.                                                                  |
| `docker`         | `containers`       | Local services with Docker Compose            | Local services in .tools/docker-compose.yaml. The setup:external:* tasks start and stop them.                                                                                     |
| `doppler`        | `secrets`          | Secrets from Doppler                          | The Doppler CLI and the setup:secrets:10-doppler task. Select it only when the repository keeps its secrets in Doppler.                                                           |
| `dprint`         | `format-lint`      | The formatter and the shared dprint.jsonc     | dprint, the formatter of the sync and of code:format, with the shared dprint.jsonc and the setup:dprint task.                                                                     |
| `firebase`       | `cloud`            | The Firebase CLI                              | The Firebase CLI only. The repository owns its Firebase config files. It needs the node tool.                                                                                     |
| `flutter`        | `languages`        | Flutter and Dart                              | Flutter and Dart. It needs the android tool for the Android SDK.                                                                                                                  |
| `fnox`           | `secrets`          | Secrets from .config/fnox.toml                | Secrets from .config/fnox.toml. The setup:secrets:20-fnox task examines them.                                                                                                     |
| `gcloud`         | `cloud`            | The Google Cloud CLI                          | The Google Cloud CLI, for a repository that deploys to Google Cloud. It needs the python tool.                                                                                    |
| `git`            | `core`             | The shared git config and the merge tasks     | The shared .gitignore and .gitattributes, the commit convention, the git-config check, the merge tasks and worktrunk. Each repository has it.                                     |
| `github`         | `forge`            | The gh CLI, actionlint and GitHub templates   | The gh CLI, actionlint for the GitHub Actions workflows, and the issue and pull request templates.                                                                                |
| `gitlab`         | `forge`            | The glab CLI and the GitLab templates         | The glab CLI, and the issue and merge request templates of GitLab.                                                                                                                |
| `gitleaks`       | `security`         | Scans the code for secrets                    | gitleaks scans the code and the git history for secrets. It adds the code:sec:10-gitleaks task.                                                                                   |
| `go`             | `languages`        | Go, with the gofmt and go vet gates           | Go, with the gofmt and go vet gates.                                                                                                                                              |
| `grype`          | `security`         | Scans the dependencies for vulnerabilities    | grype scans the dependencies for known vulnerabilities. It adds the code:sec:20-grype task.                                                                                       |
| `java`           | `languages`        | The Java development kit (JDK)                | The Java development kit (JDK), for Gradle and the Android SDK.                                                                                                                   |
| `kotlin`         | `languages`        | Kotlin, with the ktlint and detekt gates      | The Kotlin language server (dev only), and the ktlint and detekt gates. It needs the java tool.                                                                                   |
| `linter`         | `format-lint`      | The house linter and ShellCheck               | The house linter and ShellCheck, with the code:lint and setup:linter tasks. It needs the node tool.                                                                               |
| `melange`        | `containers`       | Builds apk packages from a pipeline           | melange builds apk packages from a declarative pipeline, for apko images.                                                                                                         |
| `mise`           | `core`             | The mise config, the tasks and the sync entry | The shared mise config, the task library, the setup and code orchestrators, and the entry of the sync: setup:configs and the gate code:lint:60-bootstrap. Each repository has it. |
| `node`           | `languages`        | The Node.js runtime                           | The Node.js runtime. A repository with a package.json also selects one package manager: pnpm or aube.                                                                             |
| `osv-scanner`    | `security`         | Scans the dependencies with the OSV database  | osv-scanner examines the dependencies against the OSV vulnerability database. It adds the code:sec:30-osv-scanner task.                                                           |
| `pnpm`           | `package-managers` | The pnpm package manager for Node             | pnpm, the package manager of a repository that selects it. Select this tool or the aube tool, not both. It needs the node tool.                                                   |
| `pre-commit`     | `core`             | The shared pre-commit config and its hooks    | The shared pre-commit config, its setup task and code:precommit. Each repository has it. It needs the python tool.                                                                |
| `pulumi`         | `cloud`            | Infrastructure as code with Pulumi            | Pulumi, for infrastructure as code. A repository that deploys to Google Cloud also selects gcloud.                                                                                |
| `python`         | `languages`        | Python and uv, with ruff and pyrefly          | Python and uv. In a repository with a pyproject.toml, also ruff and pyrefly, with their code:lint and code:format tasks.                                                          |
| `rust`           | `languages`        | Rust and the zola site generator              | Rust and the zola site generator.                                                                                                                                                 |
| `swift`          | `languages`        | Swift, SwiftLint and swift-format (macOS)     | Swift and SwiftPM from the host: Xcode for a root with an .xcodeproj, else the Command Line Tools. Also the SwiftLint and swift-format gates. macOS only.                         |
| `vscode`         | `editor-ai`        | The VS Code settings and the extensions       | The VS Code settings, and the shared extensions.json with the extensions of the selected tools. Root only: the root does not list the extensions of its git submodules.           |

## Mac tools

The `description` of each Mac tool stays as it is. A backslash in a cell is only
the
Markdown escape of the next character: copy the text without it (the summary
of `cli` is `The projects folder, the aliases and func:*`).

| Tool                   | category       | summary                                        |
| ---------------------- | -------------- | ---------------------------------------------- |
| `1password`            | `passwords`    | 1Password, the SSH agent and the commit signer |
| `1password-cli`        | `passwords`    | The 1Password CLI (op)                         |
| `1password-safari`     | `passwords`    | 1Password for Safari                           |
| `age`                  | `passwords`    | age, the file encryption tool                  |
| `alfred`               | `productivity` | Alfred, the application launcher               |
| `android-studio`       | `dev`          | Android Studio, for Android apps               |
| `bash`                 | `shell`        | bash, with its rc files                        |
| `bat`                  | `cli`          | bat, cat with syntax highlighting              |
| `battery-health`       | `system`       | Battery Health 2, the battery monitor          |
| `bitwarden`            | `passwords`    | Bitwarden, the password manager                |
| `brave-browser`        | `browsers`     | The Brave web browser                          |
| `btop`                 | `cli`          | btop, the resource monitor                     |
| `cardhop`              | `productivity` | Cardhop, the contacts app                      |
| `claude-cli`           | `ai`           | Claude Code, its settings and its tasks        |
| `claude-desktop`       | `ai`           | The Claude desktop app                         |
| `claude-status`        | `ai`           | The status line of Claude Code                 |
| `cli`                  | `cli`          | The projects folder, the aliases and func:\*   |
| `codeburn`             | `ai`           | codeburn, the cost report of the AI tools      |
| `coreutils`            | `cli`          | The GNU core utilities                         |
| `curl`                 | `cli`          | curl, which transfers data with URLs           |
| `daisydisk`            | `disk`         | DaisyDisk, the disk space analyzer             |
| `dasel`                | `cli`          | dasel, a query tool for JSON, YAML and TOML    |
| `developer`            | `dev`          | Developer, the Apple developer app             |
| `devtools`             | `dev`          | The env of development tools                   |
| `diff-so-fancy`        | `cli`          | diff-so-fancy, the git diff formatter          |
| `disk-doctor`          | `disk`         | Disk Doctor, the disk cleaner                  |
| `disk-map`             | `disk`         | Disk Map, the disk space map                   |
| `doggo`                | `cli`          | doggo, the DNS lookup tool                     |
| `dprint`               | `dev`          | dprint, the formatter, and its global config   |
| `duplicate-detective`  | `disk`         | Duplicate Detective, the duplicate finder      |
| `easyfind`             | `productivity` | EasyFind, the file search app                  |
| `eza`                  | `cli`          | eza, a modern ls, and its aliases              |
| `final-cut-pro`        | `media`        | Final Cut Pro, the video editor                |
| `fish`                 | `shell`        | fish, with its config and the login shell      |
| `fnox`                 | `passwords`    | fnox, secrets from the macOS Keychain          |
| `fonts`                | `system`       | The Nerd Fonts for the terminal                |
| `fzf`                  | `cli`          | fzf, the fuzzy finder                          |
| `gcloud`               | `dev`          | The Google Cloud CLI                           |
| `gh`                   | `dev`          | The GitHub CLI and its config                  |
| `ghostty`              | `shell`        | Ghostty, the terminal emulator                 |
| `git`                  | `dev`          | git, its config and the identity per forge     |
| `git-tools`            | `dev`          | git-tools, more git commands                   |
| `github-desktop`       | `dev`          | GitHub Desktop                                 |
| `glab`                 | `dev`          | glab, the GitLab CLI                           |
| `google-chrome`        | `browsers`     | The Google Chrome web browser                  |
| `gping`                | `cli`          | gping, ping with a graph                       |
| `handbrake`            | `media`        | HandBrake, the video transcoder                |
| `homebrew`             | `system`       | Homebrew, the Brewfile and its upgrade task    |
| `httping`              | `cli`          | httping, ping over HTTP                        |
| `imagemagick`          | `cli`          | ImageMagick, the image tool                    |
| `ipcalc`               | `cli`          | ipcalc, the IP address calculator              |
| `jq`                   | `cli`          | jq, the JSON processor                         |
| `keynote`              | `office`       | Keynote, the presentation app                  |
| `kindle`               | `notes`        | Kindle, the e-book reader                      |
| `lazygit`              | `dev`          | lazygit, a terminal UI for git                 |
| `lcov`                 | `cli`          | lcov, the code coverage tool                   |
| `macos-defaults`       | `system`       | The macOS settings: Finder, Dock and more      |
| `magnet`               | `productivity` | Magnet, the window manager                     |
| `mas`                  | `system`       | mas, the Mac App Store CLI                     |
| `microsoft-excel`      | `office`       | Microsoft Excel                                |
| `microsoft-powerpoint` | `office`       | Microsoft PowerPoint                           |
| `microsoft-word`       | `office`       | Microsoft Word                                 |
| `mise`                 | `system`       | mise, its config and the Mac sync tasks        |
| `network`              | `system`       | The ips:* tasks and the network aliases        |
| `node`                 | `dev`          | Node.js (LTS)                                  |
| `notion`               | `notes`        | Notion, the notes and planning app             |
| `numbers`              | `office`       | Numbers, the spreadsheet app                   |
| `obsidian`             | `notes`        | Obsidian, notes in Markdown files              |
| `opencode`             | `ai`           | OpenCode, the AI coding agent                  |
| `openwhispr`           | `ai`           | OpenWhispr, voice-to-text dictation            |
| `orbstack`             | `dev`          | OrbStack, for Docker and Kubernetes            |
| `orca`                 | `ai`           | Orca, which runs AI coding agents              |
| `pages`                | `office`       | Pages, the word processor                      |
| `pdf-protector`        | `notes`        | PDF Protector, which locks PDF files           |
| `pdf-squeezer`         | `notes`        | PDF Squeezer, which makes PDF files smaller    |
| `prettyping`           | `cli`          | prettyping, a better ping                      |
| `pstree`               | `cli`          | pstree, the process tree                       |
| `python`               | `dev`          | Python, through mise                           |
| `ripgrep`              | `cli`          | ripgrep, the fast search tool                  |
| `ruby`                 | `dev`          | Ruby and the RubyGems config                   |
| `setapp`               | `system`       | Setapp, apps by subscription                   |
| `sill`                 | `productivity` | Sill, the menu bar app                         |
| `ssh`                  | `system`       | The ssh client config                          |
| `starship`             | `shell`        | starship, the shell prompt                     |
| `sublime-text`         | `editors`      | Sublime Text, the text editor                  |
| `superfile`            | `cli`          | superfile, the terminal file manager           |
| `tailscale`            | `system`       | Tailscale, the private network                 |
| `taskexplorer`         | `productivity` | TaskExplorer, which shows the processes        |
| `telegram`             | `chat`         | Telegram, the chat app                         |
| `testflight`           | `dev`          | TestFlight, for beta apps                      |
| `the-unarchiver`       | `productivity` | The Unarchiver, which opens archives           |
| `touchid-sudo`         | `system`       | Touch ID for sudo                              |
| `transporter`          | `dev`          | Transporter, which uploads apps to Apple       |
| `tree`                 | `cli`          | tree, the directory tree                       |
| `uv`                   | `dev`          | uv, the Python package manager                 |
| `visual-studio-code`   | `editors`      | Visual Studio Code, the code editor            |
| `vlc`                  | `media`        | VLC, the media player                          |
| `vorssaint`            | `productivity` | Vorssaint, the menu bar toolkit                |
| `warp`                 | `shell`        | Warp, the terminal                             |
| `watch`                | `cli`          | watch, which runs a command again and again    |
| `wget`                 | `cli`          | wget, the network downloader                   |
| `whatsapp`             | `chat`         | WhatsApp, the chat app                         |
| `worktrunk`            | `dev`          | worktrunk, the manager of git worktrees        |
| `xcode`                | `dev`          | Xcode, for Apple apps                          |
| `yq`                   | `cli`          | yq, the YAML processor                         |
| `zed`                  | `editors`      | Zed, the code editor                           |
| `zed-delta`            | `ai`           | Zed Delta, coding with AI agents               |
| `zoxide`               | `cli`          | zoxide, the fast directory jumper              |
| `zsh`                  | `shell`        | zsh, with zinit and its rc files               |

## The pack descriptions

Not used (D13). Plan 2 writes the descriptions of the new packs.
