package tmpl

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// The folders testdata/files/inline-*/i.* hold the inline templates that the
// bash engine rendered (gomplate -i): the template, the files of the -c flags,
// the environment variables of the engine, the output and the error. See
// files_test.go.

// loadCtxFile reads one -c file. A TOML file gives int64 numbers, as the
// gomplate TOML datasource does.
func loadCtxFile(path string) (any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	switch filepath.Ext(path) {
	case ".toml":
		var m map[string]any
		if _, err := toml.Decode(string(b), &m); err != nil {
			return nil, err
		}
		return fixTOML(m), nil
	default:
		return DecodeJSON(b)
	}
}

// fixTOML changes the types of the TOML decoder to the types of gomplate: an
// array of tables is a []any of maps.
func fixTOML(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			x[k] = fixTOML(e)
		}
		return x
	case []map[string]any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = fixTOML(e)
		}
		return out
	case []any:
		for i, e := range x {
			x[i] = fixTOML(e)
		}
		return x
	}
	return v
}

// TestInlineTemplates renders each captured inline template and compares the
// output (or the error message) with the gomplate CLI.
func TestInlineTemplates(t *testing.T) {
	all, err := loadRenders()
	if err != nil {
		t.Fatal(err)
	}
	var cases []render
	for _, r := range all {
		if strings.HasPrefix(path.Base(r.name), "i.") {
			cases = append(cases, r)
		}
	}
	if len(cases) == 0 {
		t.Fatal("no captured inline renders")
	}
	tmp := scratch(t)
	sub := func(s string) string { return strings.ReplaceAll(s, "@TMP@", tmp) }
	templates := map[string]bool{}
	for _, r := range cases {
		ctx := map[string]any{}
		var root any
		var names []string
		for n := range r.files {
			if strings.HasPrefix(n, "ctx/") {
				names = append(names, n)
			}
		}
		sort.Strings(names)
		for _, n := range names {
			cp := filepath.Join(t.TempDir(), path.Base(n))
			if err := os.WriteFile(cp, []byte(sub(r.files[n])), 0o644); err != nil {
				t.Fatal(err)
			}
			v, err := loadCtxFile(cp)
			if err != nil {
				t.Fatalf("%s/%s: %v", r.name, n, err)
			}
			base := path.Base(n)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if name == "." {
				root = v
				continue
			}
			ctx[name] = v
		}
		var data any = ctx
		if root != nil {
			data = root
		}
		var env map[string]string
		if err := json.Unmarshal([]byte(sub(r.get("env.json"))), &env); err != nil {
			t.Fatal(err)
		}
		inline := sub(r.get("inline"))
		got, err := Render("<arg>", []byte(inline), data, &Options{
			MissingKey: r.get("missingkey"),
			Env:        func(k string) string { return env[k] },
		})
		if r.get("rc") != "0" {
			want := cliMessage(sub(r.get("err")))
			if err == nil {
				t.Errorf("%s: gomplate failed (%q), Go did not\n%s", r.name, want, inline)
			} else if m := goMessage(err); m != want {
				t.Errorf("%s: want message %q, got %q (%v)", r.name, want, m, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v\n%s", r.name, err, inline)
			continue
		}
		if want := sub(r.get("out")); string(got) != want {
			t.Errorf("%s: output differs\n--- want\n%q\n--- got\n%q\n%s", r.name, want, got, inline)
			continue
		}
		templates[inline] = true
	}
	t.Logf("%d inline renders byte-equal, of %d different templates", len(cases), len(templates))
}
