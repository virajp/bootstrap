// Package scripts runs the scripts of a tool: the files `derive`, `check`
// and `after` in `<side>/tools/<tool>/scripts/` of the clone. The engine
// never copies them. Each one runs as `mise x -- bash <path>` in the target
// root, with one JSON object on stdin. The jobs run in parallel, and the
// results come back in the order of the jobs.
// Contract: docs/design/sync-engine.md ("The scripts of a tool").
package scripts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// The names of the scripts.
const (
	Derive = "derive"
	Check  = "check"
	After  = "after"
)

// Dir is the folder of the scripts in a tool folder.
const Dir = source.ScriptsDir

// Member is one member of a repository, as a script gets it.
type Member struct {
	Name   string `json:"name"`
	Folder string `json:"folder"`
	Type   string `json:"type"`
}

// Repo is the `repo` object of the input on the repo side.
type Repo struct {
	Root    string   `json:"root"`
	Name    string   `json:"name"`
	Members []Member `json:"members"`
}

// Input is the JSON object on stdin. Repo is for the repo side, Home for the
// Mac side.
type Input struct {
	Tool   string
	Values *source.Obj
	Repo   *Repo
	Home   string
}

// JSON gives the text of the input.
func (in Input) JSON() []byte {
	var b strings.Builder
	tool, _ := json.Marshal(in.Tool)
	b.WriteString(`{"tool":` + string(tool) + `,"values":`)
	if in.Values == nil {
		b.WriteString("{}")
	} else {
		b.WriteString(source.JSON(in.Values))
	}
	if in.Repo != nil {
		if in.Repo.Members == nil {
			in.Repo.Members = []Member{}
		}
		r, _ := json.Marshal(in.Repo)
		b.WriteString(`,"repo":` + string(r))
	} else {
		h, _ := json.Marshal(in.Home)
		b.WriteString(`,"home":` + string(h))
	}
	b.WriteString("}")
	return []byte(b.String())
}

// Job is one script to run.
type Job struct {
	Tool  string
	Name  string
	Input []byte
}

// Result is the result of one job. Ran is false when the tool has no such
// script.
type Result struct {
	Tool   string
	Name   string
	Ran    bool
	Code   int // The exit code; -1 when the script could not start.
	Stdout []byte
	Stderr []byte
}

// Runner holds what each job of a run shares.
type Runner struct {
	ToolsDir string // The absolute folder of the tools of the side.
	Shown    string // That folder in a message, for example repo/tools.
	Root     string // The folder where each script runs.
	Jobs     int    // The number of jobs at a time; 0 is BOOTSTRAP_JOBS, else 2 × the CPU count.
}

// Path gives the absolute path of the script <name> of <tool>.
func (r Runner) Path(tool, name string) string {
	return filepath.Join(r.ToolsDir, tool, Dir, name)
}

// ShownPath gives the script as a message names it.
func (r Runner) ShownPath(tool, name string) string {
	return r.Shown + "/" + tool + "/" + Dir + "/" + name
}

// Has tells if <tool> has the script <name>.
func (r Runner) Has(tool, name string) bool {
	st, err := os.Stat(r.Path(tool, name))
	return err == nil && st.Mode().IsRegular()
}

// Run runs the jobs, at most Jobs at a time, and gives their results in the
// order of the jobs.
func (r Runner) Run(jobs []Job) []Result {
	out := make([]Result, len(jobs))
	limit := r.Jobs
	if limit <= 0 {
		// --jobs of the command sets BOOTSTRAP_JOBS.
		limit, _ = strconv.Atoi(os.Getenv("BOOTSTRAP_JOBS"))
	}
	if limit <= 0 {
		limit = 2 * runtime.NumCPU()
	}
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i, j := range jobs {
		out[i] = Result{Tool: j.Tool, Name: j.Name}
		if !r.Has(j.Tool, j.Name) {
			continue
		}
		wg.Add(1)
		go func(i int, j Job) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = r.one(j)
		}(i, j)
	}
	wg.Wait()
	return out
}

