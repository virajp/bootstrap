package format

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// This file holds the JSON value of the merge: an ordered tree that reads and
// writes JSON in the form of jq 1.8. A number keeps its literal, so a
// canonical decNumber text comes out, as with jq. An object keeps the order
// in which its keys came in.

type kind uint8

const (
	kNull kind = iota
	kBool
	kNumber
	kString
	kArray
	kObject
)

// Value is a JSON value. The zero Value is null. A Value is never changed in
// place: each function that "changes" one gives a copy.
type Value struct {
	k kind
	b bool
	s string // The text of a string.
	n dec    // The number of a number.
	a []Value
	o *object
}

type object struct {
	keys []string
	m    map[string]Value
}

func newObject() *object { return &object{m: map[string]Value{}} }

// set gives the object with <key> set. A new key goes to the end.
func (o *object) set(key string, v Value) {
	if _, ok := o.m[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.m[key] = v
}

func (o *object) clone() *object {
	c := &object{keys: append([]string(nil), o.keys...), m: make(map[string]Value, len(o.m))}
	for k, v := range o.m {
		c.m[k] = v
	}
	return c
}

func (o *object) del(key string) {
	if _, ok := o.m[key]; !ok {
		return
	}
	delete(o.m, key)
	for i, k := range o.keys {
		if k == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			break
		}
	}
}

// Constructors.

func str(s string) Value          { return Value{k: kString, s: s} }
func arr(a []Value) Value         { return Value{k: kArray, a: a} }
func objv(o *object) Value        { return Value{k: kObject, o: o} }
func intVal(i int) Value          { return Value{k: kNumber, n: mustDec(strconv.Itoa(i))} }
func (v Value) isObject() bool    { return v.k == kObject }
func (v Value) isArray() bool     { return v.k == kArray }
func (v Value) has(k string) bool { _, ok := v.o.m[k]; return ok }

// typeName gives the name that the jq filter `type` gives.
func (v Value) typeName() string {
	switch v.k {
	case kBool:
		return "boolean"
	case kNumber:
		return "number"
	case kString:
		return "string"
	case kArray:
		return "array"
	case kObject:
		return "object"
	}
	return "null"
}

// kindOf gives table, list or scalar.
func (v Value) kindOf() string {
	switch v.k {
	case kObject:
		return "table"
	case kArray:
		return "list"
	}
	return "scalar"
}

// kindWord gives a jq type as words: `a list`, `null`, `a string`, ...
func kindWord(typ string) string {
	switch typ {
	case "array":
		return "a list"
	case "object":
		return "a table"
	case "null":
		return "null"
	}
	return "a " + typ
}

// get gives the value of <key> of an object, or null.
func (v Value) get(key string) Value {
	if v.k != kObject {
		return Value{}
	}
	return v.o.m[key]
}

// equal is the == of jq: numbers by value, objects without key order.
func equal(a, b Value) bool { return compare(a, b) == 0 }

func order(k kind, b bool) int {
	switch k {
	case kNull:
		return 0
	case kBool:
		if b {
			return 2
		}
		return 1
	case kNumber:
		return 3
	case kString:
		return 4
	case kArray:
		return 5
	}
	return 6
}

