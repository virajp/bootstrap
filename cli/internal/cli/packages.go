package cli

// The package scan of `bootstrap mac detect`. The Mac side runs no detect
// script. It lists the installed packages of the Mac and maps each one to the
// Mac tool that ships it, through the keys of [mise.bootstrap.packages] and
// [mise.tools] of each tools.toml. See docs/design/cli.md ("detect").

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
)

// The kinds of a package: a key of [mise.bootstrap.packages], or a key of
// [mise.tools].
const (
	kindPackage = "package"
	kindTool    = "tool"
)

// pkg is one installed package: its kind and its key.
type pkg struct{ kind, key string }

// table gives the [mise] table of macos.toml that holds <p>.
func (p pkg) table() string {
	if p.kind == kindTool {
		return "[mise.tools]"
	}
	return "[mise.bootstrap.packages]"
}

// expr gives the yq expression that adds <p> to macos.toml.
func (p pkg) expr() string {
	if p.kind == kindTool {
		return ".mise.tools[" + q(p.key) + "].version = \"latest\""
	}
	return ".mise.bootstrap.packages[" + q(p.key) + "] = \"latest\""
}

// The prefixes of the keys of [mise.bootstrap.packages] that the scan lists.
const (
	prefixFormula = "brew:"
	prefixCask    = "brew-cask:"
	prefixApp     = "mas:"
)

// match gives the form of <p> that the scan compares. brew prints a formula
// or a cask of a tap without its tap, so a brew: or brew-cask: key keeps only
// the part after its last "/".
func (p pkg) match() pkg {
	if p.kind == kindPackage {
		for _, pre := range []string{prefixFormula, prefixCask} {
			if name, ok := strings.CutPrefix(p.key, pre); ok {
				return pkg{p.kind, pre + name[strings.LastIndex(name, "/")+1:]}
			}
		}
	}
	return p
}

// The command of the App Store apps. The mas: entries of mise need it.
const appStoreCommand = "mas"

// packageKeys gives the keys of the [mise.bootstrap.packages] and the
// [mise.tools] tables of a tools.toml or of the values file.
func packageKeys(raw *source.Obj) []pkg {
	var out []pkg
	m, _ := raw.Get("mise").(*source.Obj)
	if m == nil {
		return nil
	}
	if b, ok := m.Get("bootstrap").(*source.Obj); ok {
		if p, ok := b.Get("packages").(*source.Obj); ok {
			for _, k := range p.Keys() {
				out = append(out, pkg{kindPackage, k})
			}
		}
	}
	if t, ok := m.Get("tools").(*source.Obj); ok {
		for _, k := range t.Keys() {
			out = append(out, pkg{kindTool, k})
		}
	}
	return out
}

// parseLines gives the words of the lines of <out>, each with <prefix>.
func parseLines(out, prefix string) []string {
	var keys []string
	for _, l := range strings.Split(out, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			keys = append(keys, prefix+l)
		}
	}
	return keys
}

// parseAppStore gives the mas: keys of the output of `mas list`. Each line
// starts with the ID of the app.
func parseAppStore(out string) []string {
	var keys []string
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) > 0 && strings.Trim(f[0], "0123456789") == "" {
			keys = append(keys, prefixApp+f[0])
		}
	}
	return keys
}

// parseMiseTools gives the tool names of the output of
// `mise ls --global --installed --json`.
func parseMiseTools(out []byte) ([]string, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(out, &m); err != nil {
		return nil, err
	}
	var names []string
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}

// run runs <name> with <args> in <dir>, and gives its stdout. A failure
// gives an error with the stderr of the command.
func run(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		msg := err.Error()
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		return nil, fmt.Errorf("%s %s failed: %s", name, strings.Join(args, " "), msg)
	}
	return out, nil
}

