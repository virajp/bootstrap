// Package format holds the formats of a target: the format from its path, the
// parse check of a file and the merge of a merge source into a live file.
//
// It replaces _se_format and _se_parse_check of the bash files write and
// merge, with the same results, messages and exit codes. JSON is
// read and written as jq does it. YAML and TOML go through the command of yq,
// so their comments stay. Contract: docs/design/cli.md ("The parse
// check" and "Merge mode").
package format

import (
	"regexp"
	"strings"
)

// Format is the format of a target.
type Format string

// The formats. Text is every target that is not JSON, YAML or TOML, and a
// JSON target with a comment line (JSONC): the engine does not examine it.
const (
	JSON Format = "json"
	YAML Format = "yaml"
	TOML Format = "toml"
	Text Format = "text"
)

var commentLine = regexp.MustCompile(`^[ \t\n\v\f\r]*//`)

// Detect gives the format of <target>, whose content is <content>. A given
// <explicit> format wins. Else the extension of <target> without `.tmpl`
// decides it. A `.json` file with a line that starts with `//` after spaces is
// JSONC, which jq cannot read, so its format is Text. <content> can be nil
// when only the path matters.
func Detect(target string, content []byte, explicit string) Format {
	if explicit != "" {
		return Format(explicit)
	}
	target = strings.TrimSuffix(target, ".tmpl")
	switch {
	case strings.HasSuffix(target, ".json"):
		for _, line := range strings.Split(string(content), "\n") {
			if commentLine.MatchString(line) {
				return Text
			}
		}
		return JSON
	case strings.HasSuffix(target, ".yaml"), strings.HasSuffix(target, ".yml"):
		return YAML
	case strings.HasSuffix(target, ".toml"):
		return TOML
	}
	return Text
}

// Failure is a failed parse check or merge. The caller reports it in this
// order: first it prints Stderr as it is to stderr, then (with Text) it calls
// the outcome `failed` for the target, with Text and Line (Line is the log line
// of the verbose mode), then it prints IndentedDetail to stderr. A Failure
// without Text has no outcome: the bash code printed only the output of the
// tool.
type Failure struct {
	Code   int    // The exit code: 2, or 1 for a failed yq call at the end of a merge.
	Text   string // The reason of the outcome `failed`; it does not repeat the path.
	Line   string // The line of the verbose log, with the path.
	Stderr string // The stderr of a tool in a merge, to print before the outcome.
	Detail string // The output of the tool in a parse check, to print after the outcome.
}

func (f *Failure) Error() string {
	if f.Line != "" {
		return f.Line
	}
	return f.Stderr
}

// IndentedDetail gives Detail as the parse check prints it: two spaces before
// each line and a newline at the end. It gives "" for no detail.
func (f *Failure) IndentedDetail() string {
	d := strings.TrimRight(f.Detail, "\n")
	if d == "" {
		return ""
	}
	return "  " + strings.ReplaceAll(d, "\n", "\n  ") + "\n"
}
