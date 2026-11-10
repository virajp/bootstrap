package source

import "strings"

// When tells if an entry with the `when` value path <path> applies: the value
// at that path in <all>, the values of all the tools by tool name, is true. A
// ! at the start negates it. An empty path always applies.
func When(all *Obj, path string) bool {
	if path == "" {
		return true
	}
	p, neg := strings.CutPrefix(path, "!")
	var v any = all
	for _, k := range strings.Split(p, ".") {
		o, ok := v.(*Obj)
		if !ok || o == nil {
			v = nil
			break
		}
		v = o.Get(k)
	}
	b, _ := v.(bool)
	return b != neg
}
