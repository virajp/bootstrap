import { describe, expect, it } from "@effect/vitest";
import { coreTools, nonCoreToolNames, toolByPath, tools, toolsByName } from "@/tool/catalog";

const t = (path: string) => `.config/mise/tasks/${path}`;

/** The Catalog (1.0) and Target paths (1.0) tables of docs/blueprint/entities/tool/index.md. */
const expected: Record<string, { kind: "core" | "non-core"; paths: ReadonlyArray<string> }> = {
  mise: {
    kind: "core",
    paths: [
      ".config/mise.toml",
      ".config/miserc.toml",
      ".config/mise/conf.d/_base/mise.toml",
      ".config/mise/conf.d/_base/mise.dev.toml",
      ".config/mise/conf.d/_base/mise.ci.toml",
      t("_scripts/helpers"),
      t("_scripts/checks"),
      t("_scripts/merge"),
      t("_scripts/placeholder"),
      t("code/all"),
      t("code/count"),
      t("code/worktrees"),
      t("code/merge/develop"),
      t("code/merge/main"),
      t("setup/all"),
      t("setup/mise"),
      t("setup/worktree"),
      t("setup/cleanup"),
      t("setup/deps/all"),
      t("setup/deps/install"),
      t("setup/deps/outdated"),
      t("setup/deps/audit"),
      t("setup/deps/upgrade"),
      t("setup/deps/cleanup"),
    ],
  },
  git: { kind: "core", paths: [".gitignore", t("code/git-config")] },
  "pre-commit": {
    kind: "core",
    paths: [
      ".config/pre-commit-config.yaml",
      ".config/git-conventional-commits.yaml",
      t("code/precommit"),
      t("setup/precommit"),
    ],
  },
  vscode: { kind: "core", paths: [".vscode/settings.json", ".vscode/extensions.json"] },
  dprint: { kind: "core", paths: ["dprint.json", ".config/dprint.json", t("code/format")] },
  taplo: { kind: "core", paths: [".config/taplo.toml"] },
  gitleaks: { kind: "core", paths: [".config/gitleaks.toml", t("code/sec")] },
  grype: { kind: "core", paths: [".config/grype.yaml"] },
  "virajp-linter": { kind: "core", paths: [".config/linter.yaml", "eslint.config.mjs", t("code/lint")] },
  claude: {
    kind: "core",
    paths: [".config/mise/conf.d/ai/mise.dev.toml", ".config/claude-status.json", t("setup/ai")],
  },
  graphify: { kind: "core", paths: [t("code/graph"), ".graphifyignore"] },
  mempalace: { kind: "core", paths: ["mempalace.yaml"] },
  fnox: { kind: "core", paths: [".config/fnox.toml", t("setup/secrets")] },
  github: {
    kind: "non-core",
    paths: [
      ".github/pull_request_template.md",
      ".github/ISSUE_TEMPLATE/bug.yml",
      ".github/ISSUE_TEMPLATE/feature.yml",
      ".github/ISSUE_TEMPLATE/config.yml",
    ],
  },
  gitlab: {
    kind: "non-core",
    paths: [
      ".gitlab/merge_request_templates/Default.md",
      ".gitlab/issue_templates/Bug.md",
      ".gitlab/issue_templates/Feature.md",
    ],
  },
};

const allPaths = tools.flatMap((tool) => tool.files);

describe("tool catalog", () => {
  it("holds exactly the 13 core tools and the non-core github and gitlab", () => {
    expect(tools.map((tool) => tool.name).sort()).toEqual(Object.keys(expected).sort());
    expect(coreTools).toHaveLength(13);
    expect(coreTools.every((tool) => tool.kind === "core")).toBe(true);
    expect([...nonCoreToolNames].sort()).toEqual(["github", "gitlab"]);
  });

  it("gives every tool its kind, a purpose and exactly its pinned target paths", () => {
    for (const tool of tools) {
      const want = expected[tool.name];
      expect(tool.kind, tool.name).toBe(want?.kind);
      expect(tool.purpose.length, tool.name).toBeGreaterThan(0);
      expect([...tool.files].sort(), tool.name).toEqual([...(want?.paths ?? [])].sort());
    }
  });

  it("puts no path in two tools", () => {
    expect(new Set(allPaths).size).toBe(allPaths.length);
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
    expect(toolsByName.get("github")?.kind).toBe("non-core");
    expect(toolsByName.get("nope")).toBeUndefined();
    expect(toolByPath.get(".gitignore")?.name).toBe("git");
    expect(toolByPath.size).toBe(allPaths.length);
    for (const tool of tools) {
      for (const path of tool.files) expect(toolByPath.get(path)).toBe(tool);
    }
  });
});
