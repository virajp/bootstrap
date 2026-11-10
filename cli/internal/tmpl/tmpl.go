// Package tmpl renders the templates of bootstrap with text/template. It
// replaces the gomplate CLI that the bash engine runs. It does not import the
// gomplate library: that library links the clients of AWS, GCP, Vault and
// Kubernetes (decision D2). Instead, it copies the few gomplate functions that
// the templates and the engine use, with the same argument order, the same
// types and the same output.
//
// The delimiters are "<%" and "%>". The context is any value that
// encoding/json can produce: use DecodeJSON for a JSON text, because it gives
// the same number types as gomplate (int, uint64 and float64).
//
// A namespace function such as test.IsKind is a function that returns a value
// with methods, as in gomplate. So the errors have the same text, for example
// "error calling Fail: template generation failed: msg".
//
// Inventory of the functions (46 .tmpl files, the tools.toml files and the
// inline templates of cli/). The root functions are has, default,
// dict, join and required. The standard functions of text/template are and,
// or, not, eq, ne, lt, gt, index, len, print, printf and slice.
//
//	coll:    Append Has Merge Omit Pick Slice
//	conv:    Join ToString
//	data:    JSONArray ToJSON ToJSONPretty ToTOML
//	env:     Getenv
//	file:    Read
//	math:    Add Seq Sub
//	regexp:  Replace
//	strings: Contains HasPrefix Quote ReplaceAll Split SplitN ToUpper
//	         TrimPrefix TrimSpace
//	test:    Fail IsKind Kind
//	tmpl:    Exec
//
// A template can define named blocks with "define", and tmpl.Exec runs them.
package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Options changes one render. A nil *Options gives the defaults.
type Options struct {
	// MissingKey is "error" (the default), "default" or "zero". It is the
	// flag --missing-key of gomplate. With "default", a missing map key prints
	// as "<no value>".
	MissingKey string
	// Env reads an environment variable for env.Getenv. A nil Env reads the
	// real environment.
	Env func(name string) string
}

// Render renders src as a template. The name is only for the error messages.
// The context ctx is the value of "." in the template. It returns the
// rendered bytes, or an error with the text/template message.
func Render(name string, src []byte, ctx any, opt *Options) ([]byte, error) {
	if opt == nil {
		opt = &Options{}
	}
	missing := opt.MissingKey
	if missing == "" {
		missing = "error"
	}
	getenv := opt.Env
	if getenv == nil {
		getenv = os.Getenv
	}
	t := template.New(name).Delims("<%", "%>").Option("missingkey=" + missing)
	t.Funcs(funcMap(t, ctx, getenv))
	if _, err := t.Parse(string(src)); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeJSON decodes a JSON text as gomplate decodes a JSON datasource: an
// object is a map[string]any, an array is a []any, an integer is an int (or a
// uint64 if it is too large for an int), and any other number is a float64.
func DecodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected data after the JSON value")
	}
	return normalize(v), nil
}

func normalize(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			x[k] = normalize(e)
		}
		return x
	case []any:
		for i, e := range x {
			x[i] = normalize(e)
		}
		return x
	case json.Number:
		s := x.String()
		if !strings.ContainsAny(s, ".eE") {
			if i, err := strconv.ParseInt(s, 10, strconv.IntSize); err == nil {
				return int(i)
			}
			if u, err := strconv.ParseUint(s, 10, 64); err == nil {
				return u
			}
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// gomplate reads a number that is too large for a float64 as a string.
			return s
		}
		return f
	}
	return v
}

var callRe = regexp.MustCompile(`.*error calling [A-Za-z.]*: `)

// Message returns the text that the bash engine shows for a render error: the
// text after the last "error calling <function>: ", else the whole error.
// The bash engine cuts the text with sed, one line at a time.
func Message(err error) string {
	s := err.Error()
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if loc := callRe.FindStringIndex(line); loc != nil {
			return line[loc[1]:]
		}
	}
	return s
}
