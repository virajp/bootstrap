package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// NameOK tells if <name> can be a tool or pack name: one word, not a path.
func NameOK(name string) bool {
	return name != "" && !strings.Contains(name, "/") && !strings.HasPrefix(name, ".") &&
		strings.IndexFunc(name, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
		}) < 0
}

// Index is the checked set of the tools.toml files of the clone: every tool,
// selected or not.
type Index struct {
	Source    Source
	Names     []string // The tools with a tools.toml, sorted.
	Manifests map[string]*Manifest
	// MiseOwner maps a mise tool name to the tool that defines it (repo side).
	MiseOwner map[string]string
	// StageInit is the set of the tools with `stage = "init"` (Mac side).
	StageInit map[string]bool
	// Categories is the categories file of the side.
	Categories *Categories
}

// Has tells if the index has the tool <name>.
func (x *Index) Has(name string) bool { _, ok := x.Manifests[name]; return ok }

// Open reads and checks the tools.toml files of the clone for the side of the
// source. It is `_rs_mise_index` on the repo side and `_ss_manifests` on the
// Mac side. The error carries each message, in the order of bash, and the
// highest code. The index is complete also after an error, as far as it was
// read.
func (s Source) Open() (*Index, error) {
	if s.Side == Macos {
		return s.openMacos()
	}
	return s.openRepo()
}

func lessThan(a, b string) bool { return a < b }

// firstFrom gives the first name that is not less than <limit>.
func firstFrom(names []string, limit string) string {
	for _, n := range names {
		if !lessThan(n, limit) {
			return n
		}
	}
	return ""
}

func (s Source) openRepo() (*Index, error) {
	x := &Index{Source: s, Manifests: map[string]*Manifest{}, MiseOwner: map[string]string{}}
	shown := s.ToolsShown()
	folders, err := s.ToolFolders()
	if err != nil {
		return x, fail.New(fail.Usage, "cannot list the tool folders of the clone")
	}
	if late := firstFrom(folders, "zz-repo"); late != "" {
		return x, fail.New(fail.Usage, "tool folder %s/%s sorts after zz-repo or is named zz-repo - rename it, because conf.d/zz-repo/ must load last", shown, late)
	}
	for _, f := range folders {
		if _, err := os.Stat(filepath.Join(s.ToolDir(f), "tools.toml")); err == nil {
			x.Names = append(x.Names, f)
		}
	}
	var loadErr error
	for _, tool := range x.Names {
		m, err := s.LoadManifest(tool)
		if err != nil {
			loadErr = err
			break
		}
		x.Manifests[tool] = m
	}
	if loadErr != nil {
		// Bash reads each file alone, to name the first one that fails. The
		// loop above stops at that file.
		return x, loadErr
	}
	cats, probs := s.ReadCategories()
	x.Categories = cats
	for _, tool := range x.Names {
		m := x.Manifests[tool]
		hasDev := false
		for _, e := range m.DefaultEnvs {
			if e == "dev" {
				hasDev = true
			}
		}
		if !hasDev {
			probs = append(probs, fail.New(fail.Usage, "default_envs of %s/%s/tools.toml has no dev - each tool must be in dev, for example default_envs = [\"dev\", \"ci\"]", shown, tool))
			return x, Problems(probs)
		}
		for _, mt := range m.Mise {
			if _, dup := x.MiseOwner[mt.Name]; dup {
				probs = append(probs, fail.New(fail.Usage, "mise tool '%s' is defined in %s/%s and in another tool - keep one", mt.Name, shown, tool))
				return x, Problems(probs)
			}
			x.MiseOwner[mt.Name] = tool
		}
		probs = append(probs, x.repoRules(tool, m)...)
		probs = append(probs, x.catalogRules(tool)...)
		probs = append(probs, x.categoryRule(tool)...)
		probs = append(probs, m.probs...)
		probs = append(probs, x.parentRule(tool)...)
	}
	return x, Problems(probs)
}

