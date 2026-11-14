package cli

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"text/tabwriter"

	"github.com/virajp/bootstrap/cli/internal/config"
	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
)

// setting is the value of one key and its source.
type setting struct {
	key           config.Key
	value, source string // An empty value is the default.
}

// settings are the config file and the value of each key.
type settings struct {
	path, pathSource string
	missing          bool
	keys             []setting
}

// applySettings loads the config file and sets the variable of each key: from
// the flag, else a set variable stays, else from the file. An error of the
// file stops the command with code 2.
func (a *app) applySettings(o opts) fail.Code {
	s := &a.settings
	s.path, s.pathSource = config.Path(o.values["config"], a.getenv)
	if o.values["config"] != "" {
		// A child bootstrap reads the same file.
		_ = os.Setenv("BOOTSTRAP_CONFIG", s.path)
	}
	f, err := config.Load(s.path, a.getenv("HOME"))
	if err != nil {
		a.errorf("%s", err)
		return fail.Usage
	}
	s.missing = f.Missing
	s.keys = nil
	for _, k := range config.Keys {
		v := setting{key: k, source: "default"}
		if fv := o.values[k.Flag]; fv != "" {
			v.value, v.source = fv, "flag --"+k.Flag
			_ = os.Setenv(k.Var, fv)
		} else if ev := a.getenv(k.Var); ev != "" {
			v.value, v.source = ev, "variable "+k.Var
		} else if fv, ok := f.Value(k.Name); ok {
			v.value, v.source = fv, "file"
			_ = os.Setenv(k.Var, fv)
		}
		s.keys = append(s.keys, v)
	}
	if o.verbose {
		_ = os.Setenv("BOOTSTRAP_VERBOSE", "1")
	}
	return fail.OK
}

// config is `bootstrap config`: it prints the config file and each key, with
// its value and its source. It changes nothing. Run gives it <o>, as each
// other command; applySettings already recorded the values of the flags of
// <o> in a.settings.
func (a *app) config(o opts) fail.Code {
	s := a.settings
	w := tabwriter.NewWriter(a.stdout, 0, 0, 2, ' ', 0)
	path := s.path
	if s.missing {
		path += " (missing)"
	}
	_, _ = fmt.Fprintf(w, "file\t%s\t%s\n", path, s.pathSource)
	for _, v := range s.keys {
		value := v.value
		if value == "" {
			value = a.defaultOf(v.key.Name)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", v.key.Name, value, v.source)
	}
	_ = w.Flush()
	return fail.OK
}

// defaultOf gives the default value of the key <name>, or "-" for no value.
func (a *app) defaultOf(name string) string {
	switch name {
	case "jobs":
		// The default of the script runner: internal/scripts/scripts.go (Run).
		return strconv.Itoa(2 * runtime.NumCPU())
	case "cache_dir":
		// The default of core.OpenClone in internal/core/clone.go.
		xdg := a.getenv("XDG_CACHE_HOME")
		if xdg == "" {
			xdg = a.getenv("HOME") + "/.cache"
		}
		return xdg + "/bootstrap"
	case "url":
		return core.DefaultURL
	}
	return "-"
}
