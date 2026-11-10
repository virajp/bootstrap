package misegen

// The templates of sync-engine/mise, byte for byte. The test
// TestTemplatesEqualBash compares them with the bash variables.

// defines is _SE_MISE_DEFINES: the named templates that the override check
// and the mise template share. "se_leaves" prints the key path of each value
// of the table .v that is not a table, as a JSON list after the path .p, one
// per line. "se_has" prints 1 when the table .v has a value at the path .p
// that is not a table.
const defines = `<% define "se_leaves" %><% range $k, $x := .v %><% $p := coll.Append $k $.p %><% if test.IsKind "map" $x %><% template "se_leaves" (dict "p" $p "v" $x) %><% else %><% $p | data.ToJSON %>
<% end %><% end %><% end -%>
<% define "se_has" %><% $v := .v %><% $ok := true %><% range .p %><% if not (test.IsKind "map" $v) %><% $ok = false %><% else if not (has $v .) %><% $ok = false %><% else %><% $v = index $v . %><% end %><% end %><% if and $ok (not (test.IsKind "map" $v)) %>1<% end %><% end -%>
`

// overridesTemplate is _SE_MISE_OVERRIDES_TEMPLATE without the defines. It
// examines the mise and mise_env keys of [overrides] of the values file (.o)
// against the tools.toml of each selected tool (.tool_<name>). It prints one
// "<problem> <path> <arg>" line, tab separated, for each problem.
const overridesTemplate = defines + `<% if has .o "overrides" -%>
<% $ov := .o.overrides -%>
<% $side := env.Getenv "BOOTSTRAP_MISE_SIDE" -%>
<% $envs := env.Getenv "BOOTSTRAP_MISE_ENVS" -%>
<% $tools := strings.Split " " (env.Getenv "BOOTSTRAP_MISE_TOOLS") -%>
<% if not (test.IsKind "map" $ov) %>table	overrides	-
<% else -%>
<% $ov = coll.Pick "mise" "mise_env" $ov -%>
<% range $k, $x := $ov -%>
<% if and (eq $side "macos") (eq $k "mise_env") %>noenv	overrides.mise_env	-
<% else if not (test.IsKind "map" $x) %>table	overrides.<% $k %>	-
<% else if eq $k "mise_env" -%>
<% range $e, $y := $x -%>
<% if not (has (coll.Slice "dev" "ci") $e) %>envname	overrides.mise_env.<% $e %>	<% $e %>
<% else if not (strings.Contains (printf " %s " $e) $envs) %>env	overrides.mise_env.<% $e %>	<% $e %>
<% else if not (test.IsKind "map" $y) %>table	overrides.mise_env.<% $e %>	-
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% range strings.Split "\n" (tmpl.Exec "se_leaves" (dict "p" (coll.Slice) "v" $ov)) -%>
<% if . -%>
<% $p := data.JSONArray . -%>
<% $k := index $p 0 -%>
<% $rest := coll.Slice -%>
<% $only := "" -%>
<% if and (eq $k "mise") (gt (len $p) 1) %><% $rest = slice $p 1 -%>
<% else if and (eq $k "mise_env") (gt (len $p) 2) (has (coll.Slice "dev" "ci") (index $p 1)) (strings.Contains (printf " %s " (index $p 1)) $envs) %><% $rest = slice $p 2 %><% $only = index $p 1 %><% end -%>
<% if $rest -%>
<% if and (gt (len $rest) 2) (eq (index $rest 0) "tools") (has (coll.Slice "envs" "depends" "requires" "requires_found" "files") (index $rest 2)) -%>
lock	overrides.<% join $p "." %>	-
<% else -%>
<% $found := false -%>
<% range $tools -%>
<% if . -%>
<% $t := index $ (printf "tool_%s" .) -%>
<% if tmpl.Exec "se_has" (dict "v" (index $t "mise") "p" $rest) %><% $found = true %><% end -%>
<% if has $t "mise_env" -%>
<% range $e, $y := $t.mise_env -%>
<% if and (strings.Contains (printf " %s " $e) $envs) (or (not $only) (eq $only $e)) (tmpl.Exec "se_has" (dict "v" $y "p" $rest)) %><% $found = true %><% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% if not $found -%>
none	overrides.<% join $p "." %>	<% if $only %>mise_env.<% $only %>.<% join $rest "." %> or <% end %>mise.<% join $rest "." %>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>`

