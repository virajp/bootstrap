package source

import (
	"regexp"
	"sort"
	"strings"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// eachVarRe is one {name} part of an `each` path or of a target path.
var eachVarRe = regexp.MustCompile(`\{([a-z][a-z0-9_]*)\}`)

// eachKeyRe is a key of a table that an `each` path walks: it becomes a part
// of a target path.
var eachKeyRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// EachValue is the name of the table of one expansion in its Vars.
const EachValue = "value"

// EachVars gives the names of the {name} parts of <s>, in order.
func EachVars(s string) []string {
	var out []string
	for _, m := range eachVarRe.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

func sameVars(a, b []string) bool {
	x := append([]string(nil), a...)
	y := append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	return strings.Join(x, " ") == strings.Join(y, " ")
}

// EachOK tells if <w> is an `each` path: words joined with dots, the first a
// word, at least one part a whole {name}, and no name twice or `value`.
func EachOK(w string) bool {
	parts := strings.Split(w, ".")
	seen := map[string]bool{}
	vars := 0
	for i, p := range parts {
		if m := eachVarRe.FindStringSubmatch(p); m != nil {
			if m[0] != p || i == 0 || seen[m[1]] || m[1] == EachValue {
				return false
			}
			seen[m[1]] = true
			vars++
			continue
		}
		if p == "" || strings.ContainsAny(p, " \t\n!{}") {
			return false
		}
	}
	return vars > 0
}

// Expand gives one File for each table that the `each` path of <f> finds in
// <values>, the values of repo.toml, in key order. Each File has the target
// with its {name} parts replaced, and Vars. A key that cannot be a part of a
// path gives an error.
func Expand(values any, f File) ([]File, error) {
	var out []File
	var walk func(v any, parts []string, vars map[string]any) error
	walk = func(v any, parts []string, vars map[string]any) error {
		if len(parts) == 0 {
			one := f
			one.Vars = map[string]any{EachValue: v}
			target := f.Target
			for k, x := range vars {
				one.Vars[k] = x
				target = strings.ReplaceAll(target, "{"+k+"}", x.(string))
			}
			one.Target = target
			out = append(out, one)
			return nil
		}
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		p := parts[0]
		if mm := eachVarRe.FindStringSubmatch(p); mm != nil {
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if !eachKeyRe.MatchString(k) {
					return fail.New(fail.Usage, "%s: the key '%s' of %s cannot be a part of a path", f.Path, k, f.Each)
				}
				next := map[string]any{mm[1]: k}
				for a, b := range vars {
					next[a] = b
				}
				if err := walk(m[k], parts[1:], next); err != nil {
					return err
				}
			}
			return nil
		}
		return walk(m[p], parts[1:], vars)
	}
	err := walk(values, strings.Split(f.Each, "."), map[string]any{})
	return out, err
}