// repoRules is `_rs_manifest_rules`: the old keys, the `[files]` tables and
// `requires_found` of a mise entry of one tool.
func (x *Index) repoRules(tool string, m *Manifest) []*fail.Error {
	at := x.Source.ManifestShown(tool)
	var out []*fail.Error
	add := func(format string, args ...any) { out = append(out, fail.New(fail.Usage, format, args...)) }
	seen := map[string]bool{}
	for _, l := range m.lines {
		switch l.field {
		case "bad":
			add("%s: %s", at, l.a)
		case "old":
			if seen[l.a] {
				continue
			}
			seen[l.a] = true
			switch l.a {
			case "seed":
				add("%s: seed is not supported; use [files.\"<target>\"] with mode = \"seed\"", at)
			case "seed_requires":
				add("%s: [seed_requires] is not supported; use requires in [files.\"<target>\"]", at)
			case "files":
				add("%s: files in [mise.tools.%s] is not supported; use requires_found in [files.\"<target>\"]", at, l.b)
			case "requires":
				add("%s: requires of mise tool %s is the old key - write requires_found = [\"%s\"]", at, l.b, l.c)
			case "extensions":
				add("%s: extensions is the old key - write extensions = [...] in [contributes.vscode]", at)
			case "precommit":
				add("%s: [[precommit]] is the old form - write [[contributes.pre-commit.repos]] with an id", at)
			}
		case "mise_env_key":
			add("%s: %s of mise tool %s in [mise_env.%s.tools] is valid only in [mise.tools]", at, l.a, l.b, l.c)
		case "mise_found":
			if l.b == "" || strings.Contains(l.b, "/") {
				add("%s: requires_found of mise tool %s: '%s' must be a file name, with no /", at, l.a, l.b)
			}
		}
	}
	return out
}

// badRelPath is the case `” | /* | .. | ../* | */../* | */..` of bash.
func badRelPath(p string) bool {
	return p == "" || strings.HasPrefix(p, "/") || p == ".." || strings.HasPrefix(p, "../") ||
		strings.Contains(p, "/../") || strings.HasSuffix(p, "/..")
}

func (s Source) openMacos() (*Index, error) {
	x := &Index{Source: s, Manifests: map[string]*Manifest{}, StageInit: map[string]bool{}}
	shown := s.ToolsShown()
	folders, err := s.ToolFolders()
	if err != nil {
		folders = nil
	}
	for _, name := range folders {
		if !NameOK(name) {
			return x, fail.New(fail.Usage, "%s/%s: a tool folder name has no whitespace and does not start with a dot", shown, name)
		}
		if !s.HasManifest(name) {
			return x, fail.New(fail.Usage, "%s/%s has no tools.toml - each tool folder needs its manifest", shown, name)
		}
	}
	x.Names = folders
	if len(x.Names) == 0 {
		return x, fail.New(fail.Usage, "the clone has no tool in %s", shown)
	}
	if late := firstFrom(x.Names, "zz-macos"); late != "" {
		return x, fail.New(fail.Usage, "tool folder %s/%s sorts after zz-macos or is named zz-macos - rename it, because conf.d/zz-macos/ must load last", shown, late)
	}
	for _, name := range x.Names {
		m, err := s.LoadManifest(name)
		if err != nil {
			return x, err
		}
		x.Manifests[name] = m
	}
	cats, probs := s.ReadCategories()
	x.Categories = cats
	for _, name := range x.Names {
		probs = append(probs, x.macosRules(name)...)
		probs = append(probs, x.catalogRules(name)...)
		probs = append(probs, x.categoryRule(name)...)
		probs = append(probs, x.Manifests[name].probs...)
		probs = append(probs, x.parentRule(name)...)
	}
	return x, Problems(probs)
}

