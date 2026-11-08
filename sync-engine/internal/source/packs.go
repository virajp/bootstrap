package source

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Member is one key of a `[pack.<name>]` table. State is `true` or `false` (the
// value of `default`), or `bad:key <key>` or `bad:default <kind>`.
type Member struct {
	Pack, Key, State string
}

// BadValue is a key of a pack table whose value is not a table. Kind is the
// kind of the value: `bool`, `string`, `slice`, `int64`, `float64`.
type BadValue struct {
	Pack, Key, Kind string
}

// Order is a member that the line scan found on one line, in file order.
type Order struct {
	Pack, Key string
}

// Packs is the shape of a packs file, as `_se_selected` of the bash engine
// reads it.
type Packs struct {
	Shown   string
	Names   []string   // The packs, sorted.
	Members []Member   // The table members, sorted by pack and key.
	Bad     []BadValue // The keys that are not tables, sorted.
	Order   []Order    // The one-line members, in file order.
	// DefaultTrue holds `<pack>/<key>` for each member table whose `default` is
	// the boolean true, whatever its other keys are.
	DefaultTrue map[string]bool
}

// IsDefaultTrue tells if the member table has `default = true`.
func (p *Packs) IsDefaultTrue(pack, key string) bool { return p.DefaultTrue[pack+"/"+key] }

// HasPack tells if the packs file has the pack <name>.
func (p *Packs) HasPack(name string) bool {
	for _, n := range p.Names {
		if n == name {
			return true
		}
	}
	return false
}

// HasMember tells if <pack> has the member <key> with the state <state>.
func (p *Packs) HasMember(pack, key, state string) bool {
	for _, m := range p.Members {
		if m.Pack == pack && m.Key == key && m.State == state {
			return true
		}
	}
	return false
}

// HasOrder tells if the line scan found the member on one line.
func (p *Packs) HasOrder(pack, key string) bool {
	for _, o := range p.Order {
		if o.Pack == pack && o.Key == key {
			return true
		}
	}
	return false
}

// kindOf gives the kind name that gomplate `test.Kind` gives.
func kindOf(v any) string {
	switch v.(type) {
	case bool:
		return "bool"
	case string:
		return "string"
	case int64:
		return "int64"
	case float64:
		return "float64"
	case []any:
		return "slice"
	case *Obj:
		return "map"
	case time.Time:
		return "struct"
	}
	return "invalid"
}

// ReadPacks reads the packs file. A missing file gives no pack. It returns the
// message `cannot read <shown>` (code 2) when the file does not parse.
func (s Source) ReadPacks() (*Packs, error) {
	p := &Packs{Shown: s.PacksShown(), DefaultTrue: map[string]bool{}}
	file := s.PacksFile()
	st, err := os.Stat(file)
	if err != nil || !st.Mode().IsRegular() {
		return p, nil
	}
	bad := readErr("cannot read %s", p.Shown)
	root, err := readTOML(file)
	if err != nil {
		return nil, bad
	}
	if root.Has("pack") {
		packs, ok := root.Get("pack").(*Obj)
		if !ok {
			return nil, bad
		}
		for _, name := range packs.Keys() {
			p.Names = append(p.Names, name)
			table, ok := packs.Get(name).(*Obj)
			if !ok {
				return nil, bad
			}
			for _, key := range table.Keys() {
				v := table.Get(key)
				e, ok := v.(*Obj)
				if !ok {
					p.Bad = append(p.Bad, BadValue{name, key, kindOf(v)})
					continue
				}
				state := "false"
				if e.Has("default") {
					d := e.Get("default")
					if b, ok := d.(bool); ok {
						p.DefaultTrue[name+"/"+key] = b
						state = map[bool]string{true: "true", false: "false"}[b]
					} else {
						state = "bad:default " + kindOf(d)
					}
				}
				for _, ek := range e.Keys() {
					if ek != "default" {
						state = "bad:key " + ek
					}
				}
				p.Members = append(p.Members, Member{name, key, state})
			}
		}
	}
	order, err := scanOrder(file)
	if err != nil {
		return nil, bad
	}
	p.Order = order
	sort.Strings(p.Names)
	return p, nil
}

var (
	reTableLine = regexp.MustCompile(`^[ \t]*\[`)
	rePackHead  = regexp.MustCompile(`^[ \t]*\[pack\.[^].]+\][ \t]*(#.*)?$`)
	reMemberRow = regexp.MustCompile(`^[ \t]*[^#= \t][^=]*=[ \t]*\{.*\}[ \t]*(#.*)?$`)
)

// bare removes each quote, as the awk function `bare` does.
func bare(s string) string {
	return strings.NewReplacer(`"`, "", `'`, "").Replace(s)
}

// scanOrder is the line scan of the bash engine: the members of the
// `[pack.<name>]` tables that are on one line, in file order.
func scanOrder(file string) ([]Order, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Order
	name := ""
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if reTableLine.MatchString(line) {
			name = ""
			if rePackHead.MatchString(line) {
				n := strings.TrimLeft(line, " \t")
				n = strings.TrimPrefix(n, "[pack.")
				n = n[:strings.Index(n, "]")]
				name = bare(n)
			}
			continue
		}
		if name != "" && reMemberRow.MatchString(line) {
			key := strings.TrimLeft(line, " \t")
			key = strings.TrimRight(key[:strings.Index(key, "=")], " \t")
			out = append(out, Order{name, bare(key)})
		}
	}
	return out, sc.Err()
}
