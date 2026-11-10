package repo

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/misegen"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/tmpl"
	"github.com/virajp/bootstrap/cli/internal/values"
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

// overrides is `_rs_overrides`. It examines `[overrides]` of the values file:
// misegen examines the `mise` and `mise_env` keys, and this function the
// rest. An old form stops the run.
// <local> is for `repo_sync --check`: it reads no clone, and it skips the
// problems that need the tools or the shipped files. It gives 2, after a
// message for each problem.
func (s *state) overrides(g *misegen.Generator, local bool) fail.Code {
	rc := fail.OK
	v, err := s.valuesDS()
	if err != nil {
		gomplateError(s.stderr, "failed to read datasource o ("+s.values+"): "+err.Error())
		s.log("cannot read [overrides] of " + s.show)
		return fail.Usage
	}
	if m, ok := v.(map[string]any); ok && m["overrides"] != nil {
		o, isMap := m["overrides"].(map[string]any)
		if !isMap {
			s.log("overrides in " + s.show + " must be a table")
			rc = fail.Usage
		}
		var keys []string
		for k := range o {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			switch k {
			case "mise", "mise_env":
			case "gitignore":
				s.log("[overrides.gitignore] in " + s.show + " is the old form - " + oldform.Hint)
				rc = fail.Usage
			case "graphifyignore":
				s.log("overrides." + k + " in " + s.show + " is no longer used: bootstrap does not ship .graphifyignore - remove the table (M24 of docs/sync/migrate.md)")
				rc = fail.Usage
			default:
				s.log("overrides." + k + " in " + s.show + ": the only keys of [overrides] are mise and mise_env")
				rc = fail.Usage
			}
		}
	}
	var tools []string
	if s.res != nil {
		tools = s.res.Tools
	}
	if s.logAll(g.Overrides(tools, s.envs, local)) != fail.OK {
		rc = fail.Usage
	}
	return rc
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