func (r Runner) one(j Job) Result {
	res := Result{Tool: j.Tool, Name: j.Name, Ran: true}
	cmd := exec.Command("mise", "x", "--", "bash", r.Path(j.Tool, j.Name))
	cmd.Dir = r.Root
	cmd.Stdin = bytes.NewReader(j.Input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res.Stdout, res.Stderr = stdout.Bytes(), stderr.Bytes()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.Code = ee.ExitCode()
		} else {
			res.Code = -1
			res.Stderr = append(res.Stderr, []byte(err.Error()+"\n")...)
		}
	}
	return res
}

// lines gives the lines of <b> that are not empty.
func lines(b []byte) []string {
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// failed gives the message of a script that failed in a way that its kind
// does not allow, with code 2, and the stderr lines of the script.
func (r Runner) failed(res Result, what string) []*fail.Error {
	var probs []*fail.Error
	for _, l := range lines(res.Stderr) {
		probs = append(probs, fail.New(fail.Usage, "%s: %s", r.ShownPath(res.Tool, res.Name), l))
	}
	return append(probs, fail.New(fail.Usage, "%s %s", r.ShownPath(res.Tool, res.Name), what))
}

// Derived reads the result of a derive script: a JSON object. It gives nil
// and no problem when the script did not run.
func (r Runner) Derived(res Result) (*source.Obj, []*fail.Error) {
	if !res.Ran {
		return nil, nil
	}
	if res.Code != 0 {
		return nil, r.failed(res, fmt.Sprintf("failed with exit code %d", res.Code))
	}
	o, err := FromJSON(res.Stdout)
	if err != nil {
		return nil, r.failed(res, "did not print a JSON object: "+err.Error())
	}
	return o, nil
}

// Checked reads the result of a check script: exit 0, or the lines of
// stdout as messages and exit 2. Each message has code 2.
func (r Runner) Checked(res Result) []*fail.Error {
	if !res.Ran || res.Code == 0 {
		return nil
	}
	if res.Code == 2 {
		var probs []*fail.Error
		for _, l := range lines(res.Stdout) {
			probs = append(probs, fail.New(fail.Usage, "%s", l))
		}
		if len(probs) == 0 {
			probs = append(probs, fail.New(fail.Usage, "%s found a problem and printed no message", r.ShownPath(res.Tool, res.Name)))
		}
		return probs
	}
	return r.failed(res, fmt.Sprintf("failed with exit code %d", res.Code))
}

// Restart reads the result of an after script: `{ "restart": "<reason>" }`
// asks for a new terminal. It gives the reason, or empty.
func (r Runner) Restart(res Result) (string, []*fail.Error) {
	if !res.Ran {
		return "", nil
	}
	if res.Code != 0 {
		return "", r.failed(res, fmt.Sprintf("failed with exit code %d", res.Code))
	}
	if strings.TrimSpace(string(res.Stdout)) == "" {
		return "", nil
	}
	o, err := FromJSON(res.Stdout)
	if err != nil {
		return "", r.failed(res, "did not print a JSON object: "+err.Error())
	}
	v, has := o.Lookup("restart")
	if !has || v == nil {
		return "", nil
	}
	s, ok := v.(string)
	if !ok {
		return "", r.failed(res, "printed a restart that is not a string")
	}
	return s, nil
}

// FromJSON decodes a JSON object into an object with sorted keys. A whole
// number becomes an int64.
func FromJSON(b []byte) (*source.Obj, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("more than one JSON value")
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("not an object")
	}
	o, _ := source.FromMap(numbers(m)).(*source.Obj)
	return o, nil
}

func numbers(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			x[k] = numbers(e)
		}
		return x
	case []any:
		for i, e := range x {
			x[i] = numbers(e)
		}
		return x
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return i
		}
		f, _ := x.Float64()
		return f
	}
	return v
}
