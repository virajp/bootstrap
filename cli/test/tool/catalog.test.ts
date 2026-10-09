import { describe, expect, it } from "@effect/vitest";
import {
  checkRequires,
  defaultSelection,
  pathIndex,
  type Tool,
  toolByPath,
  tools,
  toolsByName,
  unremovableTools,
} from "@/tool/catalog";
import { categoriesByName } from "@/tool-category/catalog";

const t = (path: string) => `.config/mise/tasks/${path}`;
const c = (path: string) => `.config/mise/conf.d/${path}`;

/** The Catalog (1.0) table of docs/blueprint/entities/tool/index.md, in display order. */
const catalog = [
  ["mise", "tool-manager", "Installs the tools and runs the task library", true, false, false, "on", [], []],
  ["git", "version-control", "Ignore rules and git checks", true, false, false, "on", [], []],
  ["pre-commit", "git-hooks", "Runs the gates before each commit", true, false, false, "on", [], []],
  ["vscode", "editor", "Editor settings and extensions", true, true, true, "on", [], []],
  ["dprint", "formatter", "Formats code and documents", true, true, true, "on", [], []],
  ["taplo", "formatter", "Formats TOML files, run by dprint", true, true, true, "on", [], ["dprint"]],
  ["virajp-linter", "linter", "Lints code with the house rules", true, true, true, "on", [], []],
  ["gitleaks", "secret-scanner", "Finds secrets in code and commits", false, true, true, "on", [], []],
  ["grype", "vulnerability-scanner", "Finds known vulnerabilities", false, true, true, "on", [], []],
  ["fnox", "secrets-manager", "Gives secrets to the tasks", false, true, true, "off", [], []],
  ["claude", "ai-agent", "Status line and settings for Claude Code", false, true, true, "on", [], []],
  ["graphify", "knowledge-graph", "Builds a knowledge graph of the code", false, true, true, "on", [], []],
  ["mempalace", "agent-memory", "Memory store for AI agents", false, true, true, "on", [], []],
  [
    "github",
    "forge",
    "Pull request and issue templates for GitHub",
    false,
    true,
    true,
    "origin",
    ["github.com"],
    [],
  ],
  [
    "gitlab",
    "forge",
    "Merge request and issue templates for GitLab",
    false,
    true,
    true,
    "origin",
    ["gitlab.com"],
    [],
  ],
] as const;

/** The Target paths (1.0) table of docs/blueprint/entities/tool/index.md, in table order. */
const paths: Record<string, ReadonlyArray<string>> = {
  mise: [
    ".config/mise.toml",
    ".config/miserc.toml",
    c("_base/mise.toml"),
    c("_base/mise.dev.toml"),
    c("_base/mise.ci.toml"),
    t("_scripts/helpers"),
    t("_scripts/checks"),
    t("_scripts/merge"),
    t("_scripts/placeholder"),
    t("code/all"),
    t("code/count"),
    t("code/worktrees"),
    t("code/merge/develop"),
    t("code/merge/main"),
    t("code/format/_default"),
    t("code/lint/_default"),
    t("code/sec/_default"),
    t("code/graph/_default"),
    t("setup/all"),
    t("setup/mise"),
    t("setup/worktree"),
    t("setup/cleanup"),
    t("setup/deps/all"),
    t("setup/deps/install/_default"),
    t("setup/deps/outdated/_default"),
    t("setup/deps/audit/_default"),
    t("setup/deps/upgrade/_default"),
    t("setup/deps/cleanup/_default"),
    t("setup/secrets/_default"),
    t("setup/ai/_default"),
  ],
  git: [".gitignore", t("code/git-config")],
  "pre-commit": [
    ".config/pre-commit-config.yaml",
    ".config/git-conventional-commits.yaml",
    t("code/precommit"),
    t("setup/precommit"),
  ],
  vscode: [".vscode/settings.json", ".vscode/extensions.json"],
  dprint: ["dprint.json", t("code/format/dprint")],
  taplo: [".config/taplo.toml"],
  "virajp-linter": [".config/linter.yaml", "eslint.config.mjs", t("code/lint/virajp-linter")],
  gitleaks: [".config/gitleaks.toml", c("gitleaks/mise.dev.toml"), t("code/sec/gitleaks")],
  grype: [".config/grype.yaml", c("grype/mise.dev.toml"), t("code/sec/grype")],
  fnox: [".config/fnox.toml", c("fnox/mise.dev.toml"), t("setup/secrets/fnox")],
  claude: [".config/claude-status.json", t("setup/ai/claude")],
  graphify: [".graphifyignore", c("graphify/mise.dev.toml"), t("code/graph/graphify")],
  mempalace: ["mempalace.yaml", c("mempalace/mise.dev.toml")],
  github: [
    ".github/pull_request_template.md",
    ".github/ISSUE_TEMPLATE/bug.yml",
    ".github/ISSUE_TEMPLATE/feature.yml",
    ".github/ISSUE_TEMPLATE/config.yml",
  ],
  gitlab: [
    ".gitlab/merge_request_templates/Default.md",
    ".gitlab/issue_templates/Bug.md",
    ".gitlab/issue_templates/Feature.md",
  ],
};

/** The only `create_only` path of the Target paths table. */
const createOnly = ["mempalace.yaml"];

const allPaths = tools.flatMap(tool => tool.files.map(file => file.path));
const names = (selection: ReadonlyArray<Tool>) => selection.map(tool => tool.name);
const categoryOf = (tool: Tool) => categoriesByName.get(tool.category);

