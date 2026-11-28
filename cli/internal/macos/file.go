package macos

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	"github.com/virajp/bootstrap/cli/internal/header"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/tmpl"
)

// file is `_ss_file`: it syncs one file of the plan. The steps are the real
// path of the target, the seed gate, `requires`, the unblock, the
// `dir_perm` of the folder, the render or the copy, the parse check, the merge, and the
// compare and the write. A generated file with a missing `requires` leaves
// the handled targets, so the stale rule removes a listed copy. It sets the
// change flag for a change on disk, and adds the manifest row of a generated
// file.
func (s *state) file(it planItem) fail.Code {
	src := it.path
	if !strings.HasPrefix(src, "/") {
		src = s.src.ToolDir(it.tool) + "/" + it.path
	}
	explicit := it.format
	if explicit == "-" {
		explicit = ""
	}
	show := s.showTarget(it.target, it.backup)
	s.show = show
	// The root of the target: $HOME, or the backup clone.
	root, rootReal := s.home, s.homeReal
	if it.backup {
		if s.clone == "" {
			s.r.Outcome(report.Skipped, show, "", "skip file  "+show+" (no backup in "+valuesName+")")
			if it.mode != "seed" {
				s.unhandle(it.row())
			}
			return fail.OK
		}
		root, rootReal = s.clone, s.cloneReal
		if rootReal == "" {
			rootReal = root
		}
	}
	abs := TargetPath(s.home, s.clone, it.target, it.backup)
	real, ok := manifest.Real(abs)
	if !ok {
		s.r.Outcome(report.Failed, show, "cannot resolve - a symlink loop", "cannot resolve "+show+" - a symlink loop")
		return fail.Other
	}
	if real != rootReal+"/"+it.target {
		show += " (-> " + real + ")"
	}
	s.show = show

	if it.mode == "seed" && target.SeedKept(real) {
		s.r.Outcome(report.Kept, show, "", "kept seed  "+show)
		return fail.OK
	}
	if fe := s.entry(it.tool, it.row()); fe != nil {
		// A requires path of a target in the backup clone is relative to
		// the clone.
		for _, req := range fe.Requires {
			if exists(root + "/" + req) {
				continue
			}
			need := s.showPath(root + "/" + req)
			switch it.mode {
			case "seed":
				s.r.Outcome(report.Skipped, show, "", "skip seed  "+show+" (no "+need+")")
			case "merge":
				s.r.Outcome(report.Skipped, show, "", "skip merge "+show+" (no "+need+")")
			default:
				s.r.Outcome(report.Skipped, show, "", "skip file  "+show+" (no "+need+")")
				s.unhandle(it.row())
			}
			return fail.OK
		}
	}
	if isDir(real) {
		text := "is a folder, but a file goes there - move it"
		s.r.Outcome(report.Failed, show, text, show+" "+text)
		return fail.Other
	}
	// A file with the header where a folder goes is removed. The real path
	// can be outside the root ($HOME or the backup clone), so the walk then
	// stops at the first folder of that path. The walk never goes above the
	// root.
	top := rootReal
	if !strings.HasPrefix(real, rootReal+"/") {
		first, _, _ := strings.Cut(strings.TrimPrefix(real, "/"), "/")
		top = "/" + first
	}
	if s.e.Unblock(target.Dirname(real), top, nil, show) != fail.OK {
		return fail.Other
	}

	var destdir string
	if s.e.DryRun {
		destdir = target.DryRunDir(s.getenv)
	} else {
		destdir = target.Dirname(real)
		if err := os.MkdirAll(destdir, 0o777); err != nil {
			s.r.Outcome(report.Failed, show, "cannot create "+destdir, "cannot create "+destdir)
			return fail.Other
		}
	}
	fe := s.entry(it.tool, it.row())
	if fe != nil && fe.DirPerm != 0 {
		dir := target.Dirname(it.target)
		if _, code := s.e.DirPerm(root+"/"+dir, s.showPath(root+"/"+dir), fe.DirPerm, s.dirDone); code != fail.OK {
			return code
		}
	}
	// The temp file is next to the real path, so the final rename is atomic.
	tmp, err := s.e.Temps.Create(destdir)
	if err != nil {
		text := "cannot create a temp file in " + destdir
		s.r.Outcome(report.Failed, show, text, text)
		return fail.Other
	}
	rc := fail.OK
	if it.tmpl {
		if !s.renderTemplate(it.tool, src, tmp) {
			s.r.Outcome(report.Failed, show, "gomplate cannot render the template",
				s.src.ToolsShown()+"/"+it.tool+"/"+it.path+": gomplate cannot render the template")
			rc = fail.Usage
		}
	} else if !copyFile(src, tmp) {
		rc = fail.Other
		s.r.Outcome(report.Failed, show, "cannot write the target", "")
	}
	if rc == fail.OK && it.mode != "merge" {
		_, rc = s.e.Check(tmp, it.target, explicit, false, show)
	}
	var items []string
	if rc == fail.OK && it.mode == "merge" {
		items, rc = s.merge(it, tmp, real, explicit, show)
	}
	if rc != fail.OK {
		s.e.Abort(tmp, real, show)
		return rc
	}

	// The target takes its `perm`, else the mode of its source, also when it
	// exists already; an existing merge target without `perm` keeps its own.
	var mode os.FileMode
	if fe != nil && fe.Perm != 0 {
		mode = fe.Perm
	} else if it.mode == "merge" && exists(real) {
		mode, _ = target.Mode(real)
	} else {
		mode, _ = target.Mode(src)
	}
	if code := s.e.Finish(mode, tmp, real, show, it.mode); code != fail.OK {
		return code
	}
	if it.mode == "generated" {
		s.rows = append(s.rows, manifest.Row{Tool: it.tool, Path: it.row()})
	}
	if it.mode == "merge" {
		s.merged[it.target] = true
		for _, item := range items {
			s.rows = append(s.rows, manifest.Row{Tool: it.tool, Path: it.target, Item: item})
		}
	}
	return fail.OK
}

