package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/ugorji/go/codec"
)

// funcMap builds the function set of one render. The functions copy gomplate:
// the argument order (the input is often the last argument, so that a
// pipeline works), the types and the output.
func funcMap(t *template.Template, root any, getenv func(string) string) template.FuncMap {
	coll, conv, data, env := &collNS{}, &convNS{}, &dataNS{}, &envNS{getenv}
	file, mth, rx, str := &fileNS{}, &mathNS{}, &regexpNS{}, &stringsNS{}
	tst, tm := &testNS{}, &tmplNS{t: t, root: root}
	return template.FuncMap{
		"coll":    func() *collNS { return coll },
		"conv":    func() *convNS { return conv },
		"data":    func() *dataNS { return data },
		"env":     func() *envNS { return env },
		"file":    func() *fileNS { return file },
		"math":    func() *mathNS { return mth },
		"regexp":  func() *regexpNS { return rx },
		"strings": func() *stringsNS { return str },
		"test":    func() *testNS { return tst },
		"tmpl":    func() *tmplNS { return tm },
		// The root aliases of gomplate.
		"has":      coll.Has,
		"dict":     coll.Dict,
		"join":     conv.Join,
		"default":  conv.Default,
		"required": conv.Required,
	}
}

// ToString is conv.ToString of gomplate.
func toString(in any) string {
	if in == nil {
		return "nil"
	}
	switch v := in.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	}
	rv := reflect.ValueOf(in)
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	return fmt.Sprintf("%v", rv.Interface())
}

// ---- coll ----

type collNS struct{}

func (*collNS) Slice(args ...any) []any {
	if args == nil {
		return []any{}
	}
	return args
}

// Has reports whether a map has the key, or whether a list has the value.
func (*collNS) Has(in any, key string) bool {
	if in == nil {
		return false
	}
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Map:
		kt := rv.Type().Key()
		if kt.Kind() != reflect.String {
			return false
		}
		return rv.MapIndex(reflect.ValueOf(key).Convert(kt)).IsValid()
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if reflect.DeepEqual(rv.Index(i).Interface(), key) {
				return true
			}
		}
	}
	return false
}

// Dict builds a map from key and value pairs. An odd last key gets "".
func (*collNS) Dict(in ...any) (map[string]any, error) {
	out := make(map[string]any, len(in)/2)
	for i := 0; i < len(in); i += 2 {
		k := toString(in[i])
		var v any = ""
		if i+1 < len(in) {
			v = in[i+1]
		}
		out[k] = v
	}
	return out, nil
}

// Append returns a copy of the list with v at the end.
func (*collNS) Append(v any, list any) ([]any, error) {
	rv := reflect.ValueOf(list)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected an array or slice, but got a %s", rv.Kind())
	}
	out := make([]any, 0, rv.Len()+1)
	for i := 0; i < rv.Len(); i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return append(out, v), nil
}

func lastMap(in []any) (map[string]any, []string, error) {
	if len(in) < 2 {
		return nil, nil, fmt.Errorf("wrong number of args: wanted 2 or more, got %d", len(in))
	}
	keys := make([]string, 0, len(in)-1)
	for _, k := range in[:len(in)-1] {
		ks, ok := k.(string)
		if !ok {
			return nil, nil, fmt.Errorf("wrong key type: must be string, got %T (%v)", k, k)
		}
		keys = append(keys, ks)
	}
	m, ok := in[len(in)-1].(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("wrong map type: must be map[string]any, got %T", in[len(in)-1])
	}
	return m, keys, nil
}

// Omit returns a copy of the map (the last argument) without the keys.
func (*collNS) Omit(in ...any) (map[string]any, error) {
	m, keys, err := lastMap(in)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	for _, k := range keys {
		delete(out, k)
	}
	return out, nil
}

// Pick returns a copy of the map (the last argument) with only the keys.
func (*collNS) Pick(in ...any) (map[string]any, error) {
	m, keys, err := lastMap(in)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

// Merge merges the maps. The first map has the highest priority: a key that
// it has stays as it is, even if its value is empty. Two maps at the same key
// merge in the same way. The first map is not changed.
func (*collNS) Merge(dst map[string]any, srcs ...map[string]any) (map[string]any, error) {
	dst = deepCopy(dst).(map[string]any)
	for _, src := range srcs {
		mergeInto(dst, src)
	}
	return dst, nil
}

func mergeInto(dst, src map[string]any) {
	for k, sv := range src {
		dv, ok := dst[k]
		if !ok {
			dst[k] = deepCopy(sv)
			continue
		}
		dm, dok := dv.(map[string]any)
		sm, sok := sv.(map[string]any)
		if dok && sok {
			mergeInto(dm, sm)
		}
	}
}

func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		if x == nil {
			return x
		}
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = deepCopy(e)
		}
		return out
	case []any:
		if x == nil {
			return x
		}
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepCopy(e)
		}
		return out
	}
	return v
}

