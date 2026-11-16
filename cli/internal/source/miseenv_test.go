package source

import (
	"reflect"
	"testing"
)

// TestMiseEnvRequiresFound: a [mise_env.<env>.tools] entry of a repo tool can
// have requires_found, with the shape rules of a [mise.tools] entry; requires
// and files stay refused there (D40).
func TestMiseEnvRequiresFound(t *testing.T) {
	cases := []struct {
		name, body string
		want       []string
		env        []MiseEnvTool
	}{
		{"accepted", "[mise_env.dev.tools.da]\nversion = \"1\"\nrequires_found = [ \"package.json\", \"tsconfig.json\" ]\n", nil,
			[]MiseEnvTool{{Env: "dev", Name: "da", RequiresFound: []string{"package.json", "tsconfig.json"}}}},
		{"without the key", "[mise_env.dev.tools.da]\nversion = \"1\"\n", nil, nil},
		{"requires", "[mise_env.dev.tools.da]\nversion = \"1\"\nrequires = \"x\"\n",
			[]string{"repo/tools/a/tools.toml: requires of mise tool da in [mise_env.dev.tools] is valid only in [mise.tools]"}, nil},
		{"files", "[mise_env.dev.tools.da]\nversion = \"1\"\nfiles = [ \"x\" ]\n",
			[]string{"repo/tools/a/tools.toml: files of mise tool da in [mise_env.dev.tools] is valid only in [mise.tools]"}, nil},
		{"empty list", "[mise_env.dev.tools.da]\nversion = \"1\"\nrequires_found = []\n",
			[]string{"repo/tools/a/tools.toml: requires_found of mise tool da in [mise_env.dev.tools] must be a non-empty list of file names"}, nil},
		{"a path", "[mise_env.dev.tools.da]\nversion = \"1\"\nrequires_found = [ \"web/package.json\" ]\n",
			[]string{"repo/tools/a/tools.toml: requires_found of mise tool da: 'web/package.json' must be a file name, with no /"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, map[string]string{
				"repo/categories.toml":    "[category.core]\ntitle = \"Core\"\n",
				"repo/tools/a/tools.toml": lspTool(c.body),
			})
			x, err := Source{Root: root, Side: Repo}.Open()
			var got []string
			for _, e := range Messages(err) {
				got = append(got, e.Msg)
				if e.Code != 2 {
					t.Errorf("code %d for %s", e.Code, e.Msg)
				}
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("messages:\n%q\nwant\n%q", got, c.want)
			}
			if c.want == nil && !reflect.DeepEqual(x.Manifests["a"].MiseEnv, c.env) {
				t.Errorf("MiseEnv = %+v, want %+v", x.Manifests["a"].MiseEnv, c.env)
			}
		})
	}
}