// merge merges the shared file of <it>: the rendered spec in the temp file
// <tmp> into the live target <real>. It gives the item rows of the manifest.
func (s *state) merge(it planItem, tmp, real, explicit, show string) ([]string, fail.Code) {
	var ov []byte
	if v, ok := s.all.Get(it.tool).(*source.Obj); ok {
		if files, ok := v.Get("files").(*source.Obj); ok && files.Has(it.target) {
			ov = []byte(source.JSON(files.Get(it.target)))
		}
	}
	name := s.src.ToolsShown() + "/" + it.tool + "/" + it.path
	return s.e.Shared(tmp, real, format.SharedInput{
		Target:    it.target,
		Explicit:  explicit,
		Overrides: ov,
		Records:   manifest.Items(s.oldRows)[it.target],
		Header:    header.Shared(name, valuesName),
		Show:      show,
		SpecShow:  name,
	})
}

// context gives the template context of <tool>: .config, .all, .tools,
// .cli_dir and .env, as the ctx/<tool>.json file of bash.
func (s *state) context(tool string) (any, error) {
	if c, ok := s.ctx[tool]; ok {
		return c, nil
	}
	names := make([]any, len(s.res.Tools))
	for i, t := range s.res.Tools {
		names[i] = t
	}
	env := source.NewObj()
	env.Set("HOME", s.getenv("HOME"))
	env.Set("USER", s.getenv("USER"))
	o := source.NewObj()
	o.Set("config", s.all.Get(tool))
	o.Set("all", s.all)
	o.Set("tools", names)
	o.Set("cli_dir", s.dir+"/"+source.CLIDir)
	o.Set("env", env)
	c, err := tmpl.DecodeJSON([]byte(source.JSON(o)))
	if err != nil {
		return nil, err
	}
	s.ctx[tool] = c
	return c, nil
}

// renderTemplate renders the template <src> of <tool> into <tmp>. On a
// failure it prints the error to stderr, in the form of the log line of
// gomplate, and gives false.
func (s *state) renderTemplate(tool, src, tmp string) bool {
	text, err := os.ReadFile(src)
	if err != nil {
		s.gomplateError("renderTemplate: failed to render template " + src + ": " + err.Error())
		return false
	}
	ctx, err := s.context(tool)
	if err != nil {
		s.gomplateError("renderTemplate: failed to render template " + src + ": " + err.Error())
		return false
	}
	out, err := tmpl.Render(src, text, ctx, nil)
	if err != nil {
		var ee template.ExecError
		if !errors.As(err, &ee) && strings.HasPrefix(err.Error(), "template: ") {
			s.gomplateError("renderTemplate: parse template " + src + ": " + err.Error())
		} else {
			s.gomplateError("renderTemplate: failed to render template " + src + ": " + err.Error())
		}
		return false
	}
	return os.WriteFile(tmp, out, 0o600) == nil
}

// gomplateError prints <msg> as the JSON log line of a failed gomplate.
func (s *state) gomplateError(msg string) {
	_, _ = fmt.Fprintf(s.stderr, `{"time":"%s","level":"ERROR","msg":"","err":%s}`+"\n",
		time.Now().Format("2006-01-02T15:04:05.999999Z07:00"), source.JSON(msg))
}

// copyFile is `cp -- <src> <dst>` into an existing <dst>: the content only.
func copyFile(src, dst string) bool {
	b, err := os.ReadFile(src)
	if err != nil {
		return false
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644)
	if err != nil {
		return false
	}
	_, err = f.Write(b)
	return f.Close() == nil && err == nil
}
