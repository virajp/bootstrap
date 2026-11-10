package repo

// The inline templates of the bash engine, byte for byte. Each one runs
// through internal/tmpl, as bash runs it through gomplate.

// tmplMembers is _rs_members: the `[member.<name>]` tables as `<name><TAB><folder><TAB><type>` lines.
const tmplMembers = `<% if has .v "member" %><% range $k, $x := .v.member %><% $k %>	<% if has $x "folder" %><% $x.folder %><% else %><% $k %><% end %>	<% if has $x "type" %><% $x.type %><% end %>
<% end %><% end %>`

// tmplRepoName is `repo_name` of a values file (_rs_members).
const tmplRepoName = `<% if has .v "repo_name" %><% .v.repo_name %><% end %>`

// tmplContext is _rs_context: the shared template context as JSON.
const tmplContext = `<% $ctx := dict %><% range (coll.Slice "members" "submodules" "tools") %><% $l := coll.Slice %><% range (env.Getenv (printf "REPO_SYNC_%s" (strings.ToUpper .)) | strings.Split "\n") %><% if . %><% $l = coll.Append . $l %><% end %><% end %><% $ctx = coll.Merge $ctx (dict . $l) %><% end %><% $mf := dict %><% $mt := dict %><% $fl := env.Getenv "REPO_SYNC_MEMBER_FOLDERS" | strings.Split "\n" %><% $tl := env.Getenv "REPO_SYNC_MEMBER_TYPES" | strings.Split "\n" %><% range $i, $n := $ctx.members %><% $mf = coll.Merge (dict $n (index $fl $i)) $mf %><% $mt = coll.Merge (dict $n (index $tl $i)) $mt %><% end %><% $ctx = coll.Merge $ctx (dict "member_folders" $mf "member_types" $mt "sync_engine" (env.Getenv "REPO_SYNC_SYNC_ENGINE") "all" .all_values) %><% coll.Merge $ctx (coll.Omit "all" "config" .v) | data.ToJSON %>`
