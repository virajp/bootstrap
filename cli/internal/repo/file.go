package repo

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/format"
	hdr "github.com/virajp/bootstrap/cli/internal/header"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/tmpl"
)

// fileMode is `_rs_file_mode`: the mode of the target <tool>/<target>;
// generated without a `mode` key.
func (s *state) fileMode(key string) string {
	if m, ok := s.fileModes[key]; ok {
		return m
	}
	return "generated"
}

// spared is `_rs_spared`: <path> is a seed or a merge target of a tool of
// this run.
func (s *state) spared(path string) bool {
	for _, p := range s.spare {
		if p == path {
			return true
		}
	}
	return false
}

// remove is `_rs_remove`: it removes the file <path> of the tool from this
// repo, if it exists.
func (s *state) remove(path string) fail.Code {
	if s.spared(path) {
		return fail.OK
	}
	dest := s.root + "/" + path
	show := s.path(dest)
	if !exists(dest) {
		return fail.OK
	}
	if s.e.DryRun {
		s.e.Plan(dest, report.Removed)
		s.r.Outcome(report.Removed, show, "", "would remove "+show)
		return fail.OK
	}
	// rm -f: a folder is not removed.
	info, err := os.Lstat(dest)
	if err == nil && info.IsDir() {
		err = errors.New("is a directory")
	} else if err == nil {
		if err = os.Remove(dest); os.IsNotExist(err) {
			err = nil
		}
	}
	if err != nil {
		s.r.Outcome(report.Failed, show, "cannot remove", "cannot remove "+show)
		return fail.Other
	}
	s.r.Outcome(report.Removed, show, "", "removed    "+show)
	return fail.OK
}

// skipRequired is `_rs_skip_required`: the target of the file <path> of
// <tool> is not written, because an item of its `requires` or
// `requires_found` is not there. A seed that exists stays, as with no
// condition. A seed or a merge target stays as it is. A generated target is
// only logged: it gets no manifest row, so the stale rule removes a listed
// copy with the header.
func (s *state) skipRequired(tool, out string) {
	requires := s.reqSkip[tool+"/"+out]
	dest := s.root + "/" + out
	show := s.path(dest)
	switch s.fileMode(tool + "/" + out) {
	case "seed":
		s.handled = append(s.handled, out)
		if target.SeedKept(dest) {
			s.r.Outcome(report.Kept, show, "", "kept seed  "+show)
		} else {
			s.r.Outcome(report.Skipped, show, "", "skip seed  "+show+" (no "+requires+")")
		}
		return
	case "merge":
		s.handled = append(s.handled, out)
		s.r.Outcome(report.Skipped, show, "", "skip merge "+show+" (no "+requires+")")
		return
	}
	s.r.Outcome(report.Skipped, show, "", "skip file  "+show+" (no "+requires+")")
}

// envOf gives the environment of a mise file of a tool,
// `.config/mise/conf.d/*/mise.*.toml` as a bash pattern (a `*` matches a `/`
// too), and true. It gives false for another path.
func envOf(out string) (string, bool) {
	const prefix = ".config/mise/conf.d/"
	rest, ok := strings.CutPrefix(out, prefix)
	if !ok || !strings.HasSuffix(rest, ".toml") {
		return "", false
	}
	match := false
	for i := 0; i+6 <= len(rest); i++ {
		if rest[i:i+6] == "/mise." && strings.HasSuffix(rest[i+6:], ".toml") {
			match = true
			break
		}
	}
	if !match {
		return "", false
	}
	env := out[strings.LastIndex(out, "/mise.")+6:]
	return strings.TrimSuffix(env, ".toml"), true
}

