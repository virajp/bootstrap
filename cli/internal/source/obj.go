package source

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Obj is a JSON object with a live key order. The jq merge of the bash engine
// keeps the order of the keys of the first document and appends the new keys,
// so the order is part of the output.
//
// The other JSON values are nil, bool, int64, float64, string, []any and *Obj.
type Obj struct {
	keys []string
	m    map[string]any
}

// NewObj gives an empty object.
func NewObj() *Obj { return &Obj{m: map[string]any{}} }

// Keys gives the keys in the live order. The caller must not change the slice.
func (o *Obj) Keys() []string { return o.keys }

// Len gives the number of keys.
func (o *Obj) Len() int { return len(o.keys) }

// Has tells if the object has <key>.
func (o *Obj) Has(key string) bool { _, ok := o.m[key]; return ok }

// Get gives the value of <key>, or nil.
func (o *Obj) Get(key string) any { return o.m[key] }

// Lookup gives the value of <key> and tells if it is there.
func (o *Obj) Lookup(key string) (any, bool) { v, ok := o.m[key]; return v, ok }

// Set sets <key>. A new key goes at the end.
func (o *Obj) Set(key string, v any) {
	if _, ok := o.m[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.m[key] = v
}

// Clone gives a copy of the object. The values are shared.
func (o *Obj) Clone() *Obj {
	c := &Obj{keys: slices.Clone(o.keys), m: make(map[string]any, len(o.m))}
	for k, v := range o.m {
		c.m[k] = v
	}
	return c
}

// FromMap converts a decoded TOML value. The keys of each table are sorted,
// as gomplate sorts them when it writes JSON. A time value becomes a string.
func FromMap(v any) any {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		o := &Obj{keys: keys, m: make(map[string]any, len(x))}
		for _, k := range keys {
			o.m[k] = FromMap(x[k])
		}
		return o
	case []map[string]any:
		l := make([]any, len(x))
		for i, e := range x {
			l[i] = FromMap(e)
		}
		return l
	case []any:
		l := make([]any, len(x))
		for i, e := range x {
			l[i] = FromMap(e)
		}
		return l
	case int:
		return int64(x)
	case time.Time:
		return x.Format(time.RFC3339Nano)
	default:
		return v
	}
}

// Kind gives the jq type name of <v>.
func Kind(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case int64, float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case *Obj:
		return "object"
	}
	return fmt.Sprintf("%T", v)
}

func num(v any) (float64, bool) {
	switch x := v.(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, true
	}
	return 0, false
}

// Equal is the jq compare: objects ignore the key order, and numbers compare
// by value.
func Equal(a, b any) bool {
	if fa, ok := num(a); ok {
		fb, ok := num(b)
		return ok && fa == fb
	}
	switch x := a.(type) {
	case *Obj:
		y, ok := b.(*Obj)
		if !ok || len(x.keys) != len(y.keys) {
			return false
		}
		for _, k := range x.keys {
			v, ok := y.m[k]
			if !ok || !Equal(x.m[k], v) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !Equal(x[i], y[i]) {
				return false
			}
		}
		return true
	}
	return a == b
}

// JSON gives the compact JSON text of <v>, as `tojson` of jq gives it.
func JSON(v any) string {
	var b strings.Builder
	writeJSON(&b, v)
	return b.String()
}

func writeJSON(b *strings.Builder, v any) {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case float64:
		switch {
		case math.IsNaN(x) || math.IsInf(x, 0):
			b.WriteString("null")
		case x == math.Trunc(x) && math.Abs(x) < 1e17:
			b.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
		default:
			b.WriteString(strconv.FormatFloat(x, 'g', 17, 64))
		}
	case string:
		writeString(b, x)
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSON(b, e)
		}
		b.WriteByte(']')
	case *Obj:
		b.WriteByte('{')
		for i, k := range x.keys {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, k)
			b.WriteByte(':')
			writeJSON(b, x.m[k])
		}
		b.WriteByte('}')
	default:
		writeString(b, fmt.Sprint(x))
	}
}

// writeString writes a JSON string with the escapes of jq: the quote, the
// backslash, the control characters and DEL.
func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && n == 1:
			b.WriteString("�")
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(b, `\u%04x`, r)
		default:
			b.WriteString(s[i : i+n])
		}
		i += n
	}
	b.WriteByte('"')
}
