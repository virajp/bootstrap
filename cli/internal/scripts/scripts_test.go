package scripts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

// writeScript writes the script <name> of <tool> below <dir>.
func writeScript(t *testing.T, dir, tool, name, body string) {
	t.Helper()
	p := filepath.Join(dir, tool, "scripts", name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestInputJSON(t *testing.T) {
	v := source.FromMap(map[string]any{"b": 1, "a": "x"}).(*source.Obj)
	got := string(Input{Tool: "git", Values: v, Repo: &Repo{Root: "/r", Name: "n"}}.JSON())
	want := `{"tool":"git","values":{"a":"x","b":1},"repo":{"root":"/r","name":"n","members":[]}}`
	if got != want {
		t.Errorf("repo input = %s, want %s", got, want)
	}
	got = string(Input{Tool: "fish", Home: "/h"}.JSON())
	if want = `{"tool":"fish","values":{},"home":"/h"}`; got != want {
		t.Errorf("mac input = %s, want %s", got, want)
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	root := t.TempDir()
	writeScript(t, dir, "a", Derive, "in=$(cat); printf '{\"tool\": %s, \"pwd\": \"%s\", \"n\": 2}' \"$(printf '%s' \"$in\" | sed 's/.*\"tool\":\\(\"[a-z]*\"\\).*/\\1/')\" \"$PWD\"\n")
	writeScript(t, dir, "b", Derive, "cat > /dev/null; echo 'not json'\n")
	writeScript(t, dir, "c", Check, "cat > /dev/null; echo 'first problem'; echo; echo 'second problem'; exit 2\n")
	writeScript(t, dir, "d", Check, "cat > /dev/null; echo 'oops' >&2; exit 3\n")
	writeScript(t, dir, "e", After, "cat > /dev/null; echo '{\"restart\": \"a new shell\"}'\n")
	writeScript(t, dir, "f", After, "cat > /dev/null\n")
	r := Runner{ToolsDir: dir, Shown: "repo/tools", Root: root}
	res := r.Run([]Job{
		{Tool: "a", Name: Derive, Input: []byte(`{"tool":"a"}`)},
		{Tool: "b", Name: Derive},
		{Tool: "z", Name: Derive},
		{Tool: "c", Name: Check},
		{Tool: "d", Name: Check},
		{Tool: "e", Name: After},
		{Tool: "f", Name: After},
	})
	if len(res) != 7 || res[0].Tool != "a" || res[2].Ran {
		t.Fatalf("results = %+v", res)
	}
	o, probs := r.Derived(res[0])
	realRoot, _ := filepath.EvalSymlinks(root)
	if probs != nil || o.Get("tool") != "a" || o.Get("n") != int64(2) || (o.Get("pwd") != root && o.Get("pwd") != realRoot) {
		t.Errorf("derive a = %s, %v", source.JSON(o), probs)
	}
	if _, probs := r.Derived(res[1]); len(probs) != 1 || !strings.Contains(probs[0].Error(), "repo/tools/b/scripts/derive did not print a JSON object") {
		t.Errorf("derive b = %v", probs)
	}
	if o, probs := r.Derived(res[2]); o != nil || probs != nil {
		t.Errorf("no script = %v, %v", o, probs)
	}
	if probs := r.Checked(res[3]); len(probs) != 2 || probs[0].Error() != "check script of c: first problem" || probs[1].Error() != "check script of c: second problem" {
		t.Errorf("check c = %v", probs)
	}
	if probs := r.Checked(res[4]); len(probs) != 2 || probs[0].Error() != "repo/tools/d/scripts/check: oops" ||
		probs[1].Error() != "repo/tools/d/scripts/check failed with exit code 3" {
		t.Errorf("check d = %v", probs)
	}
	if reason, probs := r.Restart(res[5]); reason != "a new shell" || probs != nil {
		t.Errorf("after e = %q, %v", reason, probs)
	}
	if reason, probs := r.Restart(res[6]); reason != "" || probs != nil {
		t.Errorf("after f = %q, %v", reason, probs)
	}
}