// macosRules is the template and the loop of `_ss_manifests` for one tool.
func (x *Index) macosRules(tool string) []*fail.Error {
	m := x.Manifests[tool]
	at := x.Source.ManifestShown(tool)
	var out []*fail.Error
	add := func(format string, args ...any) { out = append(out, fail.New(fail.Usage, format, args...)) }
	bad := func(format string, args ...any) { add("%s: %s", at, fmt.Sprintf(format, args...)) }
	for _, k := range m.Raw.Keys() {
		v := m.Raw.Get(k)
		switch k {
		case "depends":
			l, ok := v.([]any)
			if !ok {
				bad("depends must be a list of tool names")
				break
			}
			for _, e := range l {
				if _, ok := e.(string); !ok {
					bad("depends must be a list of tool names")
				}
			}
		case "stage":
			s, ok := v.(string)
			switch {
			case !ok:
				bad("stage must be a string - the only stage is \"init\"")
			case s == "init":
				x.StageInit[tool] = true
			case s == "shell":
				bad("stage = \"shell\" is the old name - write stage = \"init\"")
			default:
				bad("stage = \"%s\" - the only stage is \"init\"", s)
			}
		case "description":
			if _, ok := v.(string); !ok {
				bad("description must be a string")
			}
		case "summary", "category", "selectable":
			// See catalogRules.
		case "defaults":
			if _, ok := v.(*Obj); !ok {
				bad("[defaults] must be a table")
			}
		case "mise":
			if _, ok := v.(*Obj); !ok {
				bad("[mise] must be a table")
			}
		case "questions":
			qs, ok := v.(*Obj)
			if !ok {
				bad("[questions] must be a table of [questions.<key>] tables")
				break
			}
			for _, q := range qs.Keys() {
				e, ok := qs.Get(q).(*Obj)
				if !ok {
					bad("[questions.%s] must be a table", q)
					continue
				}
				for _, qk := range e.Keys() {
					if !QuestionKey(qk) {
						bad("[questions.%s] %s is not a key of a question - the keys are %s", q, qk, QuestionKeys)
					}
				}
				for _, text := range QuestionTypes(q, e) {
					bad("%s", text)
				}
				if !e.Has("question") {
					bad("[questions.%s] needs a question string", q)
				} else if _, ok := e.Get("question").(string); !ok {
					bad("[questions.%s] question must be a string", q)
				}
				if e.Has("detect") {
					if _, ok := e.Get("detect").(string); !ok {
						bad("[questions.%s] detect must be a string (a shell command)", q)
					}
				}
			}
		case "contributes":
			cs, ok := v.(*Obj)
			if !ok {
				bad("[contributes] must be a table of [contributes.<tool>] tables")
				break
			}
			for _, o := range cs.Keys() {
				if _, ok := cs.Get(o).(*Obj); !ok {
					bad("[contributes.%s] must be a table", o)
				}
			}
		case "files", "tasks":
			// The entries: see entries.
		case "parent":
			// See parentRule.
		case "extensions", "precommit":
			bad("%s is the old key - the values go in [contributes.vscode] and [[contributes.pre-commit.repos]]", k)
		case "seed":
			bad("seed is the old key - write a [files.\"<target>\"] table with mode = \"seed\" for each seed")
		case "seed_requires":
			bad("[seed_requires] is the old table - write requires = [...] in the [files.\"<target>\"] table of the seed")
		case "default_envs", "root_only", "mise_env", "formatter", "hosts":
			bad("%s is for repo tools only", k)
		default:
			bad("%s is not a key of tools.toml - the keys are description, summary, category, selectable, stage, depends, parent, [defaults], [questions], [contributes], [files], [tasks] and [mise]", k)
		}
	}
	return out
}

// parentRule examines `parent`: the name of another tool of the clone.
func (x *Index) parentRule(tool string) []*fail.Error {
	m := x.Manifests[tool]
	v, has := m.Raw.Lookup("parent")
	if !has {
		return nil
	}
	p, ok := v.(string)
	if !ok || p == tool || !x.Has(p) {
		return []*fail.Error{fail.New(fail.Usage, "%s: parent must be the name of another tool of %s", x.Source.ManifestShown(tool), x.Source.ToolsShown())}
	}
	return nil
}

// QuestionKeys names the keys of a `[questions.<key>]` table, as a message
// lists them.
const QuestionKeys = "question, detect, example, group, note and optional"

// QuestionKey tells if <k> is a key of a `[questions.<key>]` table.
func QuestionKey(k string) bool {
	switch k {
	case "question", "detect", "example", "group", "note", "optional":
		return true
	}
	return false
}

// QuestionTypes gives a message for each wrong value of the keys group,
// note and optional of the question <q>, and for a question that has a
// group and optional = true.
func QuestionTypes(q string, e *Obj) []string {
	var out []string
	if e.Has("group") {
		if _, ok := e.Get("group").(string); !ok {
			out = append(out, fmt.Sprintf("[questions.%s] group must be a string", q))
		}
	}
	if e.Has("note") {
		if _, ok := e.Get("note").(string); !ok {
			out = append(out, fmt.Sprintf("[questions.%s] note must be a string", q))
		}
	}
	if e.Has("optional") {
		b, ok := e.Get("optional").(bool)
		if !ok {
			out = append(out, fmt.Sprintf("[questions.%s] optional must be true or false", q))
		} else if b && e.Has("group") {
			out = append(out, fmt.Sprintf("[questions.%s] has a group, so it cannot be optional", q))
		}
	}
	return out
}
