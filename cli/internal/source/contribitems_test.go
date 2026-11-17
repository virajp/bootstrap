package source

import (
	"reflect"
	"testing"
)

// ownerTool is the tool b of the contributed-items cases: it ships the
// merge-mode file list.txt and the copy plain.txt.
const ownerTool = `
[files."list.txt"]
mode = "merge"
source = "list.merge.toml"

[files."plain.txt"]
source = "plain.txt"
`

// TestContribItemsRules: the check refuses, with code 2, a contributed item
// for a target that the owner does not declare as merge-mode, and an owner
// that the catalog does not have (D7).
func TestContribItemsRules(t *testing.T) {
	const at = "repo/tools/a/tools.toml: "
	cases := []struct {
		name, a string
		want    []string
	}{
		{"a merge-mode target", "[[contributes_items.b.\"list.txt\".item]]\nvalue = \"x\"\n", nil},
		{"a target that is not merge-mode", "[[contributes_items.b.\"plain.txt\".item]]\nvalue = \"x\"\n",
			[]string{at + `[contributes_items.b."plain.txt"]: b declares no merge-mode file plain.txt - name a [files."<target>"] of b with mode = "merge"`}},
		{"a target that the owner does not have", "[[contributes_items.b.\"nope.txt\".item]]\nvalue = \"x\"\n",
			[]string{at + `[contributes_items.b."nope.txt"]: b declares no merge-mode file nope.txt - name a [files."<target>"] of b with mode = "merge"`}},
		{"an owner that the catalog does not have", "[[contributes_items.nope.\"list.txt\".item]]\nvalue = \"x\"\n",
			[]string{at + "[contributes_items.nope] names a tool that bootstrap does not have"}},
		{"not a table", "contributes_items = 1\n",
			[]string{at + "[contributes_items] must be a table of [contributes_items.<tool>] tables"}},
		{"an item that is not a table", "[contributes_items.b.\"list.txt\"]\nitem = [ \"x\" ]\n",
			[]string{at + `[contributes_items.b."list.txt"] item must be a list of [[contributes_items.b."list.txt".item]] tables`}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, map[string]string{
				"repo/categories.toml":               "[category.core]\ntitle = \"Core\"\n",
				"repo/tools/a/tools.toml":            lspTool(c.a),
				"repo/tools/b/tools.toml":            lspTool(ownerTool),
				"repo/tools/b/files/list.merge.toml": "format = \"text\"\n",
				"repo/tools/b/files/plain.txt":       "x\n",
			})
			_, err := Source{Root: root, Side: Repo}.Open()
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
		})
	}
}

// TestContributedItemsOrder: the items of the given tools in the order of
// the tool names; KnownItems holds the items of each tool of the catalog.
func TestContributedItemsOrder(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"repo/categories.toml":               "[category.core]\ntitle = \"Core\"\n",
		"repo/tools/b/tools.toml":            lspTool(ownerTool),
		"repo/tools/b/files/list.merge.toml": "format = \"text\"\n",
		"repo/tools/b/files/plain.txt":       "x\n",
		"repo/tools/z/tools.toml":            lspTool("[[contributes_items.b.\"list.txt\".item]]\nvalue = \"z1\"\n[[contributes_items.b.\"list.txt\".item]]\nvalue = \"z2\"\n"),
		"repo/tools/c/tools.toml":            lspTool("[[contributes_items.b.\"list.txt\".item]]\nvalue = \"c1\"\n"),
	})
	x, err := Source{Root: root, Side: Repo}.Open()
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	values := func(l []any) []string {
		var out []string
		for _, e := range l {
			out = append(out, e.(*Obj).Get("value").(string))
		}
		return out
	}
	if got := values(x.ContributedItems("b", "list.txt", []string{"z", "c", "b"})); !reflect.DeepEqual(got, []string{"c1", "z1", "z2"}) {
		t.Errorf("ContributedItems = %q", got)
	}
	if got := values(x.ContributedItems("b", "list.txt", []string{"b"})); got != nil {
		t.Errorf("ContributedItems of the owner alone = %q", got)
	}
	if got := values(x.KnownItems("b", "list.txt")); !reflect.DeepEqual(got, []string{"c1", "z1", "z2"}) {
		t.Errorf("KnownItems = %q", got)
	}
}