// miseTemplate is _SE_MISE_TEMPLATE without the defines. It builds one output
// from the [mise] and [mise_env] tables of a tools.toml (.t), as TOML, with the
// overrides of the values file (.o). After the TOML, a "#@override <path>" line
// names each output path that an override changed.
const miseTemplate = defines + `<% $m := dict %><% if has .t "mise" %><% $m = .t.mise %><% end -%>
<% $macos := eq (env.Getenv "BOOTSTRAP_MISE_SIDE") "macos" -%>
<% if and $macos (has .t "mise_env") %><% test.Fail "[mise_env.*] is not valid on the macos side, which has no environments" %><% end -%>
<% $env := env.Getenv "BOOTSTRAP_MISE_ENV" -%>
<% $me := dict %><% if and (has .t "mise_env") (has .t.mise_env $env) %><% $me = index .t.mise_env $env %><% end -%>
<% $marks := coll.Slice -%>
<% if and (not (env.Getenv "BOOTSTRAP_MISE_OWN")) (has .o "overrides") (test.IsKind "map" .o.overrides) -%>
<% $ov := coll.Pick "mise" "mise_env" .o.overrides -%>
<% range strings.Split "\n" (tmpl.Exec "se_leaves" (dict "p" (coll.Slice) "v" $ov)) -%>
<% if . -%>
<% $p := data.JSONArray . -%>
<% $k := index $p 0 -%>
<% $rest := coll.Slice -%>
<% if and (eq $k "mise") (gt (len $p) 1) %><% $rest = slice $p 1 -%>
<% else if and (eq $k "mise_env") (gt (len $p) 2) (eq (index $p 1) $env) %><% $rest = slice $p 2 %><% end -%>
<% if $rest -%>
<% $v := $ov %><% range $p %><% $v = index $v . %><% end -%>
<% $n := $v %><% range math.Seq (math.Sub (len $rest) 1) 0 %><% $n = dict (index $rest .) $n %><% end -%>
<% $inm := tmpl.Exec "se_has" (dict "v" $m "p" $rest) -%>
<% $ine := tmpl.Exec "se_has" (dict "v" $me "p" $rest) -%>
<% if and $inm (eq $k "mise") %><% $m = coll.Merge $n $m %><% end -%>
<% if or $ine (and $inm (eq $k "mise_env")) %><% $me = coll.Merge $n $me %><% end -%>
<% if or $inm $ine %><% $marks = coll.Append (join $rest ".") $marks %><% end -%>
<% end -%>
<% end -%>
<% end -%>
<% end -%>
<% $envw := env.Getenv "BOOTSTRAP_MISE_ENVS" -%>
<% $entry := coll.Slice "tools" "tool_alias" "plugins" -%>
<% $folder := coll.Slice "settings" "env" "tool_config" "task_config" "task_templates" "doctor" "deps" "oci" -%>
<% $devonly := coll.Slice "vars" "secrets" "redactions" "shell_alias" "hooks" "watch_files" "daemons" "daemon_groups" "daemons_settings" "daemon_providers" "tasks" "bootstrap" "dotfiles" -%>
<% $in := or $macos (eq (env.Getenv "BOOTSTRAP_MISE_IN") "1") -%>
<% $out := dict -%>
<% range $s, $body := $m -%>
<% if has $entry $s -%>
<% $a := dict -%>
<% range $k, $v := $body -%>
<% if and (eq $s "tools") (test.IsKind "map" $v) -%>
<% if $macos -%>
<% range coll.Slice "envs" "requires_found" %><% if has $v . %><% test.Fail (printf "%s of mise tool %s is not valid on the macos side, which has no environments and no requires_found" . $k) %><% end %><% end -%>
<% end -%>
<% $v = coll.Omit "envs" "requires_found" $v -%>
<% end -%>
<% if or $macos (strings.Contains (printf " %s:%s " $k $env) $envw) %><% $a = coll.Merge (dict $k $v) $a %><% end -%>
<% end -%>
<% if $a %><% $out = coll.Merge (dict $s $a) $out %><% end -%>
<% else if has $folder $s -%>
<% if $in %><% $out = coll.Merge (dict $s $body) $out %><% end -%>
<% else if has $devonly $s -%>
<% if and $in (or $macos (eq $env "dev")) %><% $out = coll.Merge (dict $s $body) $out %><% end -%>
<% else -%>
<% test.Fail (printf "[mise.%s] is not supported in a tools.toml" $s) -%>
<% end -%>
<% end -%>
<% range $s, $body := $me -%>
<% if not (or (has $entry $s) (has $folder $s) (has $devonly $s)) %><% test.Fail (printf "[mise_env.%s.%s] is not supported in a tools.toml" $env $s) %><% end -%>
<% if $in -%>
<% if has $out $s %><% $body = coll.Merge $body (index $out $s) %><% end -%>
<% $out = coll.Merge (dict $s $body) $out -%>
<% end -%>
<% end -%>
<% if $out %><% $out | data.ToTOML %><% end -%>
<% range $marks %>#@override <% . %>
<% end -%>`
