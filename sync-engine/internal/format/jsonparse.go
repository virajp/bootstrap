package format

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// parseError is a parse error in the words of jq.
type parseError struct{ msg string }

func (e *parseError) Error() string { return "jq: parse error: " + e.msg }

// parser reads a stream of JSON values as jq does. It accepts what jq accepts
// and more than the JSON standard: no separator between top-level values, the
// literals nan, NaN, Infinity and -Infinity, and numbers such as +1, .5, 1.
// and 0123. It rejects what jq rejects.
type parser struct {
	s   string
	pos int
}

func (p *parser) fail(format string, args ...any) *parseError {
	line, col := 1, 0
	for i := 0; i < p.pos && i < len(p.s); i++ {
		if p.s[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	msg := fmt.Sprintf(format, args...)
	if p.pos >= len(p.s) {
		msg += " at EOF"
	}
	return &parseError{fmt.Sprintf("%s at line %d, column %d", msg, line, col)}
}

// parseStream gives each value of <data>, in order.
func parseStream(data []byte) ([]Value, *parseError) {
	p := &parser{s: strings.TrimPrefix(string(data), "\xef\xbb\xbf")}
	var out []Value
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return out, nil
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
}

func (p *parser) skipSpace() {
	for p.pos < len(p.s) {
		switch p.s[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func isStructural(c byte) bool {
	switch c {
	case '[', ']', '{', '}', ',', ':', '"', ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func (p *parser) value() (Value, *parseError) {
	switch c := p.s[p.pos]; c {
	case '[':
		return p.array()
	case '{':
		return p.object()
	case '"':
		s, err := p.str()
		return str(s), err
	case ']':
		p.pos++
		return Value{}, p.fail("Unmatched ']'")
	case '}':
		p.pos++
		return Value{}, p.fail("Unmatched '}'")
	case ',':
		p.pos++
		return Value{}, p.fail("Expected value before ','")
	case ':':
		p.pos++
		return Value{}, p.fail("':' not as part of an object")
	}
	return p.token()
}

func (p *parser) token() (Value, *parseError) {
	start := p.pos
	for p.pos < len(p.s) && !isStructural(p.s[p.pos]) {
		p.pos++
	}
	tok := p.s[start:p.pos]
	switch tok {
	case "true":
		return Value{k: kBool, b: true}, nil
	case "false":
		return Value{k: kBool}, nil
	case "null", "nan", "NaN", "-nan", "-NaN":
		return Value{}, nil
	case "Infinity":
		return infinity(false), nil
	case "-Infinity":
		return infinity(true), nil
	}
	d, ok := parseDec(tok)
	if !ok {
		if c := tok[0]; c == 't' || c == 'f' || c == 'n' {
			return Value{}, p.fail("Invalid literal")
		}
		return Value{}, p.fail("Invalid numeric literal")
	}
	return Value{k: kNumber, n: d}, nil
}

// infinity is the largest double, as jq reads Infinity.
func infinity(neg bool) Value {
	d := mustDec("1.7976931348623157e+308")
	d.neg = neg
	d.text = "1.7976931348623157e+308"
	if neg {
		d.text = "-" + d.text
	}
	return Value{k: kNumber, n: d}
}

func (p *parser) array() (Value, *parseError) {
	p.pos++
	items := []Value{}
	p.skipSpace()
	if p.pos < len(p.s) && p.s[p.pos] == ']' {
		p.pos++
		return arr(items), nil
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		if p.s[p.pos] == ']' {
			p.pos++
			return Value{}, p.fail("Expected another array element")
		}
		v, err := p.value()
		if err != nil {
			return Value{}, err
		}
		items = append(items, v)
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		switch p.s[p.pos] {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return arr(items), nil
		default:
			p.pos++
			return Value{}, p.fail("Expected separator between values")
		}
	}
}

func (p *parser) object() (Value, *parseError) {
	p.pos++
	o := newObject()
	p.skipSpace()
	if p.pos < len(p.s) && p.s[p.pos] == '}' {
		p.pos++
		return objv(o), nil
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		if p.s[p.pos] != '"' {
			if p.s[p.pos] == '}' {
				p.pos++
				return Value{}, p.fail("Objects must consist of key:value pairs")
			}
			if _, err := p.value(); err != nil {
				return Value{}, err
			}
			return Value{}, p.fail("Object keys must be strings")
		}
		key, err := p.str()
		if err != nil {
			return Value{}, err
		}
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		if p.s[p.pos] != ':' {
			p.pos++
			return Value{}, p.fail("Expected separator between values")
		}
		p.pos++
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		v, err := p.value()
		if err != nil {
			return Value{}, err
		}
		o.set(key, v)
		p.skipSpace()
		if p.pos >= len(p.s) {
			return Value{}, p.fail("Unfinished JSON term")
		}
		switch p.s[p.pos] {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return objv(o), nil
		default:
			p.pos++
			return Value{}, p.fail("Expected separator between values")
		}
	}
}

func (p *parser) str() (string, *parseError) {
	p.pos++ // The opening quote.
	var b strings.Builder
	for {
		if p.pos >= len(p.s) {
			return "", p.fail("Unfinished string")
		}
		c := p.s[p.pos]
		switch {
		case c == '"':
			p.pos++
			return validUTF8(b.String()), nil
		case c < 0x20:
			p.pos++
			return "", p.fail("Invalid string: control characters from U+0000 through U+001F must be escaped")
		case c != '\\':
			b.WriteByte(c)
			p.pos++
			continue
		}
		p.pos++
		if p.pos >= len(p.s) {
			return "", p.fail("Unfinished string")
		}
		e := p.s[p.pos]
		p.pos++
		switch e {
		case '"', '\\', '/':
			b.WriteByte(e)
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case 'u':
			r, err := p.hex4()
			if err != nil {
				return "", err
			}
			if utf16.IsSurrogate(r) && r < 0xdc00 {
				// A high surrogate needs a low one after it.
				if !strings.HasPrefix(p.s[p.pos:], `\u`) {
					return "", p.fail(`Invalid \uXXXX\uXXXX surrogate pair escape`)
				}
				p.pos += 2
				lo, err := p.hex4()
				if err != nil {
					return "", err
				}
				if lo < 0xdc00 || lo > 0xdfff {
					return "", p.fail(`Invalid \uXXXX\uXXXX surrogate pair escape`)
				}
				r = utf16.DecodeRune(r, lo)
			} else if utf16.IsSurrogate(r) {
				r = utf8.RuneError
			}
			b.WriteRune(r)
		default:
			return "", p.fail("Invalid escape")
		}
	}
}

func (p *parser) hex4() (rune, *parseError) {
	if p.pos+4 > len(p.s) {
		p.pos = len(p.s)
		return 0, p.fail(`Invalid \uXXXX escape`)
	}
	var r rune
	for i := 0; i < 4; i++ {
		c := p.s[p.pos+i]
		var d byte
		switch {
		case c >= '0' && c <= '9':
			d = c - '0'
		case c >= 'a' && c <= 'f':
			d = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			d = c - 'A' + 10
		default:
			p.pos += i + 1
			return 0, p.fail(`Invalid \uXXXX escape`)
		}
		r = r<<4 | rune(d)
	}
	p.pos += 4
	return r, nil
}
