package core

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/source"
)

func TestParseLines(t *testing.T) {
	got := parseLines("bat\n\n  jq \nowner/tap/widget\n", prefixFormula)
	want := []string{"brew:bat", "brew:jq", "brew:owner/tap/widget"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseLines = %v, want %v", got, want)
	}
}

func TestParseAppStore(t *testing.T) {
	out := "1569813296  1Password for Safari  (8.12.44)\n" +
		" 497799835  Xcode                 (16.0)\n" +
		"No installed apps found\n\n"
	got := parseAppStore(out)
	want := []string{"mas:1569813296", "mas:497799835"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseAppStore = %v, want %v", got, want)
	}
}

func TestParseMiseTools(t *testing.T) {
	got, err := parseMiseTools([]byte(`{"uv": [{"version": "0.9"}], "1password": [], "github:virajp/bootstrap": []}`))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"1password", "github:virajp/bootstrap", "uv"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseMiseTools = %v, want %v", got, want)
	}
	if _, err := parseMiseTools([]byte("not json")); err == nil {
		t.Error("parseMiseTools accepts text that is not JSON")
	}
}

func TestPackageKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tools.toml")
	text := `description = "x"
summary = "x"
category = "core"

[mise.bootstrap.packages]
"brew:bat" = "latest"
"brew-cask:vlc" = "latest"

[mise.tools.uv]
version = "latest"
`
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := source.ReadTOML(path)
	if err != nil {
		t.Fatal(err)
	}
	got := packageKeys(raw)
	want := []Package{{kindPackage, "brew-cask:vlc"}, {kindPackage, "brew:bat"}, {kindTool, "uv"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("packageKeys = %v, want %v", got, want)
	}
	if got := packageKeys(source.NewObj()); got != nil {
		t.Errorf("packageKeys of a file without [mise] = %v, want nil", got)
	}
}

func TestPkgExpr(t *testing.T) {
	cases := []struct {
		p           Package
		expr, table string
	}{
		{Package{kindPackage, "brew-cask:htop"}, `.mise.bootstrap.packages["brew-cask:htop"] = "latest"`, "[mise.bootstrap.packages]"},
		{Package{kindTool, "npm:codeburn"}, `.mise.tools["npm:codeburn"].version = "latest"`, "[mise.tools]"},
	}
	for _, c := range cases {
		if got := c.p.expr(); got != c.expr {
			t.Errorf("%v.expr() = %s, want %s", c.p, got, c.expr)
		}
		if got := c.p.Table(); got != c.table {
			t.Errorf("%v.table() = %s, want %s", c.p, got, c.table)
		}
	}
}

func TestPkgMatch(t *testing.T) {
	cases := []struct{ in, want Package }{
		{Package{kindPackage, "brew-cask:stablyai/orca/orca"}, Package{kindPackage, "brew-cask:orca"}},
		{Package{kindPackage, "brew:virajp/tap/claude-status"}, Package{kindPackage, "brew:claude-status"}},
		{Package{kindPackage, "brew:bat"}, Package{kindPackage, "brew:bat"}},
		{Package{kindPackage, "mas:497799835"}, Package{kindPackage, "mas:497799835"}},
		{Package{kindTool, "github:virajp/bootstrap"}, Package{kindTool, "github:virajp/bootstrap"}},
	}
	for _, c := range cases {
		if got := c.in.match(); got != c.want {
			t.Errorf("%v.match() = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestKindOf(t *testing.T) {
	for _, c := range []struct {
		body string
		want AppKind
	}{
		{"", ConfigOnly},
		{"[mise.bootstrap.packages]\n\"brew:bat\" = \"latest\"\n", TerminalApp},
		{"[mise.tools.uv]\nversion = \"latest\"\n", TerminalApp},
		{"[mise.bootstrap.packages]\n\"brew-cask:vlc\" = \"latest\"\n", DesktopApp},
		{"[mise.bootstrap.packages]\n\"mas:409183694\" = \"latest\"\n", DesktopApp},
		{"[mise.bootstrap.packages]\n\"brew:x\" = \"latest\"\n\"brew-cask:y\" = \"latest\"\n", DesktopApp},
	} {
		path := filepath.Join(t.TempDir(), "tools.toml")
		if err := os.WriteFile(path, []byte("summary = \"x\"\n"+c.body), 0o644); err != nil {
			t.Fatal(err)
		}
		raw, err := source.ReadTOML(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := KindOf(raw); got != c.want {
			t.Errorf("KindOf(%q) = %v, want %v", c.body, got, c.want)
		}
	}
}
