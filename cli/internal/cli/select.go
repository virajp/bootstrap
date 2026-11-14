package cli

import (
	"fmt"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/report"
)

// The selection commands: add tool, add pack, remove tool, replace tool and
// detect. Each one builds a change list of the [tools.<name>] tables of the
// values file, then plans and applies it with the steps of init (finish).

// cmdName gives the command as the user types it.
func cmdName(mac bool, verb string) string { return core.CmdName(mac, verb) }

// add is `bootstrap [mac] add tool <tool...>` and `bootstrap [mac] add pack
// <pack...> [--exclude <member>]...`.
func (a *app) add(mac bool, o opts) fail.Code { return a.change(mac, o, itemOf("add", o)) }

// itemOf gives the item of the change list of the command <cmd> (add,
// remove or replace) with the kind word and the names of <o>.
func itemOf(cmd string, o opts) core.Item {
	switch {
	case cmd == "add" && o.kind == "pack":
		return core.Item{Kind: core.AddPack, Names: o.names, Exclude: o.exclude}
	case cmd == "add":
		return core.Item{Kind: core.AddTool, Names: o.names}
	case cmd == "remove":
		return core.Item{Kind: core.RemoveTool, Names: o.names}
	}
	return core.Item{Kind: core.Replace, Old: o.names[0], New: o.names[1]}
}

// remove is `bootstrap [mac] remove tool <tool...>`.
func (a *app) remove(mac bool, o opts) fail.Code {
	return a.change(mac, o, itemOf("remove", o))
}

// replace is `bootstrap [mac] replace tool <old> <new>`.
func (a *app) replace(mac bool, o opts) fail.Code {
	if len(o.names) != 2 {
		a.usageErr(cmdWord(mac, "replace tool")+" needs two tools", usageLine(mac, "replace", "dry-run yes answer jobs verbose"))
		return fail.Usage
	}
	return a.change(mac, o, itemOf("replace", o))
}

// change does the preflight, takes the clone, and plans and applies the
// change list of <items>.
func (a *app) change(mac bool, o opts, items ...core.Item) fail.Code {
	sd, c, code := a.side(mac, false, mac)
	if c != nil {
		defer c.Cleanup()
	}
	if code != fail.OK {
		return code
	}
	return a.finish(sd, o, core.List{Items: items})
}

// detect is `bootstrap [mac] detect`. With no flag, it lists the detected
// tools and exits 1 when one is not selected; with --add, it adds each one
// after the consent.
func (a *app) detect(mac bool, o opts) fail.Code {
	sd, c, code := a.side(mac, false, mac)
	if c != nil {
		defer c.Cleanup()
	}
	if code != fail.OK {
		return code
	}
	d, err := core.Detect(sd)
	if d != nil {
		for _, n := range d.Notes {
			a.infof("%s", n)
		}
	}
	if err != nil {
		return a.logErr(err)
	}
	if d.NoScripts {
		a.reasons = append(a.reasons, "no tool has a detect script")
		a.nothingToChange()
		return fail.OK
	}
	s := d.State
	code = a.logErr(d.Mismatch)

	// The tools that match and are not selected: the ones to offer, and the
	// ones whose evidence a selected tool has already.
	type notSelected struct {
		tool, status, text string
		offered            bool
	}
	var rest []notSelected
	for _, p := range d.Proposals {
		if p.Offered {
			rest = append(rest, notSelected{p.Tool, "offered", detectText(p.Found), true})
		} else {
			rest = append(rest, notSelected{p.Tool, "not offered", p.Shared + " is the evidence of " + p.Of, false})
		}
	}
	style := a.styled
	opt := treeOpts{style: style,
		include: func(t string) bool { _, ok := d.Found[t]; return ok && s.Selected(t) },
		cell: func(t string) (cell, string) {
			f, ok := d.Found[t]
			if !ok {
				return cell{}, ""
			}
			var name, word string
			switch f.Result {
			case "match":
				name, word = "ok", "found"
			case "mismatch":
				name, word = "error", "mismatch"
			case "none":
				name, word = "info", "nothing found"
				if s.Named[t] && s.PackOf(t) == "" {
					name = "none"
				}
			}
			vis, painted := report.GlyphParts(name, style)
			detail := f.Message
			if f.Result == "match" {
				detail = detectText(f)
			}
			return cell{vis: vis + " " + word, colored: painted + " " + word}, detail
		}}
	// The installed packages of no Mac tool: each one is an offer for the own
	// [mise] tables of macos.toml.
	for _, p := range d.Unknown {
		rest = append(rest, notSelected{p.Key, "offered", "no Mac tool has it: add it to " + p.Table() + " of macos.toml", true})
	}
	if len(rest) > 0 {
		opt.groupTitle = "Found, not selected"
		for _, n := range rest {
			vis, painted := report.GlyphParts("info", style)
			if n.offered {
				vis, painted = "✚", report.Mark(report.New, style)
				if !style {
					vis, painted = "[NEW]", "[NEW]"
				}
			}
			opt.group = append(opt.group, groupRow{
				lead: "  " + vis + " " + n.tool, painted: "  " + painted + " " + n.tool,
				cell: cell{vis: n.status, colored: n.status}, det: n.text,
			})
		}
	}
	tree, _ := renderTree(listRoots(s), depsOf(s), opt)
	_, _ = fmt.Fprint(a.stdout, tree)
	if code != fail.OK {
		return code
	}
	if len(d.Offers) == 0 && len(d.Unknown) == 0 {
		_, _ = fmt.Fprintln(a.stdout)
		report.NewPrinter(a.stdout, style).Success("No tool to add")
		return fail.OK
	}
	if !o.add {
		for _, t := range d.Offers {
			a.fail(t+" is found, but it is not selected", "run "+cmdName(mac, "add tool")+" "+t+", or "+cmdName(mac, "detect")+" --add")
		}
		for _, p := range d.Unknown {
			a.fail(p.Key+" is installed, but no Mac tool has it", "add it to "+p.Table()+" of macos.toml, or run "+cmdName(mac, "detect")+" --add")
		}
		return fail.Other
	}
	return a.finish(sd, o, core.List{Items: d.Items(d.Offers, d.Unknown)})
}

// detectText gives the evidence and the message of a match: the first
// evidence, "+N more" for the others, then the message.
func detectText(d core.Detected) string {
	text := ""
	if len(d.Evidence) > 0 {
		text = d.Evidence[0]
		if n := len(d.Evidence) - 1; n > 0 {
			text += fmt.Sprintf(", +%d more", n)
		}
	}
	switch {
	case text == "":
		return d.Message
	case d.Message == "":
		return text
	}
	return text + ": " + d.Message
}
