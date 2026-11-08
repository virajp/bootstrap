package format

import (
	"os"
	"strings"
)

// ParseCheck examines <content> as <f>. <show> names the target in the
// message. It gives nil when the content parses, and for Text. Else it gives a
// Failure with code 2 that holds the message and the output of the tool. An
// empty or whitespace-only JSON file does not parse.
func ParseCheck(content []byte, f Format, show string) *Failure {
	detail := ""
	ok := true
	switch f {
	case JSON:
		if strings.TrimFunc(string(content), isBashSpace) == "" {
			ok = false
		} else if _, err := parseStream(content); err != nil {
			ok = false
			detail = err.Error()
		}
	case YAML, TOML:
		tmp, err := os.CreateTemp("", "sync-engine-parse-*")
		if err != nil {
			return &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
		}
		defer os.Remove(tmp.Name())
		_, werr := tmp.Write(content)
		if cerr := tmp.Close(); werr != nil || cerr != nil {
			return &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
		}
		if r, yerr := yq(nil, "eval", "-p", string(f), "-o", "json", ".", tmp.Name()); yerr != nil {
			ok = false
			detail = r.both
		}
	}
	if ok {
		return nil
	}
	text := "the result does not parse as " + string(f) + " - the current file stays"
	return &Failure{Code: 2, Text: text, Line: show + ": " + text, Detail: detail}
}

// isBashSpace is the [:space:] class of bash.
func isBashSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	return false
}
