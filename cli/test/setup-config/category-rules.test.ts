import { parse } from "@/setup-config/io";
import type { Running } from "@/setup-config/validity";
import type * as Catalog from "@/tool/catalog";
import {
  describe,
  expect,
  it,
} from "@effect/vitest";
import {
  Effect,
  Runtime,
} from "effect";
import { vi } from "vitest";
import { stringify } from "yaml";

/**
 * The shipped catalog has no max-one category with two tools, so this file adds two fixture
 * tools: a second tool manager and a second secret scanner.
 */
vi.mock("@/tool/catalog", async importOriginal => {
  const actual = await importOriginal<typeof Catalog>();
  const extra = (name: string, category: string): Catalog.Tool => ({
    name,
    category,
    purpose: "fixture",
    base: false,
    removable: true,
    replaceable: true,
    default: "off",
    requires: [],
    files: [],
  });
  const tools = [
    ...actual.tools,
    extra("alt-manager", "tool-manager"),
    extra("alt-scanner", "secret-scanner"),
  ];
  return { ...actual, tools, toolsByName: new Map(tools.map(tool => [tool.name, tool])) };
});

const running: Running = { version: "1.2.0", format: 1 };
const unremovable = ["mise", "git", "pre-commit"];

const failure = (tools: ReadonlyArray<string>, version = running.version) =>
  Effect.flip(parse(
    stringify({
      format: 1,
      version,
      values: {
        repo: "virajp/bootstrap",
        commit_scopes: [],
        merge_model: { develop: "direct", main: "pr" },
        tools,
      },
      files: [],
    }),
    running,
  ));

/** A Validity failure: the tag, what, why and fix, and exit code 3. */
const expectValidity = (
  error: { _tag: string; what: string; why: string; fix: string; },
  tag: string,
) => {
  expect(error._tag).toBe(tag);
  expect(error.what).not.toBe("");
  expect(error.why).not.toBe("");
  expect(error.fix).not.toBe("");
  expect((error as unknown as Record<string, unknown>)[Runtime.errorExitCode]).toBe(3);
};

describe("setup config validity over categories", () => {
  it.effect("refuses two tools of one max-one category", () =>
    Effect.gen(function*() {
      const error = yield* failure([...unremovable, "gitleaks", "alt-scanner"]);
      expectValidity(error, "MaxOneExceeded");
      expect(error.what).toContain("gitleaks");
      expect(error.what).toContain("alt-scanner");
    }));

  it.effect("refuses a max-one category that misses its unremovable tool and holds another", () =>
    Effect.gen(function*() {
      const error = yield* failure(["git", "pre-commit", "alt-manager"]);
      expectValidity(error, "UnremovableReplaced");
      expect(error.what).toContain("mise");
      expect(error.what).toContain("alt-manager");
    }));

  it.effect("fails that Validity with exit 3 also when the version is older", () =>
    Effect.gen(function*() {
      const error = yield* failure(["git", "pre-commit", "alt-manager"], "1.0.0");
      expectValidity(error, "UnremovableReplaced");
    }));
});
