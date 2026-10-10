/** The Tool catalog (1.0) of docs/blueprint/entities/tool/index.md: read-only data, no I/O. */

/** One file a tool renders. */
export interface ToolFile {
  /** Target-repository path, relative to the repository root. */
  readonly path: string;
  /** Written only when absent, never changed or deleted. */
  readonly createOnly: boolean;
}

export interface Tool {
  readonly name: string;
  /** Name of a category of docs/blueprint/entities/tool-category/index.md. */
  readonly category: string;
  readonly purpose: string;
  /** The tool's mise entries render into `conf.d/_base/` rather than `conf.d/<tool>/`. */
  readonly base: boolean;
  readonly removable: boolean;
  readonly replaceable: boolean;
  /** Whether the first-run default selection holds the tool. */
  readonly default: "on" | "off" | "origin";
  /** Hosts of the git remote `origin` that select the tool by default; present only when `default` is `origin`. */
  readonly originHosts?: ReadonlyArray<string>;
  /** Names of tools that must also be selected with this tool. */
  readonly requires: ReadonlyArray<string>;
  readonly files: ReadonlyArray<ToolFile>;
}

/** A plain target path. */
const file = (path: string): ToolFile => ({ path, createOnly: false });
/** The blueprint's `tasks/` prefix. */
const task = (path: string) => file(`.config/mise/tasks/${path}`);
/** The blueprint's `conf.d/` prefix. */
const conf = (path: string) => file(`.config/mise/conf.d/${path}`);

/** The tools in display order. */
export const tools: ReadonlyArray<Tool> = [
  {
    name: "mise",
    category: "tool-manager",
    purpose: "Installs the tools and runs the task library",
    base: true,
    removable: false,
    replaceable: false,
    default: "on",
    requires: [],
    files: [
      file(".config/mise.toml"),
      file(".config/miserc.toml"),
      conf("_base/mise.toml"),
      conf("_base/mise.dev.toml"),
      conf("_base/mise.ci.toml"),
      task("_scripts/helpers"),
      task("_scripts/checks"),
      task("_scripts/merge"),
      task("_scripts/placeholder"),
      task("code/all"),
      task("code/count"),
      task("code/worktrees"),
      task("code/merge/develop"),
      task("code/merge/main"),
      task("code/format/_default"),
      task("code/lint/_default"),
      task("code/sec/_default"),
      task("code/graph/_default"),
      task("setup/all"),
      task("setup/mise"),
      task("setup/worktree"),
      task("setup/cleanup"),
      task("setup/deps/all"),
      task("setup/deps/install/_default"),
      task("setup/deps/outdated/_default"),
      task("setup/deps/audit/_default"),
      task("setup/deps/upgrade/_default"),
      task("setup/deps/cleanup/_default"),
      task("setup/secrets/_default"),
      task("setup/ai/_default"),
    ],
  },
  {
    name: "git",
    category: "version-control",
    purpose: "Ignore rules and git checks",
    base: true,
    removable: false,
    replaceable: false,
    default: "on",
    requires: [],
    files: [file(".gitignore"), task("code/git-config")],
  },
  {
    name: "pre-commit",
    category: "git-hooks",
    purpose: "Runs the gates before each commit",
    base: true,
    removable: false,
    replaceable: false,
    default: "on",
    requires: [],
    files: [
      file(".config/pre-commit-config.yaml"),
      file(".config/git-conventional-commits.yaml"),
      task("code/precommit"),
      task("setup/precommit"),
    ],
  },
  {
    name: "vscode",
    category: "editor",
    purpose: "Editor settings and extensions",
    base: true,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".vscode/settings.json"), file(".vscode/extensions.json")],
  },
  {
    name: "dprint",
    category: "formatter",
    purpose: "Formats code and documents",
    base: true,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file("dprint.json"), task("code/format/dprint")],
  },
  {
    name: "taplo",
    category: "formatter",
    purpose: "Formats TOML files, run by dprint",
    base: true,
    removable: true,
    replaceable: true,
    default: "on",
    requires: ["dprint"],
    files: [file(".config/taplo.toml")],
  },
  {
    name: "virajp-linter",
    category: "linter",
    purpose: "Lints code with the house rules",
    base: true,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".config/linter.yaml"), file("eslint.config.mjs"), task("code/lint/virajp-linter")],
  },
  {
    name: "gitleaks",
    category: "secret-scanner",
    purpose: "Finds secrets in code and commits",
    base: false,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".config/gitleaks.toml"), conf("gitleaks/mise.dev.toml"), task("code/sec/gitleaks")],
  },
  {
    name: "grype",
    category: "vulnerability-scanner",
    purpose: "Finds known vulnerabilities",
    base: false,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".config/grype.yaml"), conf("grype/mise.dev.toml"), task("code/sec/grype")],
  },
  {
    name: "fnox",
    category: "secrets-manager",
    purpose: "Gives secrets to the tasks",
    base: false,
    removable: true,
    replaceable: true,
    default: "off",
    requires: [],
    files: [file(".config/fnox.toml"), conf("fnox/mise.dev.toml"), task("setup/secrets/fnox")],
  },
  {
    name: "claude",
    category: "ai-agent",
    purpose: "Status line and settings for Claude Code",
    base: false,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".config/claude-status.json"), task("setup/ai/claude")],
  },
  {
    name: "graphify",
    category: "knowledge-graph",
    purpose: "Builds a knowledge graph of the code",
    base: false,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [file(".graphifyignore"), conf("graphify/mise.dev.toml"), task("code/graph/graphify")],
  },
  {
    name: "mempalace",
    category: "agent-memory",
    purpose: "Memory store for AI agents",
    base: false,
    removable: true,
    replaceable: true,
    default: "on",
    requires: [],
    files: [{ path: "mempalace.yaml", createOnly: true }, conf("mempalace/mise.dev.toml")],
  },
  {
    name: "github",
    category: "forge",
    purpose: "Pull request and issue templates for GitHub",
    base: false,
    removable: true,
    replaceable: true,
    default: "origin",
    originHosts: ["github.com"],
    requires: [],
    files: [
      file(".github/pull_request_template.md"),
      file(".github/ISSUE_TEMPLATE/bug.yml"),
      file(".github/ISSUE_TEMPLATE/feature.yml"),
      file(".github/ISSUE_TEMPLATE/config.yml"),
    ],
  },
  {
    name: "gitlab",
    category: "forge",
    purpose: "Merge request and issue templates for GitLab",
    base: false,
    removable: true,
    replaceable: true,
    default: "origin",
    originHosts: ["gitlab.com"],
    requires: [],
    files: [
      file(".gitlab/merge_request_templates/Default.md"),
      file(".gitlab/issue_templates/Bug.md"),
      file(".gitlab/issue_templates/Feature.md"),
    ],
  },
];