// file is `_rs_file`: the first step of the sync of the file <f> of
// <tool>. It renders the file into a temp file, and merges a merge target
// with the live target. It leaves the destination untouched when it is a
// seed that exists, or on any failure. Else it adds the target to pending
// and gives true: fileFinish does the rest after format.
func (s *state) file(tool string, f source.File, ctx *contexts) (fail.Code, bool) {
	path := f.Path
	src := s.src.ToolDir(tool) + "/" + path
	out := f.Target
	key := tool + "/" + out
	mode := s.fileMode(key)
	explicit := s.fileFormats[key]
	dest := s.root + "/" + out
	show := s.path(dest)
	// A mise file of an environment that this repo does not use. It gets no
	// manifest row, so the stale rule removes a listed copy with the header.
	// A seed stays.
	if env, ok := envOf(out); ok && !s.hasEnv(env) {
		if mode != "generated" {
			s.handled = append(s.handled, out)
		}
		s.r.Outcome(report.Skipped, show, "", "skip env   "+show)
		return fail.OK, false
	}

	if mode == "seed" && target.SeedKept(dest) {
		s.handled = append(s.handled, out)
		s.r.Outcome(report.Kept, show, "", "kept seed  "+show)
		return fail.OK, false
	}

	// A merge target below a symlinked folder: the repo side follows no
	// symlink, and that folder can point outside this repo. The walk goes
	// from the repo root down, and does not test the root itself.
	if mode == "merge" {
		folder, rest := s.root, out
		for strings.Contains(rest, "/") {
			first, next, _ := strings.Cut(rest, "/")
			folder += "/" + first
			rest = next
			if isLink(folder) {
				text := "the folder " + s.path(folder) + " is a symlink - kept it as it is"
				s.r.Outcome(report.Failed, show, text, show+": "+text)
				if exists(dest) {
					s.r.Outcome(report.Failed, show, "kept existing", "kept existing "+show)
				}
				return fail.Usage, false
			}
		}
	}

	// A synced file can become a folder in a later layout. Remove the old
	// synced file that is in the way.
	if s.e.Unblock(target.Dirname(dest), s.root, s.spare, show) != fail.OK {
		return fail.Other, false
	}

	var destdir string
	if s.e.DryRun {
		destdir = target.DryRunDir(s.getenv)
	} else {
		destdir = target.Dirname(dest)
		if os.MkdirAll(destdir, 0o777) != nil {
			text := "cannot create " + s.path(destdir)
			s.r.Outcome(report.Failed, show, text, text)
			return fail.Other, false
		}
	}
	// The temp file is in the destination folder, so the final rename is
	// atomic.
	tmp, err := s.e.Temps.Create(destdir)
	if err != nil {
		text := "cannot create a temp file in " + s.path(destdir)
		s.r.Outcome(report.Failed, show, text, text)
		return fail.Other, false
	}

	rc := fail.OK
	if f.Tmpl {
		if ctx == nil {
			rc = fail.Usage
			s.r.Outcome(report.Failed, show, "cannot write the target", "")
		} else if !s.render(tool, src, tmp, ctx) {
			s.r.Outcome(report.Failed, show, "gomplate failed to render the template",
				toolsShown+"/"+tool+"/"+path+": gomplate failed to render the template")
			rc = fail.Usage
		}
	} else if !copyFile(src, tmp) {
		rc = fail.Other
		s.r.Outcome(report.Failed, show, "cannot write the target", "")
	}

	// A JSON, YAML or TOML target must parse, before the formatter can
	// change it. A shared file has a spec as its source: the merge parses it.
	if rc == fail.OK && mode != "merge" {
		_, rc = s.e.Check(tmp, out, explicit, false, show)
	}

	// A merge target that is a symlink, also a dangling one: the merge reads
	// the live target, and a symlink can point outside this repo.
	if rc == fail.OK && mode == "merge" && isLink(dest) {
		s.r.Outcome(report.Failed, show, "is a symlink - kept it as it is", show+" is a symlink - kept it as it is")
		rc = fail.Usage
	}

	// A shared file: merge the rendered spec into the live target, before
	// the formatter.
	if rc == fail.OK && mode == "merge" {
		var ov []byte
		if files, ok := s.toolVals(tool).Get("files").(*source.Obj); ok && files.Has(out) {
			ov = []byte(source.JSON(files.Get(out)))
		}
		name := toolsShown + "/" + tool + "/" + path
		var rows []string
		rows, rc = s.e.Shared(tmp, dest, format.SharedInput{
			Target:    out,
			Explicit:  explicit,
			Overrides: ov,
			Records:   s.items[out],
			Header:    hdr.Shared(name, s.valuesInRepo()),
			Show:      show,
			SpecShow:  name,
		})
		if rc == fail.OK {
			s.newItems[out] = rows
		}
	}

	if rc != fail.OK {
		s.e.Abort(tmp, dest, show)
		return rc, false
	}

	// A rendered template, a generated mise config or a shared file goes
	// through the formatter of this repo before the compare: the file goes to
	// the stage, at its path in this repo, and format formats the stage in
	// one call of the formatter.
	if (f.Tmpl || mode == "merge" || s.generated[tool+"/"+path]) && s.stage != "" {
		p := s.stage + "/" + out
		if os.MkdirAll(target.Dirname(p), 0o777) == nil {
			copyFile(tmp, p)
		}
	}
	s.pending = append(s.pending, pend{tool, f, tmp})
	return fail.OK, true
}

