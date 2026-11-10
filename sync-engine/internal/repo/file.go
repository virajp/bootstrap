package repo

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/format"
	"github.com/virajp/bootstrap/sync-engine/internal/manifest"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/target"
	"github.com/virajp/bootstrap/sync-engine/internal/tmpl"
)

// dprintConfigs are the dprint config files of a repo that the stage gets.
var dprintConfigs = []string{"dprint.json", "dprint.jsonc", ".dprint.json", ".dprint.jsonc", ".config/dprint.json", ".config/dprint.jsonc"}

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

	// The hook revisions: each one becomes the newest tag, chosen by
	// `pre-commit autoupdate` itself. It needs the network; on a failure the
	// existing file stays. REPO_SYNC_NO_AUTOUPDATE=1 skips it.
	if rc == fail.OK && out == ".config/pre-commit-config.yaml" {
		name := toolsShown + "/" + tool + "/" + path
		switch {
		case s.getenv("REPO_SYNC_NO_AUTOUPDATE") == "1":
			s.log(name + ": REPO_SYNC_NO_AUTOUPDATE=1 - pre-commit autoupdate skipped")
		case lookPath("pre-commit", s.getenv("PATH")) == "":
			text := "pre-commit is not available to resolve the hook revisions"
			s.r.Outcome(report.Failed, show, text, name+": "+text)
			rc = fail.Network
		case !run(s.root, lookPath("pre-commit", s.getenv("PATH")), "autoupdate", "--config", tmp):
			text := "pre-commit autoupdate failed - examine the network and the hook repositories"
			s.r.Outcome(report.Failed, show, text, name+": "+text)
			rc = fail.Network
		}
	}

	// A JSON, YAML or TOML target must parse, before dprint can change it.
	// The JSONC rule is not for a merge target.
	var ff format.Format
	if rc == fail.OK {
		ff, rc = s.e.Check(tmp, out, explicit, mode == "merge", show)
	}

	// A merge target that is a symlink, also a dangling one: the merge reads
	// the live target, and a symlink can point outside this repo.
	if rc == fail.OK && mode == "merge" && isLink(dest) {
		s.r.Outcome(report.Failed, show, "is a symlink - kept it as it is", show+" is a symlink - kept it as it is")
		rc = fail.Usage
	}

	// A merge target: merge with the live target, before dprint.
	if rc == fail.OK && mode == "merge" {
		remove, items := "[]", "[]"
		if m, ok := s.fileMerge[key]; ok {
			if m[0] != "" {
				remove = m[0]
			}
			if m[1] != "" {
				items = m[1]
			}
		}
		ops, err := format.ParseOps(remove, items)
		if err != nil {
			rc = fail.Usage
		} else {
			rc = s.e.Merge(tmp, dest, ff, ops, show)
		}
	}

	if rc != fail.OK {
		s.e.Abort(tmp, dest, show)
		return rc, false
	}

	// A rendered template, a generated mise config or a merge target goes
	// through the formatter of this repo before the compare: the file goes to
	// the stage, at its path in this repo, and format formats the stage in
	// one dprint call.
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

// stageSkip is the grep of `_rs_format`: the dprint config files of the
// stage are not formatted.
var stageSkip = regexp.MustCompile(`^(\.?dprint\.jsonc?|\.config/dprint\.jsonc?)$`)

// format is `_rs_format`: it formats the files of the stage in one dprint
// call. Without dprint, or if dprint fails for a file, the raw render of that
// file stays.
func (s *state) format() {
	if s.stage == "" {
		return
	}
	var files []string
	walk(s.stage, func(p string, info os.FileInfo) bool {
		if info.Mode().IsRegular() {
			rel := strings.TrimPrefix(p, s.stage+"/")
			if !stageSkip.MatchString(rel) {
				files = append(files, rel)
			}
		}
		return true
	})
	if len(files) == 0 {
		return
	}
	cmd := exec.Command("mise", "x", "--", "bash", "-c", `cd "$1" && dprint fmt --no-gitignore --allow-no-files --stdin-files`, "_", s.stage)
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

	// [overrides.gitignore]: comment out each line of `remove`, then add the
	// lines of `add` at the end, so they win.
	rc := fail.OK
	if out == ".gitignore" && len(s.giRemove)+len(s.giAdd) > 0 {
		ignore := strings.TrimPrefix(out, ".")
		pick := func(list []giLine) string {
			var lines []string
			for _, l := range list {
				if l.ignore == ignore {
					lines = append(lines, l.line)
				}
			}
			return strings.Join(lines, "\n")
		}
		add, rem := pick(s.giAdd), pick(s.giRemove)
		if add+rem != "" && !s.gitignore(tmp, add, rem) {
			rc = fail.Other
			s.r.Outcome(report.Failed, show, "cannot write the target", "")
		}
	}
	if rc != fail.OK {
		s.e.Abort(tmp, dest, show)
		return rc
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
	return fail.OK
}

// gitignore applies `[overrides.gitignore]` to the temp file <tmp>: each line
// of <rem> becomes a comment, and the lines of <add> go at the end. It gives
// false when it cannot write.
func (s *state) gitignore(tmp, add, rem string) bool {
	b, err := os.ReadFile(tmp)
	if err != nil {
		return false
	}
	drop := map[string]bool{}
	for _, l := range strings.Split(rem, "\n") {
		if l != "" {
			drop[l] = true
		}
	}
	note := "# override: removed by " + s.valuesNote
	var out strings.Builder
	for _, l := range awkLines(b) {
		if t := trimBlanks(l); drop[t] {
			out.WriteString("# " + t + "  " + note + "\n")
		} else {
			out.WriteString(l + "\n")
		}
	}
	if add != "" {
		out.WriteString("\n# ==== Overrides from " + s.valuesNote + " ====\n" + add + "\n")
	}
	gi := tmp + ".gi"
	if os.WriteFile(gi, []byte(out.String()), 0o644) != nil {
		return false
	}
	if os.Rename(gi, tmp) != nil {
		os.Remove(gi)
		return false
	}
	return true
}