// ---- conv ----

type convNS struct{}

func (*convNS) ToString(in any) string { return toString(in) }

// Join joins a list with sep. The elements are changed with ToString.
func (*convNS) Join(in any, sep string) (string, error) {
	rv := reflect.ValueOf(in)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return "", fmt.Errorf("input to Join must be an array: expected an array or slice, but got a %s", rv.Kind())
	}
	parts := make([]string, rv.Len())
	for i := range parts {
		parts[i] = toString(rv.Index(i).Interface())
	}
	return strings.Join(parts, sep), nil
}

// Default returns in if it is true by the rules of text/template, else def.
func (*convNS) Default(def any, in any) any {
	if truth, ok := template.IsTrue(in); ok && truth {
		return in
	}
	return def
}

// Required returns the value, or an error with the message if it is nil or an
// empty string. With one argument, that argument is the value.
func (*convNS) Required(args ...any) (any, error) {
	var msg string
	var val any
	switch len(args) {
	case 1:
		val = args[0]
	case 2:
		m, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("at <1>: expected string; found %T", args[0])
		}
		msg, val = m, args[1]
	default:
		return nil, fmt.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}
	if msg == "" {
		msg = "can not render template: a required value was not set"
	}
	if val == nil {
		return nil, fmt.Errorf("%s", msg)
	}
	if s, ok := val.(string); ok && s == "" {
		return nil, fmt.Errorf("%s", msg)
	}
	return val, nil
}

// ---- data ----

type dataNS struct{}

func (*dataNS) ToJSON(in any) (string, error) {
	h := &codec.JsonHandle{}
	h.Canonical = true
	var buf bytes.Buffer
	if err := codec.NewEncoder(&buf, h).Encode(in); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (d *dataNS) ToJSONPretty(indent string, in any) (string, error) {
	s, err := d.ToJSON(in)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := json.Indent(&out, []byte(s), "", indent); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (*dataNS) ToTOML(in any) (string, error) {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(in); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ToTOMLInline gives <in> as one TOML inline value: a string, a number, a
// boolean, an array or an inline table with sorted keys. A merge spec uses it
// to write a value of the template context.
func (*dataNS) ToTOMLInline(in any) (string, error) {
	var b strings.Builder
	if err := tomlInline(&b, in); err != nil {
		return "", err
	}
	return b.String(), nil
}

// tomlString gives <s> as a TOML basic string.
func tomlString(s string) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(buf.String(), "\n")
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func tomlInline(b *strings.Builder, in any) error {
	switch x := in.(type) {
	case nil:
		return fmt.Errorf("TOML has no null value")
	case string:
		b.WriteString(tomlString(x))
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		fmt.Fprintf(b, "%d", x)
	case float32, float64:
		f := reflect.ValueOf(x).Float()
		s := strconv.FormatFloat(f, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eEn") {
			s += ".0"
		}
		b.WriteString(s)
	case []any:
		b.WriteString("[ ")
		for i, e := range x {
			if i > 0 {
				b.WriteString(", ")
			}
			if err := tomlInline(b, e); err != nil {
				return err
			}
		}
		b.WriteString(" ]")
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("{ ")
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			if bareKey.MatchString(k) {
				b.WriteString(k)
			} else {
				b.WriteString(tomlString(k))
			}
			b.WriteString(" = ")
			if err := tomlInline(b, x[k]); err != nil {
				return err
			}
		}
		b.WriteString(" }")
	default:
		rv := reflect.ValueOf(in)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			l := make([]any, rv.Len())
			for i := range l {
				l[i] = rv.Index(i).Interface()
			}
			return tomlInline(b, l)
		case reflect.Map:
			m := map[string]any{}
			for _, k := range rv.MapKeys() {
				m[fmt.Sprint(k.Interface())] = rv.MapIndex(k).Interface()
			}
			return tomlInline(b, m)
		}
		return fmt.Errorf("cannot write a %T as a TOML value", in)
	}
	return nil
}

// ToTOMLFlat gives the table <in> as a TOML document in the style of the
// dprint config of bootstrap: no indent, the keys in sorted order, the values
// of a table before its sub-tables, and the arrays as `[ "a", "b" ]`. A blank
// line comes before each top-level table. A generated file that a repository
// commits thus stays the same after dprint formats it.
func (*dataNS) ToTOMLFlat(in any) (string, error) {
	m, ok := tomlTable(in)
	if !ok {
		return "", fmt.Errorf("cannot write a %T as a TOML document", in)
	}
	var b strings.Builder
	if err := tomlFlat(&b, nil, m); err != nil {
		return "", err
	}
	return strings.TrimPrefix(b.String(), "\n"), nil
}

// tomlTable gives <in> as a map when it is a table.
func tomlTable(in any) (map[string]any, bool) {
	if m, ok := in.(map[string]any); ok {
		return m, true
	}
	rv := reflect.ValueOf(in)
	if rv.Kind() != reflect.Map {
		return nil, false
	}
	m := map[string]any{}
	for _, k := range rv.MapKeys() {
		m[fmt.Sprint(k.Interface())] = rv.MapIndex(k).Interface()
	}
	return m, true
}

func tomlKey(k string) string {
	if bareKey.MatchString(k) {
		return k
	}
	return tomlString(k)
}

// tomlFlat writes the table <m> at the path <path>: its header, its values,
// then its sub-tables.
func tomlFlat(b *strings.Builder, path []string, m map[string]any) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(path) > 0 {
		if len(path) == 1 {
			b.WriteString("\n")
		}
		parts := make([]string, len(path))
		for i, p := range path {
			parts[i] = tomlKey(p)
		}
		b.WriteString("[" + strings.Join(parts, ".") + "]\n")
	}
	var tables []string
	for _, k := range keys {
		if _, ok := tomlTable(m[k]); ok {
			tables = append(tables, k)
			continue
		}
		b.WriteString(tomlKey(k) + " = ")
		if err := tomlInline(b, m[k]); err != nil {
			return err
		}
		b.WriteString("\n")
	}
	for _, k := range tables {
		t, _ := tomlTable(m[k])
		if err := tomlFlat(b, append(append([]string{}, path...), k), t); err != nil {
			return err
		}
	}
	return nil
}

