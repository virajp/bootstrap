package format

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// This file holds the merge spec of a shared file: the rendered
// `<name>.merge.toml` of a `[files]` entry with `mode = "merge"`. Contract:
// docs/design/cli.md ("Shared files").

// JSONC is JSON with comments and trailing commas. Only a shared file has
// this format: the parse check does not examine it.
const JSONC Format = "jsonc"

// The kinds of an item.
const (
	kindKey   = "key"   // One leaf key. A list value is one value.
	kindKeys  = "keys"  // A table of leaf keys: one key item for each leaf.
	kindItem  = "item"  // One item of a list of scalars, or one line of a text file.
	kindBlock = "block" // A value that bootstrap owns.
)

// specItem is one `[[item]]` of a spec, or one leaf of a `keys` item.
type specItem struct {
	kind   string
	path   []string
	value  Value
	hasVal bool
	match  string         // The regex of an item.
	re     *regexp.Regexp // The compiled match of an item.
	keys   []string       // The match keys of a block in a table[].
	id     string
	ignore []string
	remove bool // "$mode" = "remove".
	skip   bool // "$skip" = true.
}

// Spec is a rendered merge spec.
type Spec struct {
	Format  Format // The `format` key, or "".
	Comment string // The `comment` key of a text file, or "".
	items   []*specItem
}

var specTop = []string{"format", "comment", "item"}
var specKeys = []string{"kind", "path", "value", "match", "id", "ignore", "type", "$mode", "$skip"}
var itemTypes = []string{"string", "boolean", "integer", "float", "string[]", "boolean[]", "integer[]", "float[]", "table", "table[]"}

func oneOfStr(s string, list []string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// tomlToJSON reads a TOML text with yq and gives its value, with the key
// order of the text.
func tomlToJSON(text []byte) (Value, error) {
	tmp, err := os.CreateTemp("", "bootstrap-spec-*.toml")
	if err != nil {
		return Value{}, err
	}
	defer os.Remove(tmp.Name())
	_, werr := tmp.Write(text)
	if cerr := tmp.Close(); werr != nil || cerr != nil {
		return Value{}, fmt.Errorf("cannot write a temp file")
	}
	r, yerr := yq(nil, "eval", "-p", "toml", "-o", "json", ".", tmp.Name())
	if yerr != nil {
		return Value{}, fmt.Errorf("%s", strings.TrimSpace(strings.ReplaceAll(r.stderr, tmp.Name(), "the spec")))
	}
	v, ok := oneValue([]byte(r.stdout))
	if !ok {
		return Value{}, fmt.Errorf("the spec is not one TOML table")
	}
	if v.k == kNull {
		v = objv(newObject())
	}
	return v, nil
}

// ParseSpec reads the rendered spec <text>. <show> names the spec in each
// message.
func ParseSpec(text []byte, show string) (*Spec, error) {
	v, err := tomlToJSON(text)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", show, err)
	}
	bad := func(format string, args ...any) error {
		return fmt.Errorf("%s: "+format, append([]any{show}, args...)...)
	}
	sp := &Spec{}
	for _, k := range v.o.keys {
		if !oneOfStr(k, specTop) {
			return nil, bad("%s is not a key of a merge spec - the keys are format, comment and [[item]]", k)
		}
	}
	if f := v.get("format"); f.k != kNull {
		if f.k != kString || !oneOfStr(f.s, []string{"text", "json", "jsonc", "yaml", "toml"}) {
			return nil, bad("format must be \"text\", \"json\", \"jsonc\", \"yaml\" or \"toml\"")
		}
		sp.Format = Format(f.s)
	}
	if c := v.get("comment"); c.k != kNull {
		if c.k != kString || c.s == "" || strings.ContainsAny(c.s, "\n") {
			return nil, bad("comment must be the comment marker of a line, for example \"#\"")
		}
		sp.Comment = c.s
	}
	items := v.get("item")
	if items.k != kNull && items.k != kArray {
		return nil, bad("item must be a list of [[item]] tables")
	}
	for i, it := range items.a {
		si, err := parseItem(it, sp.Format == Text)
		if err != nil {
			return nil, bad("item %d: %v", i+1, err)
		}
		sp.items = append(sp.items, si)
	}
	return sp, nil
}

func strings1(v Value) ([]string, bool) {
	if v.k != kArray {
		return nil, false
	}
	out := make([]string, 0, len(v.a))
	for _, e := range v.a {
		if e.k != kString {
			return nil, false
		}
		out = append(out, e.s)
	}
	return out, true
}

