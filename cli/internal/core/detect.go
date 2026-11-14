package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/scripts"
)

// Detected is the result of the detect script of a tool, or of the package
// scan of a Mac tool.
type Detected struct {
	Result   string   `json:"result"`
	Evidence []string `json:"evidence"`
	Message  string   `json:"message"`
}

// Proposal is a tool that matches and is not selected. A tool that is not
// offered has only the evidence Shared, which is the evidence of the
// selected tool Of.
type Proposal struct {
	Tool       string
	Offered    bool
	Shared, Of string
	Found      Detected
}

// Detection is the result of Detect.
type Detection struct {
	*State
	// NoScripts: no tool of the repo side has a detect script.
	NoScripts bool
	Tools     []string            // The tools with a detect script or a package key, sorted.
	Found     map[string]Detected // The result of each tool of Tools.
	Proposals []Proposal          // The tools that match and are not selected, in the order of Tools.
	Offers    []string            // The offered tools of Proposals.
	// Unknown are the installed packages of no Mac tool and of no own
	// [mise] table of macos.toml: each one is an offer for those tables.
	Unknown []Package
	// Notes are the lines of the package scan, for example a package manager
	// that is not installed.
	Notes []string
	// Mismatch holds a problem for each selected tool whose result is
	// mismatch (code 2).
	Mismatch  error
	storeTool string
}

// Detect reads the values file of <sd>, runs the detect scripts of the tools
// (on the Mac, the package scan), and gives the proposals. It writes nothing.
func Detect(sd Side) (*Detection, error) {
	st, err := Open(sd)
	if err != nil {
		return nil, err
	}
	return DetectState(st)
}

// DetectState is Detect on the state <st>, which the caller has read. The
// TUI calls it with its state, also the state of init before the values
// file is there (InitState). It writes nothing.
func DetectState(st *State) (*Detection, error) {
	sd := st.Side
	d := &Detection{State: st}
	if sd.Mac {
		inst, deps, notes, err := installedPackages(sd.Home())
		d.Notes = notes
		if err != nil {
			return d, err
		}
		scan := scanMac(st, inst, deps)
		d.Tools, d.Found, d.Unknown, d.storeTool = scan.tools, scan.found, scan.unknown, scan.appStoreTool()
	} else {
		if err := d.detectScripts(); err != nil || d.NoScripts {
			return d, err
		}
	}

	// The evidence of the selected tools that match.
	selEvidence := map[string]string{}
	for _, t := range d.Tools {
		if f := d.Found[t]; f.Result == "match" && st.Selected(t) {
			for _, e := range f.Evidence {
				selEvidence[e] = t
			}
		}
	}
	var mismatch []error
	for _, t := range d.Tools {
		f := d.Found[t]
		if st.Selected(t) {
			if f.Result == "mismatch" {
				mismatch = append(mismatch, problem(fail.Usage, fmt.Sprintf("%s is selected, but its detect script gives mismatch: %s", t, f.Message), ""))
			}
			continue
		}
		if f.Result != "match" || !Proposable(st.Manifest(t)) {
			continue
		}
		// Two tools with the same evidence: the selected one is enough.
		same, shared := "", ""
		for _, e := range f.Evidence {
			if u, ok := selEvidence[e]; ok {
				same, shared = u, e
			}
		}
		if same != "" {
			d.Proposals = append(d.Proposals, Proposal{Tool: t, Shared: shared, Of: same, Found: f})
			continue
		}
		d.Offers = append(d.Offers, t)
		d.Proposals = append(d.Proposals, Proposal{Tool: t, Offered: true, Found: f})
	}
	d.Mismatch = join(mismatch)
	return d, nil
}

// detectScripts runs the detect script of each tool that has one, and keeps
// the tools and their results.
func (d *Detection) detectScripts() error {
	r := d.Runner()
	for t := range d.Sel.Idx.Manifests {
		if r.Has(t, scripts.Detect) {
			d.Tools = append(d.Tools, t)
		}
	}
	sort.Strings(d.Tools)
	if len(d.Tools) == 0 {
		d.NoScripts = true
		return nil
	}
	var jobs []scripts.Job
	for _, t := range d.Tools {
		jobs = append(jobs, scripts.Job{Tool: t, Name: scripts.Detect, Input: d.Input(t, t)})
	}
	d.Found = map[string]Detected{}
	var errs []error
	for _, res := range r.Run(jobs) {
		shown := r.ShownPath(res.Tool, res.Name)
		var f Detected
		if res.Code != 0 {
			for _, l := range strings.Split(strings.TrimSpace(string(res.Stderr)), "\n") {
				if l != "" {
					errs = append(errs, problem(fail.Usage, shown+": "+l, ""))
				}
			}
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s failed with exit code %d", shown, res.Code), ""))
			continue
		}
		if err := json.Unmarshal(res.Stdout, &f); err != nil {
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s did not print a JSON object: %s", shown, err), ""))
			continue
		}
		switch f.Result {
		case "match", "none", "mismatch":
		default:
			errs = append(errs, problem(fail.Usage, fmt.Sprintf("%s printed the result '%s'", shown, f.Result), "the results are match, none and mismatch"))
			continue
		}
		d.Found[res.Tool] = f
	}
	return join(errs)
}

// Items gives the items of the change list that add the tools <tools> and
// the own packages <pkgs>. An App Store app also adds the tool of the
// command of the App Store apps, when it is not selected: the own [mise]
// tables have no depends.
func (d *Detection) Items(tools []string, pkgs []Package) []Item {
	if t := d.storeTool; t != "" && !d.Selected(t) && !slices.Contains(tools, t) {
		for _, p := range pkgs {
			if p.IsApp() {
				tools = append(append([]string(nil), tools...), t)
				break
			}
		}
	}
	var items []Item
	if len(tools) > 0 {
		items = append(items, Item{Kind: AddTool, Names: tools})
	}
	for _, p := range pkgs {
		items = append(items, Item{Kind: AddPackage, Package: p})
	}
	return items
}