func (*dataNS) JSONArray(in any) ([]any, error) {
	s := toString(in)
	if strings.TrimSpace(s) == "" {
		return []any{nil}, nil
	}
	v, err := DecodeJSON([]byte(s))
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal array %s: %w", s, err)
	}
	a, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("unable to unmarshal array %s: cannot unmarshal into []interface {}", s)
	}
	return a, nil
}

// ---- env, file ----

type envNS struct{ getenv func(string) string }

// Getenv reads a variable. A second argument is the value for an empty one.
func (e *envNS) Getenv(key any, def ...any) string {
	k := toString(key)
	v := e.getenv(k)
	if v == "" {
		if f := e.getenv(k + "_FILE"); f != "" {
			if b, err := os.ReadFile(f); err == nil {
				v = strings.TrimSpace(string(b))
			}
		}
	}
	if v == "" && len(def) > 0 {
		return toString(def[0])
	}
	return v
}

type fileNS struct{}

func (*fileNS) Read(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ---- math ----

type mathNS struct{}

// num is a number of the math functions: an int64, or a float64.
type num struct {
	i       int64
	f       float64
	isFloat bool
}

func toNum(in any) (num, error) {
	switch v := in.(type) {
	case nil:
		return num{}, fmt.Errorf("could not convert <nil> to int64")
	case bool:
		if v {
			return num{i: 1}, nil
		}
		return num{}, nil
	case string:
		s := strings.TrimSpace(v)
		if i, err := strconv.ParseInt(s, 0, 64); err == nil {
			return num{i: i}, nil
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return num{}, fmt.Errorf("could not convert %q to int64: %w", v, err)
		}
		return num{f: f, isFloat: true}, nil
	}
	rv := reflect.ValueOf(in)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return num{i: rv.Int()}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if rv.Uint() > math.MaxInt64 {
			return num{}, fmt.Errorf("could not convert %v to int64, would overflow", in)
		}
		return num{i: int64(rv.Uint())}, nil
	case reflect.Float32, reflect.Float64:
		return num{f: rv.Float(), isFloat: true}, nil
	}
	return num{}, fmt.Errorf("could not convert %v to int64", in)
}

func (n num) float() float64 {
	if n.isFloat {
		return n.f
	}
	return float64(n.i)
}

