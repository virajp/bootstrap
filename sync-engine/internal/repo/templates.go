package repo

// The inline templates of sync-engine/repo-sync, byte for byte. Each one runs
// through internal/tmpl, as bash runs it through gomplate.

// tmplMembers is _rs_members: the `[member.<name>]` tables as `<name><TAB><folder><TAB><type>` lines.
const tmplMembers = `<% if has .v "member" %><% range $k, $x := .v.member %><% $k %>	<% if has $x "folder" %><% $x.folder %><% else %><% $k %><% end %>	<% if has $x "type" %><% $x.type %><% end %>
<% end %><% end %>`

// tmplRepoName is `repo_name` of a values file (_rs_members).
const tmplRepoName = `<% if has .v "repo_name" %><% .v.repo_name %><% end %>`

// tmplGitignore is _RS_GITIGNORE_TEMPLATE.
const tmplGitignore = `<% if and (has .o "overrides") (test.IsKind "map" .o.overrides) -%>
<% range $k, $x := .o.overrides -%>
<% if not (has (coll.Slice "mise" "mise_env" "gitignore") $k) %>key	overrides.<% $k %>	-
<% else if eq $k "gitignore" -%>
<% if not (test.IsKind "map" $x) %>table	overrides.<% $k %>	-
<% else -%>
<% range $gk, $gv := $x -%>
<% if not (has (coll.Slice "add" "remove") $gk) %>gikey	overrides.<% $k %>.<% $gk %>	-
<% else if not (test.IsKind "slice" $gv) %>gilist	overrides.<% $k %>.<% $gk %>	-
<% else -%>
<% range $gv -%>
<% if not (test.IsKind "string" .) %>gilist	overrides.<% $k %>.<% $gk %>	-
<% else if or (strings.Contains "\n" .) (strings.Contains "\t" .) %>gilist	overrides.<% $k %>.<% $gk %>	-
<% else if not (strings.TrimSpace .) %>giempty	overrides.<% $k %>.<% $gk %>	-
<% else %><% $gk %>	overrides.<% $k %>.<% $gk %>	<% strings.TrimSpace . %>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>`

// tmplContext is _rs_context: the shared template context as JSON.
const tmplContext = `<% $ctx := dict %><% range (coll.Slice "members" "submodules" "tools" "extensions") %><% $l := coll.Slice %><% range (env.Getenv (printf "REPO_SYNC_%s" (strings.ToUpper .)) | strings.Split "\n") %><% if . %><% $l = coll.Append . $l %><% end %><% end %><% $ctx = coll.Merge $ctx (dict . $l) %><% end %><% $mf := dict %><% $mt := dict %><% $fl := env.Getenv "REPO_SYNC_MEMBER_FOLDERS" | strings.Split "\n" %><% $tl := env.Getenv "REPO_SYNC_MEMBER_TYPES" | strings.Split "\n" %><% range $i, $n := $ctx.members %><% $mf = coll.Merge (dict $n (index $fl $i)) $mf %><% $mt = coll.Merge (dict $n (index $tl $i)) $mt %><% end %><% $ctx = coll.Merge $ctx (dict "member_folders" $mf "member_types" $mt "sync_engine" (env.Getenv "REPO_SYNC_SYNC_ENGINE")) %><% $pc := coll.Slice %><% range $ctx.tools %><% $t := index $ (printf "tool_%s" .) %><% if has $t "precommit" %><% range $t.precommit %><% $pc = coll.Append . $pc %><% end %><% end %><% end %><% if env.Getenv "REPO_SYNC_OWN_PRECOMMIT" %><% range $.repo_precommit %><% $pc = coll.Append . $pc %><% end %><% end %><% $drop := strings.Split " " (strings.TrimSpace (env.Getenv "REPO_SYNC_HOOK_DROP")) -%>
<% $npc := coll.Slice -%>
<% range $e := $pc -%>
<% $ne := $e %><% $keep := true -%>
<% if and (test.IsKind "map" $e) (has $e "hooks") (test.IsKind "slice" (index $e "hooks")) -%>
<% $hs := coll.Slice %><% $chg := false -%>
<% range $h := index $e "hooks" -%>
<% $nh := $h %><% $hk := true -%>
<% if and (test.IsKind "map" $h) (has $h "id") (eq (printf "%v" (index $h "id")) "no-commit-to-branch") (has $h "args") (test.IsKind "slice" (index $h "args")) -%>
<% $in := index $h "args" %><% $n := len $in %><% $a := coll.Slice %><% $had := false %><% $left := false %><% $skip := false -%>
<% range $i, $x := $in -%>
<% if $skip %><% $skip = false -%>
<% else if and (eq (printf "%v" $x) "--branch") (lt (math.Add $i 1) $n) -%>
<% $had = true %><% $skip = true %><% $b := index $in (math.Add $i 1) -%>
<% if not (has $drop (printf "%v" $b)) %><% $left = true %><% $a = coll.Append $b (coll.Append $x $a) %><% end -%>
<% else %><% $a = coll.Append $x $a %><% end -%>
<% end -%>
<% if $had -%>
<% if not $left %><% $hk = false %><% $chg = true -%>
<% else if ne (len $a) $n %><% $nh = coll.Merge (dict "args" $a) $h %><% $chg = true %><% end -%>
<% end -%>
<% end -%>
<% if $hk %><% $hs = coll.Append $nh $hs %><% end -%>
<% end -%>
<% if $chg %><% if $hs %><% $ne = coll.Merge (dict "hooks" $hs) $e %><% else %><% $keep = false %><% end %><% end -%>
<% end -%>
<% if $keep %><% $npc = coll.Append $ne $npc %><% end -%>
<% end -%>
<% $ctx = coll.Merge $ctx (dict "precommit" $npc "all" .all_values) %><% coll.Merge $ctx (coll.Omit "all" "config" .v) | data.ToJSON %>`