export const toolsByName: ReadonlyMap<string, Tool> = new Map(tools.map(tool => [tool.name, tool]));

/** Every target path to the one tool that renders it; throws on a path in two tools (invariant 1). */
export const pathIndex = (from: ReadonlyArray<Tool>): ReadonlyMap<string, Tool> => {
  const index = new Map<string, Tool>();
  for (const tool of from) {
    for (const { path } of tool.files) {
      const owner = index.get(path);
      if (owner !== undefined) {
        throw new Error(`catalog path ${path} is in both ${owner.name} and ${tool.name}`);
      }
      index.set(path, tool);
    }
  }
  return index;
};

export const toolByPath: ReadonlyMap<string, Tool> = pathIndex(tools);

/** The tools that are always selected. */
export const unremovableTools: ReadonlyArray<Tool> = tools.filter(tool => !tool.removable);

/** The first-run selection: every `on` tool, plus each `origin` tool whose hosts hold `originHost`. */
export const defaultSelection = (originHost: string | undefined): ReadonlyArray<Tool> =>
  tools.filter(tool =>
    tool.default === "on"
    || (originHost !== undefined && tool.originHosts?.includes(originHost) === true)
  );

/** Each (tool, missing required tool) pair of a selection of tool names; unknown names are skipped. */
export const checkRequires = (
  selection: Iterable<string>,
): ReadonlyArray<{ readonly tool: string; readonly missing: string; }> => {
  const selected = new Set(selection);
  return [...selected].flatMap(name =>
    (toolsByName.get(name)?.requires ?? [])
      .filter(required => !selected.has(required))
      .map(missing => ({ tool: name, missing }))
  );
};
