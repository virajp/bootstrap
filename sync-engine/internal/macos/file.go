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
	"github.com/virajp/bootstrap/sync-engine/internal/manifest"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/target"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
)

// file is `_ss_file`: it syncs one file of the plan. The steps are the real
// path of the target, the seed gate, `requires`, the unblock, the mode of the
// `.ssh` folder, the render or the copy, the parse check, the merge, and the
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
	if strings.HasPrefix(it.target, ".ssh/") && s.sshFolder() != fail.OK {
		return fail.Other
	}
	// The temp file is next to the real path, so the final rename is atomic.
	tmp, err := s.e.Temps.Create(destdir)
	if err != nil {
		text := "cannot create a temp file in " + destdir
		s.r.Outcome(report.Failed, show, text, text)
		return fail.Other
	}
	rc := fail.OK
	if strings.HasSuffix(it.path, ".tmpl") {
		if !s.renderTemplate(it.tool, src, tmp) {
			s.r.Outcome(report.Failed, show, "gomplate cannot render the template",
				s.src.ToolsShown()+"/"+it.tool+"/"+it.path+": gomplate cannot render the template")
			rc = fail.Usage
		}
	} else if !copyFile(src, tmp) {
		rc = fail.Other
		s.r.Outcome(report.Failed, show, "cannot write the target", "")
	}
	var f format.Format
	if rc == fail.OK {
		f, rc = s.e.Check(tmp, it.target, explicit, it.mode == "merge", show)
	}
	if rc == fail.OK && it.mode == "merge" {
		rc = s.merge(it, tmp, real, f, show)
	}
	if rc != fail.OK {
		s.e.Abort(tmp, real, show)
		return rc
	}

	// The target takes the mode of its source, also when it exists already;
	// an existing merge target keeps its own.
	var mode os.FileMode
	if it.mode == "merge" && exists(real) {
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

// merge runs the merge mode of <it>: the temp file <tmp> over the live target
// <real>, with the `remove` and `remove_items` keys of its [files] table.
func (s *state) merge(it planItem, tmp, real string, f format.Format, show string) fail.Code {
	remove, items := "[]", "[]"
	if fe := s.entry(it.tool, it.target); fe != nil {
		if fe.Keys.Has("remove") {
			remove = source.JSON(fe.Remove)
		}
		if fe.Keys.Has("remove_items") {
			items = source.JSON(fe.RemoveItems)
		}
	}
	ops, err := format.ParseOps(remove, items)
	if err != nil {
		return fail.Usage
	}
	return s.e.Merge(tmp, real, f, ops, show)
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
	env.Set("MACOS_SETUP_CONFIG_PATH", s.getenv("MACOS_SETUP_CONFIG_PATH"))
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

// sshFolder is `_ss_ssh_folder`: it keeps the folder ~/.ssh at mode 700,
// once in each run. The check reads the path in the home folder, not the real
// path of a linked file, and `/.` reads the target when the folder is a link.
// After a failed chmod it stays not done, so each later target below `.ssh/`
// also fails.
func (s *state) sshFolder() fail.Code {
	if s.sshDone {
		return fail.OK
	}
	dir, show := s.home+"/.ssh", tilde+"/.ssh"
	if isDir(dir) {
		if mode, err := target.Mode(dir + "/."); err == nil && mode == 0o700 {
			s.sshDone = true
			return fail.OK
		}
	} else if s.e.DryRun {
		s.sshDone = true
		return fail.OK
	}
	if s.e.DryRun {
		s.sshDone = true
		s.r.Outcome(report.Changed, show, "", "would chmod 700 "+show)
		return fail.OK
	}
	if os.MkdirAll(dir, 0o777) != nil || os.Chmod(dir, 0o700) != nil {
		s.r.Outcome(report.Failed, show, "cannot chmod", "cannot chmod "+show)
		return fail.Other
	}
	s.sshDone = true
	s.r.Outcome(report.Changed, show, "", "chmod 700  "+show)
	s.changed = true
	return fail.OK
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