// installedPackages lists the installed packages of the Mac. A package
// manager that is not installed is skipped, with an info line. <deps> are the
// other installed formulae: a dependency, or a formula of a tap that brew
// does not trust. They match a Mac tool, but they are never an offer.
func (a *app) installedPackages() (out, deps []pkg, code fail.Code) {
	home := a.getenv("HOME")
	add := func(kind string, keys []string) {
		for _, k := range keys {
			out = append(out, pkg{kind, k})
		}
	}
	if _, err := exec.LookPath("brew"); err != nil {
		a.infof("brew is not installed, so Homebrew packages are not scanned")
	} else {
		f, err := run(home, "brew", "leaves", "--installed-on-request")
		if err != nil {
			return nil, nil, a.logErr(err)
		}
		add(kindPackage, parseLines(string(f), prefixFormula))
		all, err := run(home, "brew", "list", "--formula")
		if err != nil {
			return nil, nil, a.logErr(err)
		}
		for _, k := range parseLines(string(all), prefixFormula) {
			deps = append(deps, pkg{kindPackage, k})
		}
		c, err := run(home, "brew", "list", "--cask")
		if err != nil {
			return nil, nil, a.logErr(err)
		}
		add(kindPackage, parseLines(string(c), prefixCask))
	}
	if _, err := exec.LookPath(appStoreCommand); err != nil {
		a.infof("%s is not installed, so App Store apps are not scanned", appStoreCommand)
	} else {
		m, err := run(home, appStoreCommand, "list")
		if err != nil {
			return nil, nil, a.logErr(err)
		}
		add(kindPackage, parseAppStore(string(m)))
	}
	j, err := run(home, "mise", "ls", "--global", "--installed", "--json")
	if err != nil {
		return nil, nil, a.logErr(err)
	}
	names, err := parseMiseTools(j)
	if err != nil {
		return nil, nil, a.logErr(fmt.Errorf("mise ls --global --installed --json did not print a JSON object: %w", err))
	}
	add(kindTool, names)
	return out, deps, fail.OK
}

// macScan is the result of the package scan.
type macScan struct {
	tools   []string            // The Mac tools with a package key, sorted.
	found   map[string]detected // The result of each tool of tools.
	unknown []pkg               // The installed packages of no tool and of no [mise] table of macos.toml.
	owner   map[pkg]string      // The tool of each package key.
}

// scanMac maps the installed packages <inst> and <deps> to the Mac tools of
// <s>. Only <inst> gives the offers of the packages of no tool.
func scanMac(s *side, inst, deps []pkg) macScan {
	r := macScan{found: map[string]detected{}, owner: map[pkg]string{}}
	keys := map[string][]pkg{}
	for t, m := range s.sel.idx.Manifests {
		if ks := packageKeys(m.Raw); len(ks) > 0 {
			keys[t] = ks
			r.tools = append(r.tools, t)
			for _, k := range ks {
				r.owner[k] = t
				r.owner[k.match()] = t
			}
		}
	}
	sort.Strings(r.tools)
	own := map[pkg]bool{}
	for _, k := range packageKeys(s.raw) {
		own[k.match()] = true
	}
	have := map[pkg]bool{}
	for _, p := range append(append([]pkg(nil), inst...), deps...) {
		have[p.match()] = true
	}
	for _, t := range r.tools {
		var ev []string
		for _, k := range keys[t] {
			if have[k.match()] {
				ev = append(ev, k.key)
			}
		}
		if len(ev) == 0 {
			r.found[t] = detected{Result: "none", Message: "no package installed"}
			continue
		}
		sort.Strings(ev)
		r.found[t] = detected{Result: "match", Evidence: ev, Message: "installed"}
	}
	seen := map[pkg]bool{}
	for _, p := range inst {
		m := p.match()
		if _, ok := r.owner[m]; ok || own[m] || seen[m] {
			continue
		}
		seen[m] = true
		r.unknown = append(r.unknown, p)
	}
	sort.Slice(r.unknown, func(i, j int) bool { return r.unknown[i].key < r.unknown[j].key })
	return r
}

// appStoreTool gives the tool that installs the command of the App Store
// apps, or "".
func (r macScan) appStoreTool() string {
	for _, p := range []pkg{{kindPackage, prefixFormula + appStoreCommand}, {kindTool, appStoreCommand}} {
		if t, ok := r.owner[p]; ok {
			return t
		}
	}
	return ""
}
