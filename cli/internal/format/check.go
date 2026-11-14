package format

import (
	"os"
	"strings"
)

// ParseCheck examines <content> as <f>. <show> names the target in the
// message. It gives nil when the content parses, and for Text. Else it gives a
// Failure with code 2 that holds the message and the output of the tool. An
// empty or whitespace-only JSON file does not parse. For YAML and TOML, the
// output of yq names a temp copy of <content>; ParseCheckFile names the file.
func ParseCheck(content []byte, f Format, show string) *Failure {
	if f != YAML && f != TOML {
		return parseCheck(content, "", f, show)
	}
	tmp, err := os.CreateTemp("", "bootstrap-parse-*")
	if err != nil {
		return &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, werr := tmp.Write(content)
	if cerr := tmp.Close(); werr != nil || cerr != nil {
		return &Failure{Code: 1, Text: "cannot create a temp directory", Line: "cannot create a temp directory"}
	}
	return parseCheck(content, tmp.Name(), f, show)
}

// ParseCheckFile is ParseCheck of the file <path>, as _se_parse_check: the
// output of yq names <path>. A file that cannot be read has no content.
func ParseCheckFile(path string, f Format, show string) *Failure {
	content, _ := os.ReadFile(path)
	return parseCheck(content, path, f, show)
}

func parseCheck(content []byte, path string, f Format, show string) *Failure {
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
		if r, yerr := yq(nil, "eval", "-p", string(f), "-o", "json", ".", path); yerr != nil {
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
