// Package source reads the clone copy of bootstrap: the packs file, the
// tools.toml of each tool and the files of each tool folder. The root of the
// clone is a parameter. Contract: docs/design/sync-engine.md, packs.md and
// tools.md.
package source

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/virajp/bootstrap/sync-engine/internal/fail"
)

// Side is the side of the sync: the repo side or the Mac side.
type Side string

// The two sides.
const (
	Repo  Side = "repo"
	Macos Side = "macos"
)

// EnvNames are the mise environments that the repo side knows.
var EnvNames = []string{"dev", "ci"}

// Source is the clone copy of bootstrap, seen from one side.
type Source struct {
	Root string // The root of the clone (tests use the working tree).
	Side Side
}

// ToolsShown gives the folder of the tools as a message names it.
func (s Source) ToolsShown() string { return string(s.Side) + "/tools" }

// PacksShown gives the packs file as a message names it.
func (s Source) PacksShown() string { return string(s.Side) + "/packs.toml" }

// ToolsDir gives the absolute folder of the tools.
func (s Source) ToolsDir() string { return filepath.Join(s.Root, string(s.Side), "tools") }

// PacksFile gives the absolute path of the packs file.
func (s Source) PacksFile() string { return filepath.Join(s.Root, string(s.Side), "packs.toml") }

// ToolDir gives the absolute folder of <tool>.
func (s Source) ToolDir(tool string) string { return filepath.Join(s.ToolsDir(), tool) }

// ManifestShown gives `<side>/tools/<tool>/tools.toml`, as a message names it.
func (s Source) ManifestShown(tool string) string {
	return s.ToolsShown() + "/" + tool + "/tools.toml"
}

// IsDir tells if <tool> is a folder in the tools folder.
func (s Source) IsDir(tool string) bool {
	st, err := os.Stat(s.ToolDir(tool))
	return err == nil && st.IsDir()
}

// HasManifest tells if <tool> has a tools.toml that is a regular file (a link
// to one counts, as the bash test -f counts it).
func (s Source) HasManifest(tool string) bool {
	st, err := os.Stat(filepath.Join(s.ToolDir(tool), "tools.toml"))
	return err == nil && st.Mode().IsRegular()
}

// ToolFolders gives the names of the tool folders in byte order. A folder whose
// name starts with a dot does not count, as a shell glob does not match it.
func (s Source) ToolFolders() ([]string, error) {
	ents, err := os.ReadDir(s.ToolsDir())
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() {
			names = append(names, e.Name())
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			if st, err := os.Stat(filepath.Join(s.ToolsDir(), e.Name())); err == nil && st.IsDir() {
				names = append(names, e.Name())
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

// readTOML decodes a TOML file into a table with sorted keys.
func readTOML(path string) (*Obj, error) {
	var m map[string]any
	if _, err := toml.DecodeFile(path, &m); err != nil {
		return nil, err
	}
	if m == nil {
		return NewObj(), nil
	}
	return FromMap(m).(*Obj), nil
}

// ReadTOML decodes the TOML file <path> into an object. The keys of each
// table are sorted, as gomplate gives them.
func ReadTOML(path string) (*Obj, error) { return readTOML(path) }

func readErr(format string, args ...any) *fail.Error {
	return fail.New(fail.Usage, format, args...)
}
