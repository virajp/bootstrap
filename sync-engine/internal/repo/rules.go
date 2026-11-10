package repo

import (
	"os"
	"regexp"
	"strings"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/misegen"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// ds is one datasource of a gomplate call: `-c <name>=<path>`. Kind is
// "toml", "json" or "yaml".
type ds struct{ name, path, kind string }

// load reads the file of <d> as gomplate reads a datasource.
func load(d ds) (any, error) {
	if d.kind == "yaml" {
		out, _, err := format.YQ("-o", "json", d.path)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(out) == "" {
			return nil, nil
		}
		return tmpl.DecodeJSON([]byte(out))
	}
	b, err := os.ReadFile(d.path)
	if err != nil {
		return nil, err
	}
	if d.kind == "json" {
		return tmpl.DecodeJSON(b)
	}
	return tmpl.DecodeTOML(b)
}

// valuesDS gives the values file as the datasource of a gomplate call. The
// file is read once.
func (s *state) valuesDS() (any, error) {
	if s.valDS == nil && s.valErr == nil {
		s.valDS, s.valErr = load(ds{"v", s.values, "toml"})
		if s.valErr == nil && s.valDS == nil {
			s.valDS = map[string]any{}
		}
	}
	return s.valDS, s.valErr
}

// gomplate is `_se_gomplate -c <sources> -i <text>`: it renders the inline
// template <text>. <ctx> holds contexts that are already read. <env> holds
// the variables of the call, for env.Getenv. A failure prints the log line of
// gomplate to stderr, unless <quiet> (a call with `2> /dev/null`).
func (s *state) gomplate(text string, ctx map[string]any, sources []ds, env map[string]string, quiet bool) (string, bool) {
	fails := func(msg string) (string, bool) {
		if !quiet {
			gomplateError(s.stderr, msg)
		}
		return "", false
	}
	c := map[string]any{}
	for k, v := range ctx {
		c[k] = v
	}
	for _, d := range sources {
		v, err := load(d)
		if err != nil {
			return fails("failed to read datasource " + d.name + " (" + d.path + "): " + err.Error())
		}
		c[d.name] = v
	}
	out, err := tmpl.Render("<arg>", []byte(text), c, &tmpl.Options{Env: func(name string) string {
		if v, ok := env[name]; ok {
			return v
		}
		return s.getenv(name)
	}})
	if err != nil {
		return fails("renderTemplate: failed to render template <arg>: " + err.Error())
	}
	return string(out), true
}

// onValues is a gomplate call with the values file as `.v`.
func (s *state) onValues(text string, quiet bool) (string, bool) {
	v, err := s.valuesDS()
	if err != nil {
		if !quiet {
			gomplateError(s.stderr, "failed to read datasource v ("+s.values+"): "+err.Error())
		}
		return "", false
	}
	return s.gomplate(text, map[string]any{"v": v}, nil, nil, quiet)
}

// listTemplate is the template of `_rs_list` for <key>.
func listTemplate(key string) string {
	return `<% if has .v "` + key + `" %><% range (index .v "` + key + `") %><% . %>
<% end %><% end %>`
}

// list is `_rs_list <values> <key>`: the items of the list <key>, one per
// line, without the trailing newlines.
func (s *state) list(key string) (string, bool) {
	out, ok := s.onValues(listTemplate(key), false)
	return strings.TrimRight(out, "\n"), ok
}

// readEnvs is `_rs_envs`: it sets envs to the environments of the values
// file, in the order of the names of misegen. Without the key, only dev. It
// gives 2, after a message, for a name other than dev or ci, or no dev.
func (s *state) readEnvs() fail.Code {
	list, ok := s.list("environments")
	if !ok {
		s.log("cannot read " + s.show)
		return fail.Usage
	}
	if strings.Trim(list, "\n ") == "" {
		s.log(s.show + " lists no environments - using dev only; add environments = [\"dev\", ...]")
		list = "dev"
	}
	names := fields(list)
	for _, name := range names {
		if name != "dev" && name != "ci" {
			s.log("environment '" + name + "' in " + s.show + " is not one of dev, ci")
			return fail.Usage
		}
	}
	has := func(n string) bool {
		for _, x := range names {
			if x == n {
				return true
			}
		}
		return false
	}
	if !has("dev") {
		s.log("environments in " + s.show + " must include dev")
		return fail.Usage
	}
	s.envs = nil
	for _, n := range misegen.EnvNames {
		if has(n) {
			s.envs = append(s.envs, n)
		}
	}
	return fail.OK
}

func (s *state) hasEnv(env string) bool {
	for _, e := range s.envs {
		if e == env {
			return true
		}
	}
	return false
}

// readSel is `_rs_values_sel`: the `[tools.*]` and `[packs.*]` entries of
// the values file.
func (s *state) readSel() ([]values.Item, error) {
	v, err := values.ReadRepo(s.values, s.show)
	if err != nil {
		if _, derr := s.valuesDS(); derr != nil {
			gomplateError(s.stderr, "failed to read datasource v ("+s.values+"): "+derr.Error())
		}
		return nil, err
	}
	s.vals = v
	return v.Sel, nil
}

// envRules is `_rs_env_rules`: an `env` of the values file names only
// environments of this repo, and always dev: each tool must be in dev. It
// gives 2, after a message, at the first problem.
func (s *state) envRules(sel []values.Item) fail.Code {
	for _, it := range sel {
		// `read -r kind name envs` drops the blanks at the start and the end.
		envs := strings.Trim(it.Envs, " \t")
		if envs == "-" {
			continue
		}
		if envs == "" {
			s.log("env of [" + it.Kind + "." + it.Name + "] in " + s.show + " is empty - name an environment, or remove env to use default_envs")
			return fail.Usage
		}
		hasDev := strings.Contains(","+envs+",", ",dev,")
		// Split on commas only, so an item with a space fails the name check.
		for _, env := range strings.Split(envs, ",") {
			if !misegen.EnvName(env) {
				s.log("env of [" + it.Kind + "." + it.Name + "] in " + s.show + ": '" + env + "' is not one of " + strings.Join(misegen.EnvNames, ", "))
				return fail.Usage
			}
			if !s.hasEnv(env) {
				s.log("env '" + env + "' of [" + it.Kind + "." + it.Name + "] in " + s.show + " is not in environments: " + strings.Join(s.envs, " "))
				return fail.Usage
			}
		}
		if !hasDev {
			s.log("env of [" + it.Kind + "." + it.Name + "] in " + s.show + " has no dev - each tool must be in dev, for example env = [\"dev\", \"ci\"]")
			return fail.Usage
		}
	}
	return fail.OK
}

// branchExists is `_rs_branch_exists`: a branch that exists here or on
// origin. U3 (the --branch pairs of the no-commit-to-branch hook) needs it
// until the hook moves to the merge model.
func (s *state) branchExists(name string) bool {
	return run("", "git", "-C", s.root, "show-ref", "--verify", "--quiet", "refs/heads/"+name) ||
		run("", "git", "-C", s.root, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+name)
}

// overrides is `_rs_overrides`. It examines `[overrides]` of the values file:
// misegen examines the `mise` and `mise_env` keys, and this function the
// rest. Each line of `remove` of `[overrides.gitignore]` must be a line of
// the shipped file, and no line can be in `add` and in `remove`. It sets
// giAdd, giRemove and valuesNote. <local> is for `repo_sync --check`: it
// reads no clone, and it skips the problems that need the tools or the
// shipped files. It gives 2, after a message for each problem.
func (s *state) overrides(g *misegen.Generator, local bool) fail.Code {
	rc := fail.OK
	s.giAdd, s.giRemove = nil, nil
	s.valuesNote = strings.TrimPrefix(s.values, s.root+"/")
	if isFile(s.root + "/.config/repo-sync/gitignore") {
		s.log(".config/repo-sync/gitignore is no longer used - move its lines to add of [overrides.gitignore] in " + s.show + ", then delete it (M18 of docs/sync/migrate.md)")
		rc = fail.Usage
	}
	v, err := s.valuesDS()
	var out string
	ok := false
	if err != nil {
		gomplateError(s.stderr, "failed to read datasource o ("+s.values+"): "+err.Error())
	} else {
		out, ok = s.gomplate(tmplGitignore, map[string]any{"o": v}, nil, nil, false)
	}
	if !ok {
		s.log("cannot read [overrides] of " + s.show)
		return fail.Usage
	}
	for _, line := range bashLines(out) {
		problem, path, arg := readTabs(line)
		switch problem {
		case "table":
			s.log(path + " in " + s.show + " must be a table")
		case "key":
			if path == "overrides.graphifyignore" {
				s.log(path + " in " + s.show + " is no longer used: bootstrap does not ship .graphifyignore - remove the table (M24 of docs/sync/migrate.md)")
			} else {
				s.log(path + " in " + s.show + ": the only keys of [overrides] are mise, mise_env and gitignore")
			}
		case "gikey":
			s.log(path + " in " + s.show + ": the only keys of [" + path[:strings.LastIndex(path, ".")] + "] are add and remove")
		case "gilist":
			s.log(path + " in " + s.show + " must be a list of one-line strings, for example [\"/build/\"]")
		case "giempty":
			s.log(path + " in " + s.show + " has an empty line")
		case "add", "remove":
			// <path> is overrides.<ignore>.<list>.
			p := strings.TrimPrefix(path, "overrides.")
			if i := strings.LastIndex(p, "."); i >= 0 {
				p = p[:i]
			}
			if problem == "add" {
				s.giAdd = append(s.giAdd, giLine{p, arg})
			} else {
				s.giRemove = append(s.giRemove, giLine{p, arg})
			}
			continue
		default:
			continue
		}
		rc = fail.Usage
	}
	var tools []string
	if s.res != nil {
		tools = s.res.Tools
	}
	if s.logAll(g.Overrides(tools, s.envs, local)) != fail.OK {
		rc = fail.Usage
	}
	for _, r := range s.giRemove {
		if r.line == "" {
			continue
		}
		shipped := s.shippedPath(".gitignore")
		inAdd := false
		for _, a := range s.giAdd {
			if a == r {
				inAdd = true
			}
		}
		switch {
		case inAdd:
			s.log("overrides." + r.ignore + " in " + s.show + ": '" + r.line + "' is in add and in remove - keep one")
			rc = fail.Usage
		case local:
		case !shippedLine(s.dir+"/"+shipped, r.line):
			s.log("overrides." + r.ignore + ".remove in " + s.show + ": '" + r.line + "' is not a line of the shipped ." + r.ignore + " (" + shipped + ")")
			rc = fail.Usage
		}
	}
	return rc
}

// shippedPath gives the source of the tool file whose target is <target>, as
// `<side>/tools/<tool>/<source>`, or empty when no tool ships it.
func (s *state) shippedPath(target string) string {
	if s.idx == nil {
		return ""
	}
	for _, tool := range s.idx.Names {
		for _, f := range s.idx.Files(tool) {
			if f.Target == target {
				return toolsShown + "/" + tool + "/" + f.Path
			}
		}
	}
	return ""
}

// edgeBlanks are the blanks that awk `gsub(/^[ \t]+|[ \t]+$/, "")` removes.
var edgeBlanks = regexp.MustCompile(`^[ \t]+|[ \t]+$`)

// trimBlanks is the gsub of awk that removes the spaces and tabs at the start
// and the end of a line.
func trimBlanks(s string) string { return edgeBlanks.ReplaceAllString(s, "") }

// awkLines gives the lines of <b> as awk reads them.
func awkLines(b []byte) []string {
	return bashLines(string(b))
}

// shippedLine tells if a line of the file <path>, without its spaces and tabs
// at the start and the end, is <line>.
func shippedLine(path, line string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, l := range awkLines(b) {
		if trimBlanks(l) == line {
			return true
		}
	}
	return false
}

// coreTools is `_rs_core_tools`: a warning for `core_tools`, a key that is no
// longer used.
func (s *state) coreTools() {
	old, ok := s.list("core_tools")
	if ok && strings.Trim(old, "\n ") != "" {
		s.log("core_tools in " + s.show + " is no longer used - delete it; the env of each [tools.*] entry, else default_envs, decides its environments")
	}
}

// memberNameRe is the rule of a member name.
var memberNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// submodules is `git -C <root> submodule --quiet foreach --recursive 'echo
// $displaypath'`: the paths of the git submodules, one per line.
func (s *state) submodules() string {
	out, ok := git(s.root, "submodule", "--quiet", "foreach", "--recursive", "echo $displaypath")
	if !ok {
		return ""
	}
	return out
}

// lineIn is `grep -qxF -- <s> <<< <list>`.
func lineIn(s, list string) bool {
	for _, l := range strings.Split(list, "\n") {
		if l == s {
			return true
		}
	}
	return false
}

// member is one member of the repo: a line of `_rs_members`.
type member struct{ name, folder, kind string }

// members is `_rs_members`: the members of the repo, the git submodules
// first, then the other `[member.<name>]` tables. <name> starts each message.
// It gives 2, after a message, when a member breaks a rule.
func (s *state) members(name string) ([]member, fail.Code) {
	submodules := ""
	if isFile(s.root + "/.gitmodules") {
		submodules = s.submodules()
	}
	rowsText, ok := s.onValues(tmplMembers, false)
	if !ok {
		return nil, fail.Usage
	}
	var rows []member
	for _, line := range bashLines(rowsText) {
		key, folder, kind := readTabs(line)
		rows = append(rows, member{key, folder, kind})
	}
	show := s.show
	for _, r := range rows {
		if r.name == "" {
			continue
		}
		folder := strings.TrimSuffix(r.folder, "/")
		if folder == "." || folder == "" {
			s.log(name + ": [member." + r.name + "] in " + show + " has the folder of the repo itself - a member is a submodule or a folder below the root")
			return nil, fail.Usage
		}
		switch r.kind {
		case "", "submodule", "folder":
		default:
			s.log(name + ": [member." + r.name + "] in " + show + " has the type '" + r.kind + "' - use submodule or folder")
			return nil, fail.Usage
		}
		switch {
		case lineIn(folder, submodules):
			if r.kind == "folder" {
				s.log(name + ": [member." + r.name + "] in " + show + " has type folder, but " + folder + " is a submodule")
				return nil, fail.Usage
			}
		case r.kind == "submodule":
			s.log(name + ": [member." + r.name + "] in " + show + " has type submodule, but " + folder + " is not in .gitmodules")
			return nil, fail.Usage
		default:
			if info, err := os.Stat(s.root + "/" + folder); err != nil || !info.IsDir() {
				s.log(name + ": [member." + r.name + "] in " + show + " is not a submodule or a folder of the repo")
				return nil, fail.Usage
			}
		}
	}
	// The submodules, each with the name of its table, else the `repo_name`
	// of its own values file, else its path with each `/` changed to `_`.
	var out []member
	for _, path := range strings.Split(submodules, "\n") {
		if path == "" {
			continue
		}
		found := ""
		for _, r := range rows {
			if r.name != "" && strings.TrimSuffix(r.folder, "/") == path {
				found = r.name
			}
		}
		sub := s.root + "/" + path + "/.config/bootstrap.toml"
		if found == "" && isFile(sub) {
			found, _ = s.gomplate(tmplRepoName, nil, []ds{{"v", sub, "toml"}}, nil, false)
		}
		if found == "" {
			found = strings.ReplaceAll(path, "/", "_")
		}
		out = append(out, member{found, path, "submodule"})
	}
	// Then the folder members.
	for _, r := range rows {
		if r.name == "" {
			continue
		}
		folder := strings.TrimSuffix(r.folder, "/")
		if lineIn(folder, submodules) {
			continue
		}
		out = append(out, member{r.name, folder, "folder"})
	}
	// A name is a flag (`--<name>`) and a part of a shell variable of the
	// tasks.
	seen := map[string]bool{}
	for _, m := range out {
		if m.name == "" {
			continue
		}
		if !memberNameRe.MatchString(m.name) {
			s.log(name + ": the member " + m.folder + " has the name '" + m.name + "' - a member name has only letters, digits, - and _; use a [member.<name>] table with such a name in " + show)
			return nil, fail.Usage
		}
		if seen[m.name] {
			s.log(name + ": two members have the name '" + m.name + "' - write a [member.<name>] table with another name for one of them in " + show)
			return nil, fail.Usage
		}
		seen[m.name] = true
	}
	return out, fail.OK
}

// hasFile is `_rs_has_file`: this repo has a file named <name>, at the root
// or down to 4 folders below it, that git tracks or does not ignore. A git
// submodule is not searched.
func (s *state) hasFile(name string) bool {
	out, _ := git(s.root, "ls-files", "--cached", "--others", "--exclude-standard", "--", name, "*/"+name)
	for _, l := range strings.Split(out, "\n") {
		parts := strings.Split(l, "/")
		if l != "" && parts[len(parts)-1] == name && len(parts) <= 5 {
			return true
		}
	}
	return false
}

// inSubmodule is the test of `_rs_resolve` and `_rs_extensions`: this repo
// is a git submodule.
func (s *state) inSubmodule() bool {
	out, _ := git(s.root, "rev-parse", "--show-superproject-working-tree")
	return out != ""
}

// hookBranches sets hookDrop to the branches whose --branch pair leaves the
// no-commit-to-branch hook: each of main and develop with direct_commit =
// true in [tools.git.branch.<name>], and each one that does not exist. This
// is U3, which stays in the engine until the merge model replaces it. The
// check script of git examines the values.
func (s *state) hookBranches() {
	s.hookDrop = nil
	branch, _ := s.toolVals("git").Get("branch").(*source.Obj)
	for _, name := range []string{"main", "develop"} {
		if branch == nil {
			break
		}
		if b, ok := branch.Get(name).(*source.Obj); ok {
			if v, ok := b.Get("direct_commit").(bool); ok && v {
				s.hookDrop = append(s.hookDrop, name)
			}
		}
	}
	for _, name := range []string{"main", "develop"} {
		if !s.branchExists(name) {
			s.hookDrop = append(s.hookDrop, name)
		}
	}
}

// checkScripts is the part of `bootstrap check` that needs the clone: the
// selection, the values and the check script of each selected tool. Without
// BOOTSTRAP_CLONE, it skips them after a line that says so.
func (s *state) checkScripts() fail.Code {
	clone := s.getenv("BOOTSTRAP_CLONE")
	if clone == "" || !exists(clone+"/repo/tools") {
		s.log("no clone of bootstrap - the check scripts of the tools were skipped")
		return fail.OK
	}
	s.dir = clone
	s.src = source.Source{Root: clone, Side: source.Repo}
	idx, err := s.src.Open()
	s.idx = idx
	if err != nil {
		return s.logAll(err)
	}
	if code := s.resolve(); code != fail.OK {
		return code
	}
	if code := s.toolValues(); code != fail.OK {
		return code
	}
	return s.checks()
}