// render renders the template <src> of <tool> into <tmp>. On a failure it
// prints the error to stderr, in the form of the log line of gomplate, and
// gives false.
func (s *state) render(tool, src, tmp string, ctx *contexts) bool {
	text, err := os.ReadFile(src)
	if err != nil {
		gomplateError(s.stderr, "renderTemplate: failed to render template "+src+": "+err.Error())
		return false
	}
	c, err := ctx.of(tool)
	if err != nil {
		gomplateError(s.stderr, "renderTemplate: failed to render template "+src+": "+err.Error())
		return false
	}
	out, err := tmpl.Render(src, text, c, nil)
	if err != nil {
		var ee template.ExecError
		if !errors.As(err, &ee) && strings.HasPrefix(err.Error(), "template: ") {
			gomplateError(s.stderr, "renderTemplate: parse template "+src+": "+err.Error())
		} else {
			gomplateError(s.stderr, "renderTemplate: failed to render template "+src+": "+err.Error())
		}
		return false
	}
	return os.WriteFile(tmp, out, 0o600) == nil
}

// format is `_rs_format`: it formats the files of the stage in one call of
// the formatter, with the paths of the files on stdin. Without the
// formatter, or if it fails for a file, the raw render of that file stays.
func (s *state) format() {
	if s.stage == "" {
		return
	}
	var files []string
	walk(s.stage, func(p string, info os.FileInfo) bool {
		if info.Mode().IsRegular() {
			files = append(files, strings.TrimPrefix(p, s.stage+"/"))
		}
		return true
	})
	if len(files) == 0 {
		return
	}
	cmd := exec.Command("mise", "x", "--", "bash", "-c", `cd "$1" && `+s.formatter.Command, "_", s.stage)
	cmd.Dir = s.root
	cmd.Stdin = strings.NewReader(strings.Join(files, "\n") + "\n")
	_ = cmd.Run()
}

// fileFinish is `_rs_file_finish`: the second step of the sync of the file
// <f> of <tool>. It takes the formatted copy from the stage, applies the
// overrides, and compares and writes the temp file <tmp> of file.
func (s *state) fileFinish(tool string, f source.File, tmp string) fail.Code {
	src := s.src.ToolDir(tool) + "/" + f.Path
	out := f.Target
	key := tool + "/" + out
	dest := s.root + "/" + out
	show := s.path(dest)
	if s.stage != "" && isFile(s.stage+"/"+out) {
		copyFile(s.stage+"/"+out, tmp)
	}

	// The folder gets its `dir_perm` before the write.
	if f.DirPerm != 0 {
		dir := s.root + "/" + target.Dirname(out)
		if _, code := s.e.DirPerm(dir, s.path(dir), f.DirPerm, s.dirDone); code != fail.OK {
			s.e.Abort(tmp, dest, show)
			return code
		}
	}

	// The file takes its `perm`, else the mode of its source in bootstrap,
	// also when it exists. An existing merge target without `perm` keeps its
	// own mode: the app owns it.
	fileMode := s.fileMode(key)
	var mode os.FileMode
	if f.Perm != 0 {
		mode = f.Perm
	} else if exists(dest) && fileMode == "merge" {
		mode, _ = target.Mode(dest)
	} else {
		mode, _ = target.Mode(src)
	}
	if code := s.e.Finish(mode, tmp, dest, show, fileMode); code != fail.OK {
		return code
	}
	s.handled = append(s.handled, out)
	if fileMode == "generated" {
		s.rows = append(s.rows, manifest.Row{Tool: tool, Path: out})
	}
	for _, item := range s.newItems[out] {
		s.rows = append(s.rows, manifest.Row{Tool: tool, Path: out, Item: item})
	}
	return fail.OK
}

// valuesInRepo gives the values file below the root of the repository, as
// the header of a shared file names it. A values file outside the repository
// (the plan of `bootstrap init`) is named as the default values file.
func (s *state) valuesInRepo() string {
	if rel, ok := strings.CutPrefix(s.values, s.root+"/"); ok {
		return rel
	}
	return ".config/bootstrap.toml"
}