describe("tool catalog", () => {
  it("holds the 15 tools in display order, each equal to its Catalog row", () => {
    expect(
      tools.map(tool => [
        tool.name,
        tool.category,
        tool.purpose,
        tool.base,
        tool.removable,
        tool.replaceable,
        tool.default,
        tool.originHosts ?? [],
        tool.requires,
      ]),
    ).toEqual(catalog);
  });

  it("gives every tool exactly its Target paths, with create_only only on mempalace.yaml", () => {
    for (const tool of tools) {
      expect(tool.files.map(file => file.path), tool.name).toEqual(paths[tool.name]);
    }
    expect(tools.flatMap(tool => tool.files.filter(file => file.createOnly).map(file => file.path)))
      .toEqual(createOnly);
  });

  it("puts no path in two tools (invariant 1)", () => {
    expect(new Set(allPaths).size).toBe(allPaths.length);
  });

  it("names an existing category for every tool (invariant 3)", () => {
    for (const tool of tools) expect(categoryOf(tool), tool.name).toBeDefined();
  });

  it("keeps every unremovable tool unreplaceable and default on (invariants 2, 7)", () => {
    for (const tool of unremovableTools) {
      expect(tool.replaceable, tool.name).toBe(false);
      expect(tool.default, tool.name).toBe("on");
    }
  });

  it("selects at most one tool of a max-one category by default, for any host (invariants 2, 4)", () => {
    const hosts = [undefined, "example.org", ...tools.flatMap(tool => tool.originHosts ?? [])];
    for (const host of hosts) {
      const selection = defaultSelection(host);
      for (const tool of unremovableTools) expect(selection, String(host)).toContain(tool);
      const oneCategories = selection.map(categoryOf).filter(category => category?.max === "one");
      expect(new Set(oneCategories).size, String(host)).toBe(oneCategories.length);
    }
  });

  it("carries origin_hosts only on origin tools, each with at least one host (invariant 8)", () => {
    for (const tool of tools) {
      expect(tool.originHosts !== undefined, tool.name).toBe(tool.default === "origin");
      if (tool.originHosts !== undefined) expect(tool.originHosts.length, tool.name).toBeGreaterThan(0);
    }
  });

  it("shares no host between origin tools of one max-one category (invariant 9)", () => {
    const seen = new Set<string>();
    for (const tool of tools.filter(tool => categoryOf(tool)?.max === "one")) {
      for (const host of tool.originHosts ?? []) {
        const key = `${tool.category} ${host}`;
        expect(seen.has(key), key).toBe(false);
        seen.add(key);
      }
    }
  });

  it("gives every tool a boolean base (invariant 10)", () => {
    for (const tool of tools) expect(typeof tool.base, tool.name).toBe("boolean");
  });

  it("names only catalog tools in requires (invariant 11)", () => {
    for (const tool of tools) {
      for (const name of tool.requires) expect(toolsByName.has(name), `${tool.name} → ${name}`).toBe(true);
    }
  });

  it("holds at most one default-on and one unremovable tool per max-one category (invariant 12)", () => {
    for (const tool of tools) {
      if (categoryOf(tool)?.max !== "one") continue;
      const own = tools.filter(other => other.category === tool.category);
      expect(own.filter(other => other.default === "on").length, tool.category).toBeLessThanOrEqual(1);
      expect(own.filter(other => !other.removable).length, tool.category).toBeLessThanOrEqual(1);
    }
  });

  it("puts the setup config in no tool", () => {
    expect(allPaths).not.toContain(".config/bootstrap.yaml");
  });

  it("uses only relative paths with no .. segment and no wildcard", () => {
    for (const path of allPaths) {
      expect(path.startsWith("/"), path).toBe(false);
      expect(path.split("/"), path).not.toContain("..");
      expect(path, path).not.toMatch(/[*?[\]{}]/);
    }
  });

  it("looks a tool up by name and a path up to its tool", () => {
    expect(toolsByName.get("github")?.name).toBe("github");
    expect(toolsByName.get("nope")).toBeUndefined();
    expect(toolByPath.get(".gitignore")?.name).toBe("git");
    expect(toolByPath.size).toBe(allPaths.length);
    for (const tool of tools) {
      for (const file of tool.files) expect(toolByPath.get(file.path)).toBe(tool);
    }
  });

  it("fails to build the path index on a duplicate path", () => {
    const [first, second] = tools as readonly [Tool, Tool];
    const clash: Tool = { ...second, files: [...second.files, first.files[0]!] };
    expect(() => pathIndex([first, clash])).toThrow(first.files[0]!.path);
  });
});

describe("catalog functions", () => {
  it("defaultSelection holds every on tool and the origin tool of the host, never an off tool", () => {
    const github = defaultSelection("github.com");
    expect(names(github)).toEqual(
      tools.filter(tool => tool.default === "on").map(tool => tool.name).concat("github"),
    );
    expect(names(github)).not.toContain("fnox");
    expect(names(defaultSelection("gitlab.com"))).toContain("gitlab");
    expect(names(defaultSelection("gitlab.com"))).not.toContain("github");
    for (const host of ["example.org", undefined]) {
      expect(defaultSelection(host).filter(tool => tool.default === "origin"), String(host)).toEqual([]);
    }
  });

  it("unremovableTools equals the tools with removable false", () => {
    expect(unremovableTools).toEqual(tools.filter(tool => !tool.removable));
  });

  it("checkRequires returns each tool whose required tool is missing", () => {
    expect(checkRequires(["mise", "taplo"])).toEqual([{ tool: "taplo", missing: "dprint" }]);
    expect(checkRequires(["taplo", "dprint"])).toEqual([]);
    expect(checkRequires(["mise", "git"])).toEqual([]);
  });

  it("checkRequires ignores a name the catalog does not ship", () => {
    expect(checkRequires(["nope"])).toEqual([]);
  });
});
