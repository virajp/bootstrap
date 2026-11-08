package source

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// File is one file that a tool ships.
type File struct {
	Path   string // Relative to the tool folder, with `.tmpl` when a template.
	Target string // Path without `.tmpl`.
	Tmpl   bool   // The file is a template.
}

// Files gives the files that <tool> ships, in byte order of the path. Only
// regular files count; the walk does not follow a link, as `find -type f`
// does not. The tools.toml is not a shipped file: the repo side leaves out each
// file with that name (`! -name tools.toml`), and the Mac side leaves out only
// the one at the top of the tool folder (`! -path ./tools.toml`).
func (s Source) Files(tool string) ([]File, error) {
	root := s.ToolDir(tool)
	var out []File
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if s.Side == Macos {
			if rel == "tools.toml" {
				return nil
			}
		} else if d.Name() == "tools.toml" {
			return nil
		}
		f := File{Path: rel, Target: strings.TrimSuffix(rel, ".tmpl")}
		f.Tmpl = f.Target != rel
		out = append(out, f)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}
