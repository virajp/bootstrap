package selection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// A tree maps a path in the clone to the content of the file.
type tree map[string]string

func (t tree) with(over tree) tree {
	out := tree{}
	for k, v := range t {
		out[k] = v
	}
	for k, v := range over {
		if v == "-" {
			delete(out, k)
		} else {
			out[k] = v
		}
	}
	return out
}

func writeTree(t *testing.T, files tree) string {
	t.Helper()
	root := t.TempDir()
	for p, c := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const devHead = "default_envs = [ \"dev\" ]\ndepends = []\nextensions = []\nroot_only = false\n"

// rt gives a repo tools.toml.
func rt(extra string) string { return devHead + extra }

// A small repo clone: three packs, four tools.
var repoBase = tree{
	"repo/packs.toml": `[pack.base]
a = { default = true }

[pack.extra]
b = { default = true }
c = {}

[pack.nested]
"pack:extra" = { default = true }
`,
	"repo/tools/a/tools.toml": rt(""),
	"repo/tools/b/tools.toml": rt(""),
	"repo/tools/c/tools.toml": rt(""),
	"repo/tools/d/tools.toml": rt(""),
}

const repoValues = "environments = [ \"dev\" ]\n"

// A small Mac clone.
var macosBase = tree{
	"macos/packs.toml": `[pack.base]
macos-setup = { default = true }
m1 = { default = true }

[pack.p]
m2 = { default = true }
m3 = {}
`,
	"macos/tools/macos-setup/tools.toml": "description = \"setup\"\n",
	"macos/tools/m1/tools.toml":          "description = \"one\"\n",
	"macos/tools/m2/tools.toml":          "description = \"two\"\n",
	"macos/tools/m3/tools.toml":          "description = \"three\"\n",
	"macos/tools/m4/tools.toml":          "description = \"four\"\n",
}

type repoCase struct {
	name   string
	over   tree
	values string
	sub    bool
}

var repoCases = []repoCase{
	{name: "plain", values: repoValues},
	{name: "pack", values: repoValues + "[packs.extra]\n"},
	{name: "pack-nested", values: repoValues + "[packs.nested]\nenv = [ \"dev\" ]\n"},
	{name: "tool-member-empty", values: repoValues + "[tools.c]\n[tools.d]\n"},
	{name: "unknown-pack", values: repoValues + "[packs.nope]\n"},
	{name: "unknown-tool", values: repoValues + "[tools.nope]\n"},
	{name: "no-packs-file", over: tree{"repo/packs.toml": "-"}, values: repoValues},
	{name: "packs-syntax-error", over: tree{"repo/packs.toml": "[pack.base\n"}, values: repoValues},
	{name: "tool-syntax-error", over: tree{"repo/tools/b/tools.toml": "default_envs = [\n"}, values: repoValues},
	{name: "values-syntax-error", values: "environments = [\n"},
	{name: "values-list-tools", values: repoValues + "tools = [ \"a\" ]\n"},
	{name: "values-list-packs", values: repoValues + "packs = [ \"extra\" ]\n"},
	{name: "values-tool-not-table", values: repoValues + "[tools]\na = 1\n"},
	{name: "values-env-not-list", values: repoValues + "[tools.c]\nenv = \"dev\"\n"},
	{name: "values-pack-key", values: repoValues + "[packs.extra]\nfoo = 1\n"},
	{name: "pack-unknown-member", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\nzzz = { default = true }\n"}, values: repoValues},
	{name: "pack-unknown-nested", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\"pack:zzz\" = { default = true }\n"}, values: repoValues},
	{name: "pack-loop", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\n[pack.x]\n\"pack:y\" = { default = true }\n\n[pack.y]\n\"pack:z\" = {}\n\n[pack.z]\n\"pack:x\" = { default = true }\n"}, values: repoValues},
	{name: "pack-self-loop", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\"pack:base\" = {}\n"}, values: repoValues},
	{name: "pack-old-list", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\ntools = [ \"b\" ]\n"}, values: repoValues},
	{name: "pack-old-depends", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\ndepends = [ \"b\" ]\n"}, values: repoValues},
	{name: "pack-bool-member", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\nb = true\n"}, values: repoValues},
	{name: "pack-old-p-form", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\"p:extra\" = { default = true }\n"}, values: repoValues},
	{name: "pack-bad-default", over: tree{"repo/packs.toml": "[pack.base]\na = { default = \"yes\" }\n"}, values: repoValues},
	{name: "pack-multiline-member", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\n[pack.base.b]\ndefault = true\n"}, values: repoValues},
	{name: "pack-is-tool-folder", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\n[pack.d]\nb = { default = true }\n"}, values: repoValues},
	{name: "pack-empty-table", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n\n[pack.e]\n"}, values: repoValues + "[packs.e]\n"},
	{name: "pack-comment-and-quotes", over: tree{"repo/packs.toml": "# top\n[pack.base] # base\n  a   =   { default = true }   # a\n'b' = { default = true }\n\"c\" = {}\n"}, values: repoValues},
	{name: "pack-default-false", over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\nb = { default = false }\n"}, values: repoValues},
	{name: "depends", over: tree{"repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"b\", \"c\" ]\nextensions = []\nroot_only = false\n", "repo/tools/b/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"c\" ]\nextensions = []\nroot_only = false\n"}, values: repoValues},
	{name: "depends-cycle", over: tree{"repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"b\" ]\n", "repo/tools/b/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"a\" ]\n"}, values: repoValues},
	{name: "depends-unknown", over: tree{"repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"zzz\" ]\n"}, values: repoValues},
	{name: "depends-not-list", over: tree{"repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = \"b\"\n"}, values: repoValues},
	{name: "mise-dup", over: tree{
		"repo/tools/a/tools.toml": rt("[mise.tools.ta]\nversion = \"1\"\n"),
		"repo/tools/b/tools.toml": rt("[mise.tools.ta]\nversion = \"2\"\n"),
	}, values: repoValues},
	{name: "missing-key", over: tree{"repo/tools/d/tools.toml": "RAW:default_envs = [ \"dev\" ]\ndepends = []\nextensions = []\n"}, values: repoValues},
	{name: "no-dev", over: tree{"repo/tools/d/tools.toml": "default_envs = [ \"ci\" ]\n"}, values: repoValues},
	{name: "bad-default-env", over: tree{"repo/tools/a/tools.toml": "default_envs = [ \"dev\", \"qa\" ]\n"}, values: repoValues},
	{name: "late-folder", over: tree{"repo/tools/zz-repo/tools.toml": rt("")}, values: repoValues},
	{name: "late-folder-after", over: tree{"repo/tools/zzz/tools.toml": rt("")}, values: repoValues},
	{name: "root-only", over: tree{"repo/tools/c/tools.toml": "default_envs = [ \"dev\" ]\nroot_only = true\n", "repo/packs.toml": "[pack.base]\na = { default = true }\nc = { default = true }\n"}, values: repoValues},
	{name: "root-only-sub", sub: true, over: tree{"repo/tools/c/tools.toml": "default_envs = [ \"dev\" ]\nroot_only = true\n", "repo/packs.toml": "[pack.base]\na = { default = true }\nc = { default = true }\n"}, values: repoValues},
	{name: "root-only-sub-tools", sub: true, over: tree{"repo/tools/c/tools.toml": "default_envs = [ \"dev\" ]\nroot_only = true\n"}, values: repoValues + "[tools.c]\n"},
	{name: "root-only-dep-sub", sub: true, over: tree{"repo/tools/c/tools.toml": "default_envs = [ \"dev\" ]\nroot_only = true\n", "repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"c\" ]\n"}, values: repoValues},
	{name: "old-keys", over: tree{"repo/tools/a/tools.toml": rt("seed = [ \"x\" ]\n[seed_requires]\nx = 1\n[mise.tools.ta]\nrequires = \"f\"\nfiles = [ \"g\" ]\nrequires_found = [ \"a/b\" ]\n[mise.tools.tb]\nrequires = \"h\"\nrequires_found = []\n[mise_env.ci.tools.ta]\nrequires_found = [ \"x\" ]\n")}, values: repoValues},
	{name: "merge-contribute", over: tree{
		"repo/tools/a/tools.toml": rt("[defaults]\nlist = [ 1, 2 ]\nflag = false\n[defaults.t]\nk = 1\n\n[contributes.d]\nlist = [ 2, 3 ]\n[contributes.d.t]\nj = 2\n"),
		"repo/tools/b/tools.toml": rt("[contributes.d]\nlist = [ 4 ]\nname = \"x\"\n"),
		"repo/tools/d/tools.toml": rt("[defaults]\nlist = [ 0 ]\nname = \"d\"\nflag = true\n"),
	}, values: repoValues + "[packs.extra]\n[tools.d]\n"},
	{name: "merge-contribute-ignored", over: tree{
		"repo/tools/a/tools.toml": rt("[contributes.d]\nx = 1\n"),
		"repo/tools/d/tools.toml": rt(""),
	}, values: repoValues},
	{name: "merge-contribute-conflict", over: tree{
		"repo/tools/a/tools.toml": rt("[contributes.d]\nname = \"a\"\n[contributes.d.t]\nk = 1\n"),
		"repo/tools/b/tools.toml": rt("[contributes.d]\nname = \"b\"\n[contributes.d.t]\nk = 2\n"),
		"repo/tools/d/tools.toml": rt(""),
	}, values: repoValues + "[packs.extra]\n[tools.d]\n"},
	{name: "merge-contribute-conflict3", over: tree{
		"repo/tools/a/tools.toml": rt("[contributes.d]\nname = \"a\"\n"),
		"repo/tools/b/tools.toml": rt("[contributes.d]\nother = 1\n"),
		"repo/tools/c/tools.toml": rt("[contributes.d]\nname = \"c\"\n"),
		"repo/tools/d/tools.toml": rt(""),
	}, values: repoValues + "[packs.extra]\n[tools.c]\n[tools.d]\n"},
	{name: "merge-contribute-type", over: tree{
		"repo/tools/a/tools.toml": rt("[contributes.d]\nx = [ 1 ]\n"),
		"repo/tools/b/tools.toml": rt("[contributes.d]\nx = 1\n"),
		"repo/tools/d/tools.toml": rt(""),
	}, values: repoValues + "[packs.extra]\n[tools.d]\n"},
	{name: "merge-contribute-same", over: tree{
		"repo/tools/a/tools.toml": rt("[contributes.d]\nx = 1\n"),
		"repo/tools/b/tools.toml": rt("[contributes.d]\nx = 1\n"),
		"repo/tools/d/tools.toml": rt(""),
	}, values: repoValues + "[packs.extra]\n[tools.d]\n"},
	{name: "merge-contribute-unknown", over: tree{"repo/tools/a/tools.toml": rt("[contributes.zzz]\nx = 1\n")}, values: repoValues},
	{name: "merge-shape", over: tree{"repo/tools/a/tools.toml": rt("defaults = 1\n")}, values: repoValues},
	{name: "merge-shape-questions", over: tree{"repo/tools/a/tools.toml": rt("[questions.q1]\nquestion = 1\ndetect = 2\nfoo = 3\n[questions.q2]\nexample = \"x\"\n[questions]\nq3 = 1\n")}, values: repoValues},
	{name: "merge-shape-questions-str", over: tree{"repo/tools/a/tools.toml": rt("questions = \"x\"\ncontributes = 1\n")}, values: repoValues},
	{name: "merge-shape-contributes", over: tree{"repo/tools/a/tools.toml": rt("[contributes]\nb = 1\n")}, values: repoValues},
	{name: "merge-user-replace", over: tree{
		"repo/tools/d/tools.toml": rt("[defaults]\nlist = [ 1 ]\nname = \"d\"\n[defaults.t]\nk = 1\n"),
	}, values: repoValues + "[tools.d]\nenv = [ \"dev\" ]\nlist = \"now a string\"\nname = [ 1 ]\nextra = 2.5\n[tools.d.t]\nk = 2\nm = 3\n"},
	{name: "merge-user-keep", over: tree{
		"repo/tools/a/tools.toml": rt("[defaults]\nlist = [ 1 ]\nname = \"a\"\n[defaults.t]\nk = 1\n"),
	}, values: repoValues + "[tools.a]\nlist = [ 1, 2 ]\nname = \"b\"\nnew = true\n[tools.a.t]\nk = 2\nm = 3\n"},
	{name: "merge-user-keep-type", over: tree{
		"repo/tools/a/tools.toml": rt("[defaults]\nlist = [ 1 ]\n[defaults.t]\nk = 1\n"),
	}, values: repoValues + "[tools.a]\nlist = \"x\"\n[tools.a.t]\nk = \"s\"\n"},
	{name: "merge-user-keep-dep", over: tree{
		"repo/tools/a/tools.toml": "default_envs = [ \"dev\" ]\ndepends = [ \"d\" ]\n",
		"repo/tools/d/tools.toml": rt("[defaults]\nn = 1\n"),
	}, values: repoValues + "[tools.d]\nn = \"s\"\n"},
	{name: "merge-user-member-not-base", over: tree{
		"repo/tools/c/tools.toml": rt("[defaults]\nn = 1\n"),
		"repo/packs.toml":         "[pack.base]\na = { default = true }\nc = {}\n",
	}, values: repoValues + "[tools.c]\nn = \"s\"\n"},
	{name: "merge-ordered-keys", over: tree{
		"repo/tools/a/tools.toml": rt("[defaults]\nz = 1\nb = 2\n[contributes.a]\nm = 3\naa = 4\n"),
	}, values: repoValues + "[tools.a]\nzz = 1\nc = 2\nb = 5\n"},
	{name: "merge-types", over: tree{
		"repo/tools/d/tools.toml": rt("[defaults]\ni = 1\nf = 1.5\ns = \"q\\\"\\\\\\n\\t<>&é\"\nb = true\nl = [ 1, \"a\", [ 2 ], { x = 1 } ]\n"),
	}, values: repoValues + "[tools.d]\n"},
}

// fill adds the four keys that bash needs in each repo tools.toml, when a case
// leaves one out. A text that starts with "RAW:" is used as it is.
func fill(text string) string {
	if strings.HasPrefix(text, "RAW:") {
		return strings.TrimPrefix(text, "RAW:")
	}
	head := ""
	for _, kv := range []string{"default_envs = [ \"dev\" ]\n", "depends = []\n", "extensions = []\n", "root_only = false\n"} {
		key := kv[:strings.Index(kv, " ")]
		if !strings.Contains("\n"+text, "\n"+key+" =") {
			head += kv
		}
	}
	return head + text
}

func TestRepoCases(t *testing.T) {
	for _, c := range repoCases {
		t.Run(c.name, func(t *testing.T) {
			over := tree{}
			for p, v := range c.over {
				if strings.HasPrefix(p, "repo/tools/") && strings.HasSuffix(p, "/tools.toml") && v != "-" {
					v = fill(v)
				}
				over[p] = v
			}
			clone := writeTree(t, repoBase.with(over))
			vdir := writeTree(t, tree{"bootstrap.toml": c.values})
			check(t, pcase{name: "repo-" + c.name, side: source.Repo, clone: clone,
				values: filepath.Join(vdir, "bootstrap.toml"), shown: "bootstrap.toml", sub: c.sub})
		})
	}
}

type macosCase struct {
	name   string
	over   tree
	values string // The content of macos.toml; "-" means no file.
}

var macosCases = []macosCase{
	{name: "plain", values: "-"},
	{name: "empty-file", values: ""},
	{name: "pack", values: "[packs.p]\n"},
	{name: "tool", values: "[tools.m3]\nk = 1\n"},
	{name: "unknown-tool", values: "[tools.nope]\n"},
	{name: "unknown-pack", values: "[packs.nope]\n"},
	{name: "bad-names", values: "[tools.\"a/b\"]\n[packs.\".x\"]\n[tools.\"a b\"]\n"},
	{name: "values-syntax-error", values: "[tools\n"},
	{name: "values-is-dir", values: "dir"},
	{name: "values-old-tools", values: "tools = [ \"m3\" ]\n"},
	{name: "values-tool-not-table", values: "[tools]\nm3 = 1\n"},
	{name: "values-tool-env", values: "[tools.m3]\nenv = [ \"dev\" ]\n"},
	{name: "values-pack-key", values: "[packs.p]\nx = 1\ny = 2\n"},
	{name: "values-pack-not-table", values: "[packs]\np = 1\n"},
	{name: "values-packs-list", values: "packs = [ \"p\" ]\n"},
	{name: "values-commit-int", values: "bootstrap_commit = 1\n"},
	{name: "values-commit-ok", values: "bootstrap_commit = \"abc\"\n[overrides.mise.x]\nk = 1\n[mise.settings]\nk = 1\n"},
	{name: "values-overrides", values: "[overrides.mise_env.ci]\n[overrides.foo]\nk = 1\n"},
	{name: "values-overrides-str", values: "overrides = 1\nmise = 2\n"},
	{name: "values-unknown-key", values: "foo = 1\n[tools.m3]\n"},
	{name: "values-many", values: "foo = 1\n[tools.m3]\nenv = [ \"x\" ]\n[tools.\"a/b\"]\n[packs.p]\nz = 1\n"},
	{name: "packs-syntax-error", over: tree{"macos/packs.toml": "[pack.base\n"}, values: "-"},
	{name: "depends", over: tree{
		"macos/tools/m1/tools.toml": "depends = [ \"m4\" ]\n",
		"macos/tools/m4/tools.toml": "depends = [ \"m3\" ]\n",
	}, values: "-"},
	{name: "depends-unknown", over: tree{"macos/tools/m1/tools.toml": "depends = [ \"nope\" ]\n"}, values: "-"},
	{name: "depends-bad-name", over: tree{"macos/tools/m1/tools.toml": "depends = [ \"a/b\" ]\n"}, values: "-"},
	{name: "depends-bad-type", over: tree{"macos/tools/m1/tools.toml": "depends = \"m4\"\n", "macos/tools/m2/tools.toml": "depends = [ 1, \"m4\" ]\n"}, values: "-"},
	{name: "depends-cycle", over: tree{
		"macos/tools/m1/tools.toml": "depends = [ \"m4\" ]\n",
		"macos/tools/m4/tools.toml": "depends = [ \"m1\" ]\n",
	}, values: "-"},
	{name: "stage-init", over: tree{
		"macos/tools/m2/tools.toml": "stage = \"init\"\ndepends = [ \"m4\" ]\n",
		"macos/tools/m4/tools.toml": "depends = [ \"m3\" ]\n",
	}, values: "[packs.p]\n"},
	{name: "stage-init-unselected", over: tree{"macos/tools/m3/tools.toml": "stage = \"init\"\n"}, values: "-"},
	{name: "stage-shell", over: tree{"macos/tools/m1/tools.toml": "stage = \"shell\"\n"}, values: "-"},
	{name: "stage-other", over: tree{"macos/tools/m1/tools.toml": "stage = \"late\"\n", "macos/tools/m2/tools.toml": "stage = 1\n"}, values: "-"},
	{name: "manifest-questions", over: tree{"macos/tools/m1/tools.toml": "[questions.a]\nquestion = 1\ndetect = 2\nzz = 3\n[questions.b]\nexample = \"x\"\n[questions]\nc = 1\n"}, values: "-"},
	{name: "manifest-questions-str", over: tree{"macos/tools/m1/tools.toml": "questions = 1\ncontributes = 2\n"}, values: "-"},
	{name: "manifest-contributes", over: tree{"macos/tools/m1/tools.toml": "[contributes]\nm2 = 1\n"}, values: "-"},
	{name: "manifest-syntax-error", over: tree{"macos/tools/m2/tools.toml": "a = [\n"}, values: "-"},
	{name: "no-manifest", over: tree{"macos/tools/m5/x.txt": "x"}, values: "-"},
	{name: "bad-folder-name", over: tree{"macos/tools/.hidden/tools.toml": ""}, values: "-"},
	{name: "late-folder", over: tree{"macos/tools/zz-macos/tools.toml": ""}, values: "-"},
	{name: "late-folder-after", over: tree{"macos/tools/zzz/tools.toml": ""}, values: "-"},
	{name: "no-tools", over: tree{"macos/tools/macos-setup/tools.toml": "-", "macos/tools/m1/tools.toml": "-", "macos/tools/m2/tools.toml": "-", "macos/tools/m3/tools.toml": "-", "macos/tools/m4/tools.toml": "-", "macos/tools/.keep": ""}, values: "-"},
	{name: "pack-is-tool", over: tree{"macos/packs.toml": "[pack.base]\nmacos-setup = { default = true }\n[pack.m2]\nm1 = { default = true }\n"}, values: "-"},
	{name: "merge-contribute", over: tree{
		"macos/tools/m1/tools.toml": "[defaults]\nlist = [ 1 ]\nk = 1\n[contributes.m2]\nlist = [ 2 ]\n",
		"macos/tools/m2/tools.toml": "[defaults]\nlist = [ 0 ]\nk = 2\n",
	}, values: "[packs.p]\n"},
	{name: "merge-contribute-unknown", over: tree{"macos/tools/m1/tools.toml": "[contributes.zzz]\nk = 1\n"}, values: "-"},
	{name: "merge-contribute-conflict", over: tree{
		"macos/tools/m1/tools.toml": "[contributes.m2]\nk = 1\n",
		"macos/tools/m3/tools.toml": "[contributes.m2]\nk = 2\n",
	}, values: "[packs.p]\n[tools.m3]\n"},
	{name: "merge-user", over: tree{"macos/tools/m3/tools.toml": "[defaults]\nk = 1\nl = [ 1 ]\n"}, values: "[tools.m3]\nk = \"s\"\nl = \"s\"\n"},
	{name: "merge-user-base", over: tree{"macos/tools/m1/tools.toml": "[defaults]\nk = 1\nl = [ 1 ]\n"}, values: "[tools.m1]\nk = 2\nl = [ 2 ]\nz = 1\n"},
	{name: "merge-user-base-type", over: tree{"macos/tools/m1/tools.toml": "[defaults]\nk = 1\n[defaults.t]\na = 1\n"}, values: "[tools.m1]\nk = \"s\"\nt = 1\n"},
}

func TestMacosCases(t *testing.T) {
	for _, c := range macosCases {
		t.Run(c.name, func(t *testing.T) {
			clone := writeTree(t, macosBase.with(c.over))
			cfg := t.TempDir()
			switch c.values {
			case "-":
			case "dir":
				if err := os.Mkdir(filepath.Join(cfg, "macos.toml"), 0o755); err != nil {
					t.Fatal(err)
				}
			default:
				if err := os.WriteFile(filepath.Join(cfg, "macos.toml"), []byte(c.values), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			check(t, pcase{name: "macos-" + c.name, side: source.Macos, clone: clone, values: cfg, shown: "macos.toml"})
		})
	}
}
