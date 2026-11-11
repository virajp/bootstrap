package config

import (
	"os"
	"strings"
	"testing"
)

func TestPath(t *testing.T) {
	cases := []struct {
		flag         string
		env          map[string]string
		path, source string
	}{
		{"/f.toml", map[string]string{"BOOTSTRAP_CONFIG": "/v.toml", "XDG_CONFIG_HOME": "/x", "HOME": "/h"}, "/f.toml", "flag --config"},
		{"", map[string]string{"BOOTSTRAP_CONFIG": "/v.toml", "XDG_CONFIG_HOME": "/x", "HOME": "/h"}, "/v.toml", "variable BOOTSTRAP_CONFIG"},
		{"", map[string]string{"XDG_CONFIG_HOME": "/x", "HOME": "/h"}, "/x/bootstrap/config.toml", "default"},
		{"", map[string]string{"XDG_CONFIG_HOME": "", "HOME": "/h"}, "/h/.config/bootstrap/config.toml", "default"},
	}
	for _, c := range cases {
		path, source := Path(c.flag, func(k string) string { return c.env[k] })
		if path != c.path || source != c.source {
			t.Errorf("Path(%q, %v) = %q, %q; want %q, %q", c.flag, c.env, path, source, c.path, c.source)
		}
	}
}

func TestLoadMissing(t *testing.T) {
	f, err := Load(t.TempDir()+"/none/config.toml", "/h")
	if err != nil || !f.Missing {
		t.Fatalf("a missing file: %v, missing %v", err, f.Missing)
	}
	if _, ok := f.Value("jobs"); ok {
		t.Error("a missing file gives a value")
	}
}

func TestLoadFolder(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir, "/h"); err == nil || !strings.Contains(err.Error(), dir+": ") {
		t.Errorf("a folder: %v", err)
	}
}

func TestLoad(t *testing.T) {
	cases := []struct {
		name, body string
		want       map[string]string
		err        string // The text after "<file>: ".
	}{
		{"each key", "jobs = 4\ncache_dir = \"/c\"\nurl = \"https://example.com/b.git\"\nclone = \"/k\"\n[mac]\n",
			map[string]string{"jobs": "4", "cache_dir": "/c", "url": "https://example.com/b.git", "clone": "/k"}, ""},
		{"tilde expanded", "cache_dir = \"~/x\"\nclone = \"~/y/z\"\n", map[string]string{"cache_dir": "/h/x", "clone": "/h/y/z"}, ""},
		{"empty file", "", map[string]string{}, ""},
		{"verbose", "verbose = true\n", nil, "verbose: unknown key"},
		{"mac.verbose", "[mac]\nverbose = true\n", nil, "mac.verbose: unknown key"},
		{"mac.repo_path", "[mac]\nrepo_path = \"/r\"\n", nil, "mac.repo_path: unknown key"},
		{"other table", "[other]\na = 1\n", nil, "other: unknown table"},
		{"mac not a table", "mac = 1\n", nil, "mac: needs a table, not an integer"},
		{"jobs a string", "jobs = \"4\"\n", nil, "jobs: needs a number above 0, not a string"},
		{"jobs 0", "jobs = 0\n", nil, "jobs: needs a number above 0, not 0"},
		{"url an integer", "url = 3\n", nil, "url: needs a string, not an integer"},
		{"jobs a float", "jobs = 1.5\n", nil, "jobs: needs a number above 0, not a float"},
		{"clone a date", "clone = 1979-05-27\n", nil, "clone: needs a string, not a date"},
		{"mac an array of tables", "[[mac]]\n", nil, "mac: needs a table, not an array"},
		{"empty url", "url = \"\"\n", nil, "url: is empty"},
		{"relative cache_dir", "cache_dir = \"x\"\n", nil, "cache_dir: needs an absolute path or a path that starts with ~/, not 'x'"},
		{"relative clone", "clone = \"~x\"\n", nil, "clone: needs an absolute path or a path that starts with ~/, not '~x'"},
		{"not TOML", "jobs = \n", nil, "not valid TOML: line 1"},
	}
	dir := t.TempDir()
	for _, c := range cases {
		path := dir + "/" + strings.ReplaceAll(c.name, " ", "-") + ".toml"
		if err := os.WriteFile(path, []byte(c.body), 0o644); err != nil {
			t.Fatal(err)
		}
		f, err := Load(path, "/h")
		if c.err != "" {
			if err == nil || !strings.HasPrefix(err.Error(), path+": "+c.err) {
				t.Errorf("%s: error %v, want %q", c.name, err, path+": "+c.err)
			}
			continue
		}
		if err != nil || f.Missing {
			t.Errorf("%s: %v, missing %v", c.name, err, f.Missing)
			continue
		}
		for _, k := range Keys {
			got, ok := f.Value(k.Name)
			want, wantOK := c.want[k.Name]
			if got != want || ok != wantOK {
				t.Errorf("%s: %s = %q (%v), want %q (%v)", c.name, k.Name, got, ok, want, wantOK)
			}
		}
	}
}
