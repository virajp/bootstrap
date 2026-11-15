package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// TaskNames gives the names of the mise tasks of the repository <root>,
// sorted: `mise tasks ls --json --no-global --hidden`, so a global task is
// left out and a hidden task is in (D19, D27). A failure of mise, for
// example an untrusted config, is code 1.
func TaskNames(root string) ([]string, error) {
	out, err := run(root, "mise", "tasks", "ls", "--json", "--no-global", "--hidden")
	if err != nil {
		return nil, problem(fail.Other, err.Error(), "run mise trust in "+root+" when mise does not trust its config")
	}
	var tasks []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, problem(fail.Other, fmt.Sprintf("mise tasks ls --json did not print a JSON list: %v", err), "")
	}
	names := make([]string, 0, len(tasks))
	for _, t := range tasks {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names, nil
}

// PortMatch tells if the task <task> is the port <port>. A segment <...> of
// the port, for example <branch>, matches one segment of the task without a
// colon.
func PortMatch(port, task string) bool {
	ps, ts := strings.Split(port, ":"), strings.Split(task, ":")
	if len(ps) != len(ts) {
		return false
	}
	for i, p := range ps {
		if strings.HasPrefix(p, "<") && strings.HasSuffix(p, ">") {
			if ts[i] == "" {
				return false
			}
			continue
		}
		if p != ts[i] {
			return false
		}
	}
	return true
}
