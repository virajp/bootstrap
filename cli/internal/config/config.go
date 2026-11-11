// Package config reads the user config file of the engine,
// ~/.config/bootstrap/config.toml. The file holds the defaults of the
// settings of the engine. A flag and a variable of the process come before
// it. The engine never writes the file.
//
// Contract: docs/design/cli.md.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Key is one setting of the engine.
type Key struct {
	Name   string // The key in the file, and the name in `bootstrap config`.
	Var    string // The variable of the process that the engine reads.
	Flag   string // The flag, without dashes.
	InFile bool   // The file can set it.
}

// Keys are the settings, in the order of `bootstrap config`.
var Keys = []Key{
	{Name: "jobs", Var: "BOOTSTRAP_JOBS", Flag: "jobs", InFile: true},
	{Name: "cache_dir", Var: "BOOTSTRAP_CACHE", Flag: "cache-dir", InFile: true},
	{Name: "url", Var: "BOOTSTRAP_URL", Flag: "url", InFile: true},
	{Name: "clone", Var: "BOOTSTRAP_CLONE", Flag: "clone", InFile: true},
	// The mise env of an adopted Mac exports the variable, so the file has
	// no key for it.
	{Name: "mac.repo_path", Var: "MACOS_SETUP_REPO_PATH", Flag: "repo-path"},
}

// File is the content of a config file.
type File struct {
	Missing bool              // The file does not exist.
	values  map[string]string // By key name.
}

// Value gives the value of the key <name> in the file.
func (f File) Value(name string) (string, bool) {
	v, ok := f.values[name]
	return v, ok
}

// Path gives the config file and its source: the flag --config, then the
// variable BOOTSTRAP_CONFIG, then $XDG_CONFIG_HOME/bootstrap/config.toml,
// else $HOME/.config/bootstrap/config.toml. An empty value is not set.
func Path(flag string, getenv func(string) string) (path, source string) {
	if flag != "" {
		return flag, "flag --config"
	}
	if v := getenv("BOOTSTRAP_CONFIG"); v != "" {
		return v, "variable BOOTSTRAP_CONFIG"
	}
	dir := getenv("XDG_CONFIG_HOME")
	if dir == "" {
		dir = getenv("HOME") + "/.config"
	}
	return dir + "/bootstrap/config.toml", "default"
}

// raw is the decoded file. A key that is not in the file stays nil. The
// table [mac] has no key now (D16): a key in it is an unknown key.
type raw struct {
	Jobs     *int64    `toml:"jobs"`
	CacheDir *string   `toml:"cache_dir"`
	URL      *string   `toml:"url"`
	Clone    *string   `toml:"clone"`
	Mac      *struct{} `toml:"mac"`
}

// types are the TOML types of the keys of raw, for the error of a wrong type.
var types = []struct{ key, typ, need string }{
	{"jobs", "Integer", "a number above 0"},
	{"cache_dir", "String", "a string"},
	{"url", "String", "a string"},
	{"clone", "String", "a string"},
	{"mac", "Hash", "a table"},
}

// Load reads the config file <path>. A missing file gives an empty File and
// no error. <home> expands a leading ~/ of a path. Each error names the file,
// and the key when it has one.
func Load(path, home string) (File, error) {
	f := File{values: map[string]string{}}
	st, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		f.Missing = true
		return f, nil
	}
	if err != nil {
		return f, fmt.Errorf("%s: cannot read the file", path)
	}
	if st.IsDir() {
		return f, fmt.Errorf("%s: is a folder, not a file", path)
	}
	var r raw
	md, err := toml.DecodeFile(path, &r)
	bad := func(key, format string, args ...any) error {
		return fmt.Errorf("%s: %s: %s", path, key, fmt.Sprintf(format, args...))
	}
	if err != nil {
		var pe toml.ParseError
		if errors.As(err, &pe) {
			return f, fmt.Errorf("%s: not valid TOML: line %d: %s", path, pe.Position.Line, pe.Message)
		}
		// The error of a wrong type does not give the key as a field, so
		// find the key with the TOML type of each key.
		for _, t := range types {
			if got := md.Type(t.key); got != "" && got != t.typ {
				return f, bad(t.key, "needs %s, not %s", t.need, typeName(got))
			}
		}
		return f, fmt.Errorf("%s: cannot read the file", path)
	}
	if keys := md.Undecoded(); len(keys) > 0 {
		k := keys[0]
		if md.Type(k...) == "Hash" {
			return f, bad(k.String(), "unknown table")
		}
		return f, bad(k.String(), "unknown key")
	}
	if r.Jobs != nil {
		if *r.Jobs < 1 {
			return f, bad("jobs", "needs a number above 0, not %d", *r.Jobs)
		}
		f.values["jobs"] = fmt.Sprint(*r.Jobs)
	}
	for _, s := range []struct {
		name  string
		value *string
		path  bool
	}{
		{"cache_dir", r.CacheDir, true},
		{"url", r.URL, false},
		{"clone", r.Clone, true},
	} {
		if s.value == nil {
			continue
		}
		v := *s.value
		if v == "" {
			return f, bad(s.name, "is empty")
		}
		if s.path {
			switch {
			case strings.HasPrefix(v, "~/"):
				v = home + v[1:]
			case !filepath.IsAbs(v):
				return f, bad(s.name, "needs an absolute path or a path that starts with ~/, not '%s'", v)
			}
		}
		f.values[s.name] = v
	}
	return f, nil
}

// typeName gives a TOML type of MetaData.Type, with its article.
func typeName(t string) string {
	switch t {
	case "String":
		return "a string"
	case "Integer":
		return "an integer"
	case "Float":
		return "a float"
	case "Bool":
		return "a boolean"
	case "Array", "ArrayHash":
		return "an array"
	case "Hash":
		return "a table"
	default:
		return "a date"
	}
}