// compare is the order of jq: null, false, true, numbers, strings, lists,
// tables.
func compare(a, b Value) int {
	oa, ob := order(a.k, a.b), order(b.k, b.b)
	if oa != ob {
		return cmpInt(oa, ob)
	}
	switch a.k {
	case kNumber:
		return a.n.cmp(b.n)
	case kString:
		return strings.Compare(a.s, b.s)
	case kArray:
		for i := 0; i < len(a.a) && i < len(b.a); i++ {
			if c := compare(a.a[i], b.a[i]); c != 0 {
				return c
			}
		}
		return cmpInt(len(a.a), len(b.a))
	case kObject:
		ka := append([]string(nil), a.o.keys...)
		kb := append([]string(nil), b.o.keys...)
		sort.Strings(ka)
		sort.Strings(kb)
		for i := 0; i < len(ka) && i < len(kb); i++ {
			if c := strings.Compare(ka[i], kb[i]); c != 0 {
				return c
			}
		}
		if c := cmpInt(len(ka), len(kb)); c != 0 {
			return c
		}
		for _, k := range ka {
			if c := compare(a.o.m[k], b.o.m[k]); c != 0 {
				return c
			}
		}
	}
	return 0
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// dec is a decimal number as decNumber reads it: a sign, the digits of the
// coefficient without leading zeros, and a power of ten.
type dec struct {
	neg  bool
	coef string
	exp  int
	text string // Set for a number that is not printed from its parts.
}

func mustDec(s string) dec {
	d, ok := parseDec(s)
	if !ok {
		panic("bad number " + s)
	}
	return d
}

// parseDec reads a number token with the grammar of decNumber, which jq uses:
// an optional sign, digits with an optional point (or a point and digits), and
// an optional exponent. Leading zeros, a leading + and a trailing point are
// allowed.
func parseDec(tok string) (dec, bool) {
	var d dec
	i := 0
	if i < len(tok) && (tok[i] == '+' || tok[i] == '-') {
		d.neg = tok[i] == '-'
		i++
	}
	start := i
	for i < len(tok) && tok[i] >= '0' && tok[i] <= '9' {
		i++
	}
	digits := tok[start:i]
	frac := ""
	if i < len(tok) && tok[i] == '.' {
		i++
		fs := i
		for i < len(tok) && tok[i] >= '0' && tok[i] <= '9' {
			i++
		}
		frac = tok[fs:i]
	}
	if digits == "" && frac == "" {
		return d, false
	}
	exp := -len(frac)
	if i < len(tok) && (tok[i] == 'e' || tok[i] == 'E') {
		i++
		es := i
		if i < len(tok) && (tok[i] == '+' || tok[i] == '-') {
			i++
		}
		ds := i
		for i < len(tok) && tok[i] >= '0' && tok[i] <= '9' {
			i++
		}
		if ds == i || i-ds > 9 {
			return d, false
		}
		e, _ := strconv.Atoi(tok[es:i])
		exp += e
	}
	if i != len(tok) {
		return d, false
	}
	coef := strings.TrimLeft(digits+frac, "0")
	if coef == "" {
		coef = "0"
	}
	d.coef, d.exp = coef, exp
	return d, true
}

// String gives the decNumber "to scientific string" form.
func (d dec) String() string {
	if d.text != "" {
		return d.text
	}
	sign := ""
	if d.neg {
		sign = "-"
	}
	n := len(d.coef)
	adj := d.exp + n - 1
	if d.exp <= 0 && adj >= -6 {
		switch {
		case d.exp == 0:
			return sign + d.coef
		case n > -d.exp:
			return sign + d.coef[:n+d.exp] + "." + d.coef[n+d.exp:]
		}
		return sign + "0." + strings.Repeat("0", -d.exp-n) + d.coef
	}
	s := d.coef[:1]
	if n > 1 {
		s += "." + d.coef[1:]
	}
	e := strconv.Itoa(adj)
	if adj >= 0 {
		e = "+" + e
	}
	return sign + s + "E" + e
}

// normal gives the coefficient without trailing zeros and the matching
// exponent.
func (d dec) normal() (coef string, exp int) {
	if d.coef == "0" {
		return "0", 0
	}
	coef, exp = d.coef, d.exp
	t := strings.TrimRight(coef, "0")
	return t, exp + len(coef) - len(t)
}

func (d dec) cmp(o dec) int {
	ca, ea := d.normal()
	cb, eb := o.normal()
	za, zb := ca == "0", cb == "0"
	switch {
	case za && zb:
		return 0
	case za:
		if o.neg {
			return 1
		}
		return -1
	case zb:
		if d.neg {
			return -1
		}
		return 1
	case d.neg != o.neg:
		if d.neg {
			return -1
		}
		return 1
	}
	// Same sign, both not zero: compare the magnitude.
	m := magnitude(ca, ea, cb, eb)
	if d.neg {
		return -m
	}
	return m
}

func magnitude(ca string, ea int, cb string, eb int) int {
	adjA, adjB := ea+len(ca)-1, eb+len(cb)-1
	if adjA != adjB {
		return cmpInt(adjA, adjB)
	}
	n := len(ca)
	if len(cb) > n {
		n = len(cb)
	}
	ca += strings.Repeat("0", n-len(ca))
	cb += strings.Repeat("0", n-len(cb))
	return strings.Compare(ca, cb)
}

// Writing.

// marshal gives the jq form of <v>: a 2-space indent and one newline at the
// end.
func (v Value) marshal() []byte {
	var b strings.Builder
	v.write(&b, 0)
	b.WriteByte('\n')
	return []byte(b.String())
}

func (v Value) write(b *strings.Builder, depth int) {
	switch v.k {
	case kNull:
		b.WriteString("null")
	case kBool:
		if v.b {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case kNumber:
		b.WriteString(v.n.String())
	case kString:
		writeString(b, v.s)
	case kArray:
		if len(v.a) == 0 {
			b.WriteString("[]")
			return
		}
		b.WriteString("[")
		for i, e := range v.a {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\n" + strings.Repeat("  ", depth+1))
			e.write(b, depth+1)
		}
		b.WriteString("\n" + strings.Repeat("  ", depth) + "]")
	case kObject:
		if len(v.o.keys) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{")
		for i, k := range v.o.keys {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString("\n" + strings.Repeat("  ", depth+1))
			writeString(b, k)
			b.WriteString(": ")
			v.o.m[k].write(b, depth+1)
		}
		b.WriteString("\n" + strings.Repeat("  ", depth) + "}")
	}
}

const hexDigits = "0123456789abcdef"

// writeString writes a string as jq does: the escapes \" \\ \b \f \n \r \t,
// \u00xx for the other control characters and DEL, and all else as it is.
func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\b':
			b.WriteString(`\b`)
		case r == '\f':
			b.WriteString(`\f`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			b.WriteString(`\u00`)
			b.WriteByte(hexDigits[r>>4])
			b.WriteByte(hexDigits[r&0xf])
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
}

// validUTF8 gives <s> with each invalid byte replaced by U+FFFD.
func validUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && w == 1 {
			b.WriteRune(utf8.RuneError)
		} else {
			b.WriteString(s[i : i+w])
		}
		i += w
	}
	return b.String()
}
