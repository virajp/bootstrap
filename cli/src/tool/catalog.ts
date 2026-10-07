/** The Tool catalog (1.0) of docs/blueprint/entities/tool/index.md: read-only data, no I/O. */

export type ToolKind = "core" | "non-core";

export interface Tool {
  readonly name: string;
  readonly kind: ToolKind;
  readonly purpose: string;
  /** Target-repository paths, relative to the repository root. */
  readonly files: ReadonlyArray<string>;
}

/** The blueprint's `tasks/` prefix. */
const task = (path: string) => `.config/mise/tasks/${path}`;

export const tools: ReadonlyArray<Tool> = [
  {
    name: "mise",
    kind: "core",
    purpose: "Task runner and tool versions",
    files: [
      ".config/mise.toml",
      ".config/miserc.toml",
      ".config/mise/conf.d/_base/mise.toml",
      ".config/mise/conf.d/_base/mise.dev.toml",
      ".config/mise/conf.d/_base/mise.ci.toml",
      task("_scripts/helpers"),
      task("_scripts/checks"),
      task("_scripts/merge"),
      task("_scripts/placeholder"),
      task("code/all"),
      task("code/count"),
      task("code/worktrees"),
      task("code/merge/develop"),
      task("code/merge/main"),
      task("setup/all"),
      task("setup/mise"),
      task("setup/worktree"),
      task("setup/cleanup"),
      task("setup/deps/all"),
      task("setup/deps/install"),
      task("setup/deps/outdated"),
      task("setup/deps/audit"),
      task("setup/deps/upgrade"),
      task("setup/deps/cleanup"),
    ],
  },
  {
    name: "git",
    kind: "core",
    purpose: "Ignore rules and git config",
    files: [".gitignore", task("code/git-config")],
  },
  {
    name: "pre-commit",
    kind: "core",
    purpose: "Commit hooks and conventional-commit rules",
    files: [
      ".config/pre-commit-config.yaml",
      ".config/git-conventional-commits.yaml",
      task("code/precommit"),
      task("setup/precommit"),
    ],
  },
  {
    name: "vscode",
    kind: "core",
    purpose: "Editor defaults",
    files: [".vscode/settings.json", ".vscode/extensions.json"],
  },
  {
    name: "dprint",
    kind: "core",
    purpose: "Formatter",
    files: ["dprint.json", ".config/dprint.json", task("code/format")],
  },
  { name: "taplo", kind: "core", purpose: "TOML formatter", files: [".config/taplo.toml"] },
  {
    name: "gitleaks",
    kind: "core",
    purpose: "Secret scanning",
    files: [".config/gitleaks.toml", task("code/sec")],
  },
  { name: "grype", kind: "core", purpose: "Vulnerability scanning", files: [".config/grype.yaml"] },
  {
    name: "virajp-linter",
    kind: "core",
    purpose: "Linting",
    files: [".config/linter.yaml", "eslint.config.mjs", task("code/lint")],
  },
  {
    name: "claude",
    kind: "core",
    purpose: "AI coding-assistant settings and status line",
    files: [".config/mise/conf.d/ai/mise.dev.toml", ".config/claude-status.json", task("setup/ai")],
  },
  {
    name: "graphify",
    kind: "core",
    purpose: "Code knowledge graph",
    files: [task("code/graph"), ".graphifyignore"],
  },
  { name: "mempalace", kind: "core", purpose: "Memory store config", files: ["mempalace.yaml"] },
  {
    name: "fnox",
    kind: "core",
    purpose: "Secrets provider",
    files: [".config/fnox.toml", task("setup/secrets")],
  },
  {
    name: "github",
    kind: "non-core",
    purpose: "Pull-request and issue templates",
    files: [
      ".github/pull_request_template.md",
      ".github/ISSUE_TEMPLATE/bug.yml",
      ".github/ISSUE_TEMPLATE/feature.yml",
      ".github/ISSUE_TEMPLATE/config.yml",
    ],
  },
  {
    name: "gitlab",
    kind: "non-core",
    purpose: "Merge-request and issue templates",
    files: [
      ".gitlab/merge_request_templates/Default.md",
      ".gitlab/issue_templates/Bug.md",
      ".gitlab/issue_templates/Feature.md",
    ],
  },
];

export const coreTools: ReadonlyArray<Tool> = tools.filter((tool) => tool.kind === "core");

export const nonCoreToolNames: ReadonlyArray<string> = tools
  .filter((tool) => tool.kind === "non-core")
  .map((tool) => tool.name);

export const toolsByName: ReadonlyMap<string, Tool> = new Map(tools.map((tool) => [tool.name, tool]));

/** Every target path to the one tool that renders it. */
export const toolByPath: ReadonlyMap<string, Tool> = new Map(
  tools.flatMap((tool) => tool.files.map((path) => [path, tool] as const)),
);