func parseItem(it Value, text bool) (*specItem, error) {
	if it.k != kObject {
		return nil, fmt.Errorf("an item must be a table")
	}
	for _, k := range it.o.keys {
		if !oneOfStr(k, specKeys) {
			return nil, fmt.Errorf("%s is not a key of an item - the keys are %s", k, strings.Join(specKeys, ", "))
		}
	}
	si := &specItem{}
	if k := it.get("kind"); k.k == kString {
		si.kind = k.s
	} else if k.k != kNull {
		return nil, fmt.Errorf("kind must be a string")
	} else if text {
		si.kind = kindItem
	}
	if !oneOfStr(si.kind, []string{kindKey, kindKeys, kindItem, kindBlock}) {
		return nil, fmt.Errorf("kind must be \"key\", \"keys\", \"item\" or \"block\"")
	}
	if text && si.kind != kindItem {
		return nil, fmt.Errorf("a text file has items of kind \"item\" only")
	}
	if p := it.get("path"); p.k != kNull {
		l, ok := strings1(p)
		if !ok {
			return nil, fmt.Errorf("path must be a list of strings")
		}
		si.path = l
	}
	if text && len(si.path) > 0 {
		return nil, fmt.Errorf("an item of a text file has no path")
	}
	if !text && len(si.path) == 0 && si.kind != kindKeys {
		return nil, fmt.Errorf("a %s item needs a path", si.kind)
	}
	if m := it.get("$mode"); m.k != kNull {
		if m.k != kString || m.s != "remove" {
			return nil, fmt.Errorf("$mode must be \"remove\"")
		}
		si.remove = true
	}
	if s := it.get("$skip"); s.k != kNull {
		if s.k != kBool {
			return nil, fmt.Errorf("$skip must be true or false")
		}
		si.skip = s.b
	}
	if it.has("value") {
		si.value, si.hasVal = it.get("value"), true
	} else if !si.remove {
		return nil, fmt.Errorf("an item needs a value, or \"$mode\" = \"remove\"")
	}
	switch si.kind {
	case kindKeys:
		if si.hasVal && si.value.k != kObject {
			return nil, fmt.Errorf("the value of a keys item must be a table")
		}
	case kindItem:
		if si.hasVal && (si.value.k == kObject || si.value.k == kArray || si.value.k == kNull) {
			return nil, fmt.Errorf("the value of an item must be a string, a number or a boolean")
		}
		if text && si.hasVal && (si.value.k != kString || strings.ContainsAny(si.value.s, "\r\n")) {
			return nil, fmt.Errorf("the value of a line must be a string of one line")
		}
	}
	if m := it.get("match"); m.k != kNull {
		switch {
		case si.kind == kindItem && m.k == kString:
			re, err := regexp.Compile(m.s)
			if err != nil {
				return nil, fmt.Errorf("match is not a valid regex: %v", err)
			}
			si.match, si.re = m.s, re
		case si.kind == kindBlock:
			l, ok := strings1(m)
			if !ok || len(l) == 0 {
				return nil, fmt.Errorf("match of a block must be a non-empty list of keys")
			}
			si.keys = l
		default:
			return nil, fmt.Errorf("match is for an item (a regex) or a block (a list of keys) only")
		}
	}
	if id := it.get("id"); id.k != kNull {
		if id.k != kString || id.s == "" || si.kind != kindBlock {
			return nil, fmt.Errorf("id must be a string, and only a block has one")
		}
		si.id = id.s
	}
	if si.keys != nil && si.id == "" {
		return nil, fmt.Errorf("a block with match needs an id")
	}
	if si.keys != nil && si.hasVal && si.value.k != kObject {
		return nil, fmt.Errorf("the value of a block with match must be a table")
	}
	if ig := it.get("ignore"); ig.k != kNull {
		l, ok := strings1(ig)
		if !ok || si.kind != kindBlock {
			return nil, fmt.Errorf("ignore must be a list of keys, and only a block has one")
		}
		si.ignore = l
	}
	if t := it.get("type"); t.k != kNull {
		if t.k != kString || !oneOfStr(t.s, itemTypes) {
			return nil, fmt.Errorf("type must be one of %s", strings.Join(itemTypes, ", "))
		}
		if si.hasVal && !typeOK(si.value, t.s) {
			return nil, fmt.Errorf("the value is not of the type %s", t.s)
		}
	}
	return si, nil
}

// typeOK tells if <v> has the type <t> of an item.
func typeOK(v Value, t string) bool {
	if elem, ok := strings.CutSuffix(t, "[]"); ok {
		if v.k != kArray {
			return false
		}
		for _, e := range v.a {
			if !typeOK(e, elem) {
				return false
			}
		}
		return true
	}
	switch t {
	case "string":
		return v.k == kString
	case "boolean":
		return v.k == kBool
	case "integer":
		return v.k == kNumber && !strings.ContainsAny(v.n.String(), ".eE")
	case "float":
		return v.k == kNumber
	case "table":
		return v.k == kObject
	}
	return false
}

// expand gives the items of the spec, with each keys item as one key item for
// each leaf. A table with no keys is one leaf.
func (sp *Spec) expand() []*specItem {
	var out []*specItem
	var walk func(base *specItem, path []string, v Value)
	walk = func(base *specItem, path []string, v Value) {
		if v.k == kObject && len(v.o.keys) > 0 {
			for _, k := range v.o.keys {
				walk(base, append(append([]string(nil), path...), k), v.o.m[k])
			}
			return
		}
		c := *base
		c.kind, c.path, c.value, c.hasVal = kindKey, path, v, true
		out = append(out, &c)
	}
	for _, it := range sp.items {
		if it.kind != kindKeys {
			out = append(out, it)
			continue
		}
		if !it.hasVal {
			c := *it
			c.kind = kindKey
			out = append(out, &c)
			continue
		}
		if len(it.path) == 0 && len(it.value.o.keys) == 0 {
			continue
		}
		walk(it, it.path, it.value)
	}
	return out
}

// FormatOf gives the format of a shared file: the `format` of its spec, else
// <explicit> (the `format` of its [files] entry), else the extension of
// <target>.
func FormatOf(sp *Spec, explicit, target string) Format {
	if sp != nil && sp.Format != "" {
		return sp.Format
	}
	if explicit != "" {
		return Format(explicit)
	}
	switch filepath.Ext(target) {
	case ".json":
		return JSON
	case ".jsonc":
		return JSONC
	case ".yaml", ".yml":
		return YAML
	case ".toml":
		return TOML
	}
	return Text
}

// sortedStrings gives a sorted copy without duplicates.
func sortedStrings(in []string) []string {
	set := map[string]bool{}
	var out []string
	for _, s := range in {
		if !set[s] {
			set[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
