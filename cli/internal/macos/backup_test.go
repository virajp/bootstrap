package macos

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLinkPlan(t *testing.T) {
	targets := []string{
		".zshrc", ".config/mise/config.toml", ".config/mise/conf.d/a/mise.toml",
		".config/fish/conf.d/01-env.fish", ".config/fish/functions/fish_title.fish", ".config/fish/config.fish",
		".claude/CLAUDE.md", ".claude/settings.json", ".claude/output-styles/STE100.md",
		".ssh/config", ".config/starship.toml", ".config/bootstrap/macos.toml",
	}
	runtime := []string{".config/fish/fish_variables", ".claude/projects", ".ssh/known_hosts"}
	var got []string
	for _, l := range linkPlan(targets, runtime) {
		kind := "file"
		if l.folder {
			kind = "folder"
		}
		got = append(got, l.rel+" "+kind)
	}
	want := []string{
		".claude/CLAUDE.md file", ".claude/output-styles folder", ".claude/settings.json file",
		".config/bootstrap folder", ".config/fish/conf.d folder", ".config/fish/config.fish file",
		".config/fish/functions folder", ".config/mise folder", ".config/starship.toml file",
		".ssh/config file", ".zshrc file",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("links:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// tree writes each file of <files> (path to content) below <root>, and each
// link of <links> (path to target).
func tree(t *testing.T, root string, files, links map[string]string) {
	t.Helper()
	for p, c := range files {
		if err := os.MkdirAll(filepath.Dir(root+"/"+p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(root+"/"+p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for p, to := range links {
		if err := os.MkdirAll(filepath.Dir(root+"/"+p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(to, root+"/"+p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAdoption(t *testing.T) {
	show := func(p string) string { return p }
	run := func(t *testing.T, home, root string, l link) *adoption {
		a := &adoption{}
		a.plan(l, home, root, show)
		return a
	}
	kinds := func(a *adoption) string {
		var s []string
		for _, st := range a.steps {
			s = append(s, []string{"move", "remove", "mkdir", "link"}[st.kind])
		}
		return strings.Join(s, " ")
	}
	cases := []struct {
		name                   string
		home, clone, homeLinks map[string]string
		l                      link
		steps, conflict        string
	}{
		{name: "a real file, no file in the clone", home: map[string]string{".zprofile": "x"},
			l: link{rel: ".zprofile"}, steps: "move link"},
		{name: "a real file the same as the clone", home: map[string]string{".zprofile": "x"}, clone: map[string]string{".zprofile": "x"},
			l: link{rel: ".zprofile"}, steps: "remove link"},
		{name: "a real file that differs", home: map[string]string{".zprofile": "x"}, clone: map[string]string{".zprofile": "y"},
			l: link{rel: ".zprofile"}, conflict: "differ"},
		{name: "nothing yet, a file", l: link{rel: ".zshrc"}, steps: "link"},
		{name: "nothing yet, a folder", l: link{rel: ".config/git", folder: true}, steps: "mkdir link"},
		{name: "a real folder with links into the clone and new files",
			home: map[string]string{".config/git/identity": "i"}, clone: map[string]string{".config/git/config": "c"},
			homeLinks: map[string]string{".config/git/config": "CLONE/.config/git/config"},
			l:         link{rel: ".config/git", folder: true}, steps: "remove move remove link"},
		{name: "a file link where a folder link goes", clone: map[string]string{".config/ghostty/config": "g"},
			homeLinks: map[string]string{".config/ghostty/config": "CLONE/.config/ghostty/config"},
			l:         link{rel: ".config/ghostty", folder: true}, steps: "remove remove link"},
		{name: "a link elsewhere", homeLinks: map[string]string{".zshrc": "/elsewhere"},
			l: link{rel: ".zshrc"}, conflict: "is a link to /elsewhere"},
		{name: "the correct link", clone: map[string]string{".zshrc": "z"}, homeLinks: map[string]string{".zshrc": "CLONE/.zshrc"},
			l: link{rel: ".zshrc"}, steps: ""},
		{name: "a link above the link", clone: map[string]string{".config/fish/conf.d/a": "a"},
			homeLinks: map[string]string{".config/fish": "CLONE/.config/fish"},
			l:         link{rel: ".config/fish/conf.d", folder: true}, conflict: "replace the link with a folder"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			home, root := t.TempDir(), t.TempDir()
			links := map[string]string{}
			for p, to := range c.homeLinks {
				links[p] = strings.Replace(to, "CLONE", root, 1)
			}
			tree(t, home, c.home, links)
			tree(t, root, c.clone, nil)
			a := run(t, home, root, c.l)
			if c.conflict != "" {
				if len(a.conflicts) != 1 || !strings.Contains(a.conflicts[0], c.conflict) {
					t.Errorf("conflicts = %v, want %q", a.conflicts, c.conflict)
				}
				return
			}
			if len(a.conflicts) > 0 || kinds(a) != c.steps {
				t.Errorf("steps = %q, conflicts = %v, want %q", kinds(a), a.conflicts, c.steps)
			}
		})
	}
}
