package macos

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/header"
	"github.com/virajp/bootstrap/sync-engine/internal/manifest"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/target"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
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
	show := tilde + "/" + it.target
	s.show = show
	real, ok := manifest.Real(s.home + "/" + it.target)
	if !ok {
		s.r.Outcome(report.Failed, show, "cannot resolve - a symlink loop", "cannot resolve "+show+" - a symlink loop")
		return fail.Other
	}
	if real != s.homeReal+"/"+it.target {
		show += " (-> " + real + ")"
	}
	s.show = show

	if it.mode == "seed" && target.SeedKept(real) {
		s.r.Outcome(report.Kept, show, "", "kept seed  "+show)
		return fail.OK
	}
	if fe := s.entry(it.tool, it.target); fe != nil {
		for _, req := range fe.Requires {
			if exists(s.home + "/" + req) {
				continue
			}
			switch it.mode {
			case "seed":
				s.r.Outcome(report.Skipped, show, "", "skip seed  "+show+" (no "+tilde+"/"+req+")")
			case "merge":
				s.r.Outcome(report.Skipped, show, "", "skip merge "+show+" (no "+tilde+"/"+req+")")
			default:
				s.r.Outcome(report.Skipped, show, "", "skip file  "+show+" (no "+tilde+"/"+req+")")
				s.unhandle(it.target)
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
	// can be outside $HOME, so the walk then stops at the first folder of
	// that path.
	top := s.homeReal
	if !strings.HasPrefix(real, s.homeReal+"/") {
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
	fe := s.entry(it.tool, it.target)
	if fe != nil && fe.DirPerm != 0 {
		dir := target.Dirname(it.target)
		changed, code := s.e.DirPerm(s.home+"/"+dir, tilde+"/"+dir, fe.DirPerm, s.dirDone)
		if code != fail.OK {
			return code
		}
		if changed {
			s.changed = true
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
	if !s.e.DryRun && !s.inSync(tmp, real, mode) {
		s.changed = true
	}
	if code := s.e.Finish(mode, tmp, real, show, it.mode); code != fail.OK {
		return code
	}
	if it.mode == "generated" {
		s.rows = append(s.rows, manifest.Row{Tool: it.tool, Path: it.target})
	}
	if it.mode == "merge" {
		s.merged[it.target] = true
		for _, item := range items {
			s.rows = append(s.rows, manifest.Row{Tool: it.tool, Path: it.target, Item: item})
		}
	}
	return fail.OK
}

// inSync is true when <real> is a regular file with the content of <tmp> and
// the mode <mode>.
func (s *state) inSync(tmp, real string, mode os.FileMode) bool {
	info, err := os.Stat(real)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	a, err := os.ReadFile(tmp)
	if err != nil {
		return false
	}
	b, err := os.ReadFile(real)
	if err != nil || string(a) != string(b) {
		return false
	}
	cur, err := target.Mode(real)
	return err == nil && cur == mode
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
// .sync_engine and .env, as the ctx/<tool>.json file of bash.
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
	env.Set("MACOS_SETUP_REPO_PATH", s.getenv("MACOS_SETUP_REPO_PATH"))
	o := source.NewObj()
	o.Set("config", s.all.Get(tool))
	o.Set("all", s.all)
	o.Set("tools", names)
	o.Set("sync_engine", s.dir+"/sync-engine")
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
	fmt.Fprintf(s.stderr, `{"time":"%s","level":"ERROR","msg":"","err":%s}`+"\n",
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