func (*mathNS) arith(prefix string, a, b any, f func(int64, int64) int64, g func(float64, float64) float64) (any, error) {
	x, err := toNum(a)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", prefix, err)
	}
	y, err := toNum(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", prefix, err)
	}
	if x.isFloat || y.isFloat {
		return g(x.float(), y.float()), nil
	}
	return f(x.i, y.i), nil
}

func (m *mathNS) Add(n ...any) (any, error) {
	var acc any = int64(0)
	for _, v := range n {
		var err error
		acc, err = m.arith("expected number inputs", acc, v, func(a, b int64) int64 { return a + b }, func(a, b float64) float64 { return a + b })
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func (m *mathNS) Sub(a, b any) (any, error) {
	return m.arith("expected a number", a, b, func(a, b int64) int64 { return a - b }, func(a, b float64) float64 { return a - b })
}

// Seq returns the numbers from start to end, both included. One argument is
// the end, and the start is 1. A start after the end counts down.
func (*mathNS) Seq(n ...any) ([]int64, error) {
	var vals []int64
	for _, v := range n {
		x, err := toNum(v)
		if err != nil {
			return nil, fmt.Errorf("expected a number: %w", err)
		}
		if x.isFloat {
			vals = append(vals, int64(x.f))
		} else {
			vals = append(vals, x.i)
		}
	}
	start, end, step := int64(1), int64(0), int64(1)
	switch len(vals) {
	case 1:
		end = vals[0]
	case 2:
		start, end = vals[0], vals[1]
	case 3:
		start, end, step = vals[0], vals[1], vals[2]
	default:
		return nil, fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(vals))
	}
	if step == 0 {
		return []int64{}, nil
	}
	if end < start {
		if step > 0 {
			step = -step
		}
	} else if step < 0 {
		step = -step
	}
	out := []int64{}
	for i := start; (step > 0 && i <= end) || (step < 0 && i >= end); i += step {
		out = append(out, i)
	}
	return out, nil
}

// ---- regexp ----

type regexpNS struct{}

func (*regexpNS) Replace(expr, repl string, in any) (string, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return "", fmt.Errorf("error compiling expression: %w", err)
	}
	return re.ReplaceAllString(toString(in), repl), nil
}

// ---- strings ----

type stringsNS struct{}

func (*stringsNS) Contains(sub string, in any) bool   { return strings.Contains(toString(in), sub) }
func (*stringsNS) HasPrefix(p string, in any) bool    { return strings.HasPrefix(toString(in), p) }
func (*stringsNS) TrimPrefix(p string, in any) string { return strings.TrimPrefix(toString(in), p) }
func (*stringsNS) ReplaceAll(old, new string, in any) string {
	return strings.ReplaceAll(toString(in), old, new)
}
func (*stringsNS) Split(sep string, in any) []string { return strings.Split(toString(in), sep) }
func (*stringsNS) SplitN(sep string, n int, in any) []string {
	return strings.SplitN(toString(in), sep, n)
}
func (*stringsNS) ToUpper(in any) string   { return strings.ToUpper(toString(in)) }
func (*stringsNS) TrimSpace(in any) string { return strings.TrimSpace(toString(in)) }
func (*stringsNS) Quote(in any) string     { return strconv.Quote(toString(in)) }

// ---- test ----

type testNS struct{}

// Kind is the reflect kind of the value, as a string. A nil is "invalid".
func (*testNS) Kind(in any) string { return reflect.ValueOf(in).Kind().String() }

// IsKind compares the kind. The kind "number" matches each int, uint and
// float kind.
func (t *testNS) IsKind(kind string, in any) bool {
	k := t.Kind(in)
	if kind == "number" {
		switch reflect.ValueOf(in).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		}
		return false
	}
	return kind == k
}

// Fail stops the render with an error.
func (*testNS) Fail(in ...any) (string, error) {
	switch len(in) {
	case 0:
		return "", fmt.Errorf("template generation failed")
	case 1:
		return "", fmt.Errorf("template generation failed: %s", toString(in[0]))
	}
	return "", fmt.Errorf("wrong number of args: want 0 or 1, got %d", len(in))
}

// ---- tmpl ----

type tmplNS struct {
	t    *template.Template
	root any
}

// Exec runs a defined template. Without a context, it uses the root context.
func (n *tmplNS) Exec(name string, in ...any) (string, error) {
	t := n.t.Lookup(name)
	if t == nil {
		return "", fmt.Errorf("template %q not defined", name)
	}
	var ctx any = n.root
	switch len(in) {
	case 0:
	case 1:
		ctx = in[0]
	default:
		return "", fmt.Errorf("wrong number of args for Exec: want 1 or 2, got %d", len(in)+1)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}
