package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// The golden file testdata/funcs.json holds the output of the gomplate CLI
// for each case of funcCases. TestFuncsGolden compares the Go output with it.
// TestFuncsLive makes sure that the golden file is still the CLI output, when
// the gomplate CLI is on the machine. To write the golden file again from the
// CLI, run: UPDATE_GOLDEN=1 go test ./internal/tmpl

type fnCase struct {
	Tmpl    string `json:"tmpl"`
	Ctx     string `json:"ctx"`
	Missing string `json:"missing,omitempty"`
	// The result of the CLI.
	OK  bool   `json:"ok"`
	Out string `json:"out,omitempty"`
	Msg string `json:"msg,omitempty"` // the text after "error calling X: "
}

// testEnv is the environment of the env.Getenv cases.
var testEnv = map[string]string{"TMPL_T_A": "alpha", "TMPL_T_EMPTY": "", "TMPL_T_NL": "a\nb\n\nc"}

// The context of most cases.
const baseCtx = `{"s":"str","e":"","n":3,"f":1.5,"z":3.0,"big":12345678901234567890,"neg":-5,"t":true,"fa":false,
"nil":null,"l":["a",1,2.5,true,null],"el":[],"m":{"a":1,"b":{"c":[1,2]}},"em":{},
"strs":["x","y"],"path":"~/p","home":{"HOME":"/h"},"ls":[{"k":"v"},{"k":"w"}]}`

func funcCases() []fnCase {
	var c []fnCase
	add := func(tmpl string, ctx ...string) {
		x := baseCtx
		if len(ctx) > 0 {
			x = ctx[0]
		}
		c = append(c, fnCase{Tmpl: tmpl, Ctx: x})
	}
	addMissing := func(tmpl, ctx, mk string) { c = append(c, fnCase{Tmpl: tmpl, Ctx: ctx, Missing: mk}) }

	// Plain output and numbers.
	for _, k := range []string{"s", "e", "n", "f", "z", "big", "neg", "t", "fa", "nil", "l", "el", "m", "em", "ls"} {
		add("<% ." + k + " %>")
		add("<% test.Kind ." + k + " %>")
		add("<% data.ToJSON ." + k + " %>")
		add("<% data.ToJSONPretty \"  \" ." + k + " %>")
		add("<% printf \"%v|%d\" ." + k + " 1 %>")
		add("<% conv.ToString ." + k + " %>")
		add("<% strings.Quote ." + k + " %>")
		add("<% default \"d\" ." + k + " %>")
		add("<% ." + k + " | default \"d\" %>")
		add("<% required \"msg\" ." + k + " %>")
		add("<% has . \"" + k + "\" %>")
		add("<% has .l ." + k + " %>")
		add("<% has .strs ." + k + " %>")
		add("<% len ." + k + " %>")
		add("<% if ." + k + " %>yes<% else %>no<% end %>")
	}
	for _, k := range []string{"s", "n", "f", "z", "nil", "l", "m", "t"} {
		for _, kind := range []string{"string", "int", "int64", "float", "float64", "number", "bool", "map", "slice", "array", "invalid", "nil", "uint64"} {
			add("<% test.IsKind \"" + kind + "\" ." + k + " %>")
		}
	}
	add(`<% test.Kind (strings.Split "," "a,b") %>|<% test.Kind (math.Seq 3) %>|<% test.Kind (dict) %>|<% test.Kind (coll.Slice) %>|<% test.Kind (math.Sub 5 2) %>|<% test.Kind (math.Sub 5 2.5) %>|<% test.Kind (math.Add 1 2.0) %>|<% test.Kind (math.Add .z 1) %>|<% test.Kind (math.Add .n 1) %>`)
	add(`<% test.IsKind "bogus" 1 %>`)
	add(`<% test.IsKind "map" %>`)
	add(`<% test.Fail "boom" %>`)
	add(`<% test.Fail %>`)
	add(`<% test.Fail 12 %>`)
	add(`<% test.Fail "a" "b" %>`)
	add(`<% printf "%s of mise tool %s is bad" "envs" "x" | test.Fail %>`)

	// has, dict, coll
	add(`<% has . "n" %>|<% has .m "a" %>|<% has .m "zz" %>|<% has .nil "a" %>|<% has "str" "a" %>|<% has .l "a" %>|<% has .l "zz" %>|<% has .el "a" %>|<% has .em "a" %>`)
	add(`<% has 3 "a" %>`)
	add(`<% has .m 1 %>`)
	add(`<% has (coll.Slice "dev" "ci") "dev" %>|<% coll.Has (coll.Slice "dev") "ci" %>`)
	add(`<% has (dict "a" nil) "a" %>`)
	add(`<% dict "a" 1 "b" | data.ToJSON %>`)
	add(`<% dict 1 2 true 3 | data.ToJSON %>`)
	add(`<% dict | data.ToJSON %>`)
	add(`<% dict "a" .m "b" .l | data.ToJSON %>`)
	add(`<% coll.Slice | data.ToJSON %>`)
	add(`<% coll.Slice 1 "a" nil .m | data.ToJSON %>`)
	add(`<% coll.Slice "tools" "packs" | data.ToJSON %>`)
	add(`<% coll.Append "x" .strs | data.ToJSON %>`)
	add(`<% coll.Append "x" .el | data.ToJSON %>`)
	add(`<% coll.Append "x" .l | data.ToJSON %>|<% .l | data.ToJSON %>`)
	add(`<% coll.Append "x" (coll.Slice) | data.ToJSON %>`)
	add(`<% coll.Append "x" .s %>`)
	add(`<% coll.Append "x" nil %>`)
	add(`<% coll.Append .m (coll.Slice 1) | data.ToJSON %>`)
	add(`<% coll.Omit "a" .m | data.ToJSON %>`)
	add(`<% coll.Omit "a" "b" "c" .m | data.ToJSON %>`)
	add(`<% coll.Omit "zz" .m | data.ToJSON %>`)
	add(`<% coll.Omit .m | data.ToJSON %>`)
	add(`<% coll.Omit "env" (dict "env" 1 "x" 2) | data.ToJSON %>`)
	add(`<% coll.Omit 1 (dict "1" 1 "x" 2) | data.ToJSON %>`)
	add(`<% coll.Pick "a" .m | data.ToJSON %>`)
	add(`<% coll.Pick "mise" "mise_env" (dict "mise" 1 "x" 2) | data.ToJSON %>`)
	add(`<% coll.Pick "zz" .m | data.ToJSON %>`)
	add(`<% coll.Pick .m | data.ToJSON %>`)
	add(`<% coll.Pick "a" .l %>`)
	add(`<% $a := dict "x" 1 %><% $b := coll.Merge $a (dict "y" 2) %><% $a | data.ToJSON %>|<% $b | data.ToJSON %>`)
	add(`<% coll.Merge .m %>`)
	add(`<% coll.Merge .m .m | data.ToJSON %>`)
	add(`<% coll.Merge .m %>`)
	add(`<% coll.Merge .nil .m %>`)
	add(`<% coll.Merge .l .m %>`)
	add(`<% coll.Merge (dict) (dict "a" 1) (dict "a" 2 "b" 3) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" "") (dict "a" 2) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" 0) (dict "a" 2) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" false) (dict "a" true) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" nil) (dict "a" 2) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" (coll.Slice)) (dict "a" (coll.Slice 1)) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" (coll.Slice 1)) (dict "a" (coll.Slice 2)) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" (dict)) (dict "a" (dict "b" 1)) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" (dict "b" 1)) (dict "a" (dict "b" 2 "c" 3)) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" (dict "b" 1)) (dict "a" 5) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" 5) (dict "a" (dict "b" 1)) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" "x") (dict "a" 5) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" 1.5) (dict "a" 5) | data.ToJSON %>`)
	add(`<% coll.Merge (dict "a" 1) (dict "a" 1.5) | data.ToJSON %>`)
	add(`<% $x := dict "k" (dict "v" 1) %><% $y := coll.Merge (dict "t" 1) $x %><% $z := coll.Merge $y (dict "k" (dict "w" 2)) %><% $x | data.ToJSON %>|<% $y | data.ToJSON %>|<% $z | data.ToJSON %>`)

	// conv
	add(`<% conv.Join .strs "," %>|<% conv.Join .l "," %>|<% conv.Join .el "," %>|<% join .strs "|" %>`)
	add(`<% join .m "," %>`)
	add(`<% join .s "," %>`)
	add(`<% join .nil "," %>`)
	add(`<% join (strings.Split "," "a,b") "-" %>`)
	add(`<% conv.Join .strs "|" | strings.Quote %>`)
	add(`<% conv.ToString nil %>|<% conv.ToString 1.5 %>|<% conv.ToString .l %>|<% conv.ToString .m %>|<% conv.ToString true %>`)
	add(`<% default "d" %>|<% default "d" "" %>|<% default "d" 0 %>|<% default "d" "x" %>|<% default 1 2 %>|<% default "d" nil %>|<% default "d" (dict) %>`)
	add(`<% default "d" 1 2 %>`)
	add(`<% index .ls 0 | data.ToJSON %>|<% index (index .ls 0) "k" | default "z" %>|<% index (index .ls 0) "nope" | default "z" %>`)
	add(`<% required "msg" "x" %>|<% required "only" %>`)
	add(`<% required %>`)
	add(`<% required "a" "b" "c" %>`)
	add(`<% required "m" nil %>`)
	add(`<% required "m" (index .m "zz") %>`)
	add(`<% required "m" (index .m "a") %>`)
	add(`<% index .m "zz" | required "m: set it" %>`)
	add(`<% index .m "b" | required "m: set it" | data.ToJSON %>`)
	add(`<% required 5 "" %>`)
	add(`<% required "" 0 %>`)
	add(`<% required "" "" %>`)

	// data
	add(`<% data.ToJSON "é<>&\"\\/\n\t " %>`)
	add(`<% data.ToJSON "\x01\x1f" %>`)
	add(`<% data.ToJSON . %>`)
	add(`<% data.ToJSONPretty "\t" . %>`)
	add(`<% data.ToJSONPretty "" .m %>`)
	add(`<% data.ToJSON (dict "b" 1 "a" 2 "B" 3 "_" 4) %>`)
	add(`<% data.ToJSON (coll.Slice 1 2.0 "3" nil true) %>`)
	add(`<% data.ToJSON (strings.Split "," "a,b") %>`)
	add(`<% data.ToJSON (math.Seq 3) %>`)
	add(`<% data.ToJSON (math.Sub 3 1.5) %>|<% data.ToJSON (math.Add 1 2.0) %>`)
	add(`<% data.ToTOML .m %>`)
	add(`<% data.ToTOML .em %>`)
	add(`<% data.ToTOML (dict "a" 1 "b" (dict "c" "x" "d" (coll.Slice 1 2)) "e" (coll.Slice (dict "f" 1) (dict "f" 2)) "g" "s" "h" 1.5 "i" true "j" 3.0 "k" "a\"b\\c\n") %>`)
	add(`<% data.ToTOML (dict "a" (dict "b" (dict "c" 1)) "z" (dict "y" 1) "x" 2) %>`)
	add(`<% data.ToTOML (dict "k" (coll.Slice "a" 1 true)) %>`)
	add(`<% data.ToTOML (dict "k" (coll.Slice)) %>`)
	add(`<% data.ToTOML (dict "k" nil) %>`)
	add(`<% data.ToTOML "x" %>`)
	add(`<% data.ToTOML (dict "a.b" 1 "c d" 2 "e-f" 3 "" 4) %>`)
	add(`<% data.ToTOML (dict "t" (dict "a.b" (dict "c" 1))) %>`)
	add(`<% math.Add "4" 1 | test.Kind %>|<% math.Add "4.5" 1 | test.Kind %>|<% math.Add "4" 1 %>|<% math.Add "4.5" 1 %>`)
	add(`<% data.ToTOML (dict "tools" (dict "node" "22" "go" (dict "version" "1.2" "postinstall" "x") "pipx:ruff" "latest") "settings" (dict "idiomatic_version_file" true "jobs" 4 "x" 1.25 "list" (coll.Slice "a" "b")) "env" (dict "A" "b" "PATH" "~/é") "tasks" (dict "t" (dict "run" "echo \"x\"" "depends" (coll.Slice "a"))) "shell_alias" (dict "ll" "ls -l")) %>`)
	add(`<% data.ToTOML (dict "hooks" (dict "enter" (coll.Slice (dict "script" "a" "shell" "bash") (dict "script" "b")))) %>`)
	add(`<% data.ToTOML (dict "a" (coll.Slice (coll.Slice 1 2) (coll.Slice "x"))) %>`)
	add(`<% data.ToTOML (dict "big" .big) %>`)
	add(`<% data.ToTOML (dict "n" 3 "f" 3.0 "g" 1e21 "h" 0.000001 "neg" -5) %>`)
	add(`<% data.ToTOML (dict "s" "line1\nline2\ttab\u0001 \"q\" \\ 'sq' é 😀") %>`)
	add(`<% data.JSONArray "[1,2.5,\"a\",[3],{\"k\":null},true]" %>`)
	add(`<% data.JSONArray "[1,2.5,\"a\",[3],{\"k\":null},true]" | data.ToJSON %>`)
	add(`<% data.JSONArray "[]" | data.ToJSON %>`)
	add(`<% data.JSONArray "" %>`)
	add(`<% $p := data.JSONArray "[\"mise\",\"tools\",\"node\"]" %><% index $p 0 %>|<% slice $p 1 | data.ToJSON %>|<% len $p %>`)

	// env, file
	add(`<% env.Getenv "TMPL_T_A" %>|<% env.Getenv "TMPL_T_NOPE" %>|<% env.Getenv "TMPL_T_NOPE" "def" %>|<% env.Getenv "TMPL_T_EMPTY" "def" %>|<% env.Getenv "TMPL_T_A" "def" %>`)
	add(`<% range env.Getenv "TMPL_T_NL" | strings.Split "\n" %>[<% . %>]<% end %>`)
	add(`<% if env.Getenv "TMPL_T_EMPTY" %>y<% else %>n<% end %>|<% if not (env.Getenv "TMPL_T_NOPE") %>n<% end %>`)
	add(`<% env.Getenv 5 %>`)
	add(`<% env.Getenv "TMPL_T_A" "a" "b" %>`)
	add(`<% file.Read "testdata/hello.txt" %>`)
	add(`<% file.Read "testdata/hello.txt" | strings.TrimPrefix "hello\n" %>`)
	add(`<% file.Read (printf "%s/hello.txt" "testdata") | strings.TrimSpace %>`)
	add(`<% file.Read (print "testdata" "/hello.txt") | strings.ToUpper %>`)

	// math
	for _, a := range []string{"3", "0", "-2", "2.5", "3.0", `"4"`, `"x"`, "true", "nil", ".n", ".z", ".f", ".s", ".big", ".l"} {
		add(`<% math.Add ` + a + ` 1 %>|<% math.Add 1 ` + a + ` %>`)
		add(`<% math.Sub ` + a + ` 1 %>`)
		add(`<% math.Sub 5 ` + a + ` %>`)
		add(`<% math.Seq ` + a + ` %>`)
		add(`<% math.Seq ` + a + ` 3 %>`)
		add(`<% math.Sub ` + a + ` 1 | test.Kind %>`)
	}
	add(`<% math.Add %>|<% math.Add 1 %>|<% math.Add 1 2 3 4 %>`)
	add(`<% math.Sub 1 %>`)
	add(`<% math.Sub 1 2 3 %>`)
	add(`<% math.Seq %>`)
	add(`<% math.Seq 1 2 3 4 %>`)
	for _, s := range []string{"0 0", "3 1", "1 3", "5 5", "0 3", "-2 2", "2 -2", "1 5 2", "5 1 2", "1 5 -1", "5 1 -2", "1 5 0", "1 6 2", "10 0 3", "0 10 4"} {
		add(`<% math.Seq ` + s + ` %>`)
		add(`<% math.Seq ` + s + ` | data.ToJSON %>`)
	}
	add(`<% range math.Seq (math.Sub (len .strs) 1) 0 %>[<% . %>]<% end %>`)
	add(`<% range math.Seq 0 0 %>[<% . %>]<% end %>`)
	add(`<% $last := math.Sub (len .strs) 1 %><% range $i, $m := .strs %><% if $i %><% if eq $i $last %> and <% else %>, <% end %><% end %><% $m %><% end %>`)
	add(`<% $i := 0 %><% index .strs (math.Add $i 1) %>`)
	add(`<% index .strs (math.Sub 2 1) %>`)

	// regexp
	for _, a := range []string{`"^~" "/h"`, `"(a)" "[$1]"`, `"(?i)A" "x"`, `"a|b" ""`, `"(" "x"`, `"" "-"`, `"^~" "$1"`, `"x" "${0}y"`, `"[0-9]+" "N"`} {
		add(`<% regexp.Replace ` + a + ` "banana 123 ~a" %>`)
		add(`<% regexp.Replace ` + a + ` .path %>`)
		add(`<% .path | regexp.Replace ` + a + ` %>`)
	}
	add(`<% regexp.Replace "a" "x" 3 %>|<% regexp.Replace "1" "x" 1.5 %>|<% regexp.Replace "n" "x" nil %>`)
	add(`<% index .m "zz" | required "set it" | regexp.Replace "^~" .home.HOME %>`)
	add(`<% .path | regexp.Replace "^~" $.home.HOME %>`)
	add(`<% regexp.Replace "a" %>`)

	// strings
	for _, a := range []string{`"abc"`, `""`, `123`, `1.5`, `nil`, `true`, `.s`, `.nil`, `.l`, `.m`, ` "  x  \n" `, `"a,b,,c"`, `"héllo"`} {
		add(`<% strings.Contains "b" ` + a + ` %>|<% strings.Contains "" ` + a + ` %>`)
		add(`<% strings.HasPrefix "a" ` + a + ` %>|<% strings.HasPrefix "" ` + a + ` %>`)
		add(`<% strings.TrimPrefix "a" ` + a + ` %>`)
		add(`<% strings.ReplaceAll "b" "X" ` + a + ` %>`)
		add(`<% strings.Split "," ` + a + ` | data.ToJSON %>`)
		add(`<% strings.Split "" ` + a + ` | data.ToJSON %>`)
		add(`<% strings.SplitN "," 2 ` + a + ` | data.ToJSON %>`)
		add(`<% strings.SplitN "," 1 ` + a + ` | data.ToJSON %>`)
		add(`<% strings.SplitN "," 0 ` + a + ` | data.ToJSON %>`)
		add(`<% strings.SplitN "," -1 ` + a + ` | data.ToJSON %>`)
		add(`<% strings.ToUpper ` + a + ` %>`)
		add(`<% strings.TrimSpace ` + a + ` %>`)
		add(`<% strings.Quote ` + a + ` %>`)
	}
	add(`<% strings.Contains 3 "abc" %>`)
	add(`<% strings.Split "," %>`)
	add(`<% strings.SplitN "," "x" "a,b" %>`)
	add(`<% strings.Quote "tab\there \"q\" \\ é   \x7f   😀" %>`)
	add(`<% strings.Split "\n" "a\nb\n" | data.ToJSON %>`)
	add(`<% strings.Split " " (strings.TrimSpace "") | data.ToJSON %>`)
	add(`<% range $k, $v := strings.SplitN "=" 2 "a=b=c" %><% $k %>:<% $v %>;<% end %>`)
	add(`<% printf " %s " "dev" | strings.Contains "x" %>|<% strings.Contains (printf " %s " "dev") " dev ci " %>`)

	// tmpl.Exec and define
	add(`<% define "x" %>[<% . %>]<% end %><% tmpl.Exec "x" 5 %>|<% tmpl.Exec "x" . %>|<% tmpl.Exec "x" %>`)
	add(`<% tmpl.Exec "nope" 5 %>`)
	add(`<% define "se_has" %><% $v := .v %><% range .p %><% if not (and (test.IsKind "map" $v) (has $v .)) %><% $v = false %><% else %><% $v = index $v . %><% end %><% end %><% if $v %>1<% end %><% end %><% tmpl.Exec "se_has" (dict "v" .m "p" (coll.Slice "b" "c")) %>|<% tmpl.Exec "se_has" (dict "v" .m "p" (coll.Slice "zz")) %>`)
	add(`<% define "a" %>A<% tmpl.Exec "b" . %><% end %><% define "b" %>B<% . %><% end %><% tmpl.Exec "a" 1 %>`)
	add(`<% tmpl.Exec %>`)
	add(`<% template "x" %><% define "x" %>X<% end %>`)

	// Template syntax: whitespace trim, variables, range, with, break.
	add("a\n<%- .s -%>\n b\n<%- /* c */ -%>\nend")
	add("<% range $i, $x := .strs %><% if $i %>,<% end %><% $x %><% end %>")
	add("<% range $k, $v := .m %><% $k %>=<% $v %>;<% end %>")
	add("<% range .em %>x<% else %>empty<% end %>")
	add("<% with .nil %>w<% else %>none<% end %>|<% with .s %>[<% . %>]<% end %>")
	add("<% range .l %><% if eq . nil %><% continue %><% end %><% . %>;<% end %>")
	add("<% range .strs %><% . %><% break %><% end %>")
	add("<% $x := 1 %><% $x = 2 %><% $x %>")
	add("<% and 1 0 %>|<% or 0 \"\" %>|<% not 0 %>|<% and .n .s %>|<% or .e .s %>")
	add("<% eq .n 3 %>|<% eq .z 3 %>|<% eq .f 1.5 %>|<% eq .n 3.0 %>|<% ne .s \"str\" %>|<% lt .n 4 %>|<% gt .f 1 %>|<% gt .n .f %>")
	add("<% eq .n \"3\" %>")
	add("<% eq .s 3 %>")
	add("<% index .strs 5 %>")
	add("<% index .m \"zz\" %>|<% index .m \"b\" \"c\" %>")
	add("<% index .nil \"a\" %>")
	add("<% len .s %>|<% len .strs %>|<% len .m %>")
	add("<% print 1 2 \"a\" \"b\" 3 %>|<% printf \"%q %5.2f %x\" \"s\" 1.5 255 %>|<% println 1 %>")
	add("<% slice .strs 1 %>|<% slice .s 1 2 %>|<% slice (coll.Slice 1 2 3) 1 | data.ToJSON %>")
	add("<% html \"<a>\" %>|<% urlquery \"a b\" %>|<% js \"'\" %>")
	add("<% .strs | len %>")
	add("<% . %>")
	add("<% $ %>")
	add("<% .s.x %>")
	add("<% .nil.x %>")
	add("<% .m.b.c %>|<% index .m.b.c 1 %>")
	add("<% .m.zz %>")
	add("<% .m.zz.y %>")
	add("<% nofunc 1 %>")
	add("<% if %>")
	add("<% end %>")
	add("<% .s ")
	add("no actions {{ mise }} [[ bash ]]")
	add("{{ .s }}")
	add("<%/* c */%>x")
	add("<%- if true -%> a <%- end -%>")
	add("")
	// missing keys
	for _, mk := range []string{"", "error", "default", "zero", "invalid"} {
		addMissing(`<% .nokey %>`, `{}`, mk)
		addMissing(`<% .nokey.b %>`, `{}`, mk)
		addMissing(`<% index . "nokey" %>`, `{}`, mk)
		addMissing(`<% index . "nokey" | default "d" %>`, `{}`, mk)
		addMissing(`<% .nokey | default "d" %>`, `{}`, mk)
		addMissing(`<% has . "nokey" %>`, `{}`, mk)
		addMissing(`<% required "m" .nokey %>`, `{}`, mk)
		addMissing(`<% index . "nokey" | required "m: set it" %>`, `{}`, mk)
		addMissing(`<% if .nokey %>y<% else %>n<% end %>`, `{}`, mk)
		addMissing(`<% .a.nokey %>`, `{"a":{}}`, mk)
		addMissing(`<% data.ToJSON .nokey %>`, `{}`, mk)
		addMissing(`<% $m := dict "a" 1 %><% $m.zz %>`, `{}`, mk)
		addMissing(`<% .l.x %>`, `{"l":[1]}`, mk)
	}
	// Context shapes.
	add(`<% range $k, $v := . %><% $k %>=<% $v %>:<% test.Kind $v %> <% end %>`,
		`{"a":3.0,"b":1e3,"c":-5,"d":12345678901234567890,"e":1.0e0,"f":0.1,"g":9223372036854775807,"h":9223372036854775808,"i":"3","j":1E2,"k":-0.0,"l":0,"m":-9223372036854775808,"n":18446744073709551615,"o":18446744073709551616,"p":1e400,"q":5e-324}`)
	add(`<% . | data.ToJSON %>`,
		`{"a":3.0,"b":1e3,"c":-5,"d":12345678901234567890,"e":1.0e0,"f":0.1,"g":9223372036854775807,"j":1E2,"k":-0.0,"l":0,"m":1e21,"n":1e20,"o":1.5e-7,"p":123456789.123456789,"q":0.000001,"r":1e-7,"s":100000000000000000000.0,"t":0.5,"u":2.50}`)
	add(`<% . | data.ToJSONPretty "  " %>`, `{"a":[1,[2,{"b":null}]],"c":{"d":{}}}`)

	// Number forms by a generator: ToJSON, printf and string forms.
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 40; i++ {
		var f float64
		switch i % 4 {
		case 0:
			f = r.Float64() * 1000
		case 1:
			f = float64(r.Int63n(1 << 40))
		case 2:
			f = r.NormFloat64() * 1e-5
		default:
			f = r.NormFloat64() * 1e22
		}
		lit := fmt.Sprintf("%v", f)
		if !strings.ContainsAny(lit, ".eE") {
			lit += ".0"
		}
		add(`<% . | data.ToJSON %>|<% .x %>|<% test.Kind .x %>|<% math.Add .x 1 %>|<% conv.ToString .x %>`, `{"x":`+lit+`}`)
	}
	// Merge by a generator.
	for i := 0; i < 60; i++ {
		a, b, c := genMap(r, 3), genMap(r, 3), genMap(r, 2)
		add(`<% coll.Merge (index . "a") (index . "b") (index . "c") | data.ToJSON %>`, fmt.Sprintf(`{"a":%s,"b":%s,"c":%s}`, a, b, c))
		add(`<% coll.Merge (index . "b") (index . "a") | data.ToJSON %>|<% index . "a" | data.ToJSON %>`, fmt.Sprintf(`{"a":%s,"b":%s}`, a, b))
	}
	return c
}

// genMap returns a JSON object with random keys and values, with empty
// values, lists and nested objects.
func genMap(r *rand.Rand, depth int) string {
	keys := []string{"a", "b", "c", "d"}
	var parts []string
	for _, k := range keys {
		if r.Intn(3) == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%q:%s", k, genVal(r, depth)))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func genVal(r *rand.Rand, depth int) string {
	vals := []string{`""`, `"x"`, `"y"`, `0`, `1`, `2.5`, `true`, `false`, `null`, `[]`, `[1]`, `[2,3]`, `["a"]`, `{}`}
	n := len(vals)
	if depth > 0 {
		n += 3
	}
	i := r.Intn(n)
	if i < len(vals) {
		return vals[i]
	}
	return genMap(r, depth-1)
}

var goldenPath = filepath.Join("testdata", "funcs.json")

func runGo(c fnCase) (out []byte, err error) {
	ctx, derr := DecodeJSON([]byte(c.Ctx))
	if derr != nil {
		return nil, derr
	}
	return Render("<arg>", []byte(c.Tmpl), ctx, &Options{
		MissingKey: c.Missing,
		Env:        func(k string) string { return testEnv[k] },
	})
}

var errLine = regexp.MustCompile(`"err":"((?:[^"\\]|\\.)*)"`)

// gomplateBin finds the gomplate CLI, or returns "".
var gomplateBin = sync.OnceValue(func() string {
	if p, err := exec.LookPath("gomplate"); err == nil {
		return p
	}
	cmd := exec.Command("mise", "x", "gomplate", "--", "which", "gomplate")
	cmd.Dir = "/"
	cmd.Env = append(os.Environ(), "MISE_OFFLINE=1")
	b, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
})

// runCLI renders the case with gomplate, as _se_gomplate does.
func runCLI(t *testing.T, c fnCase) fnCase {
	t.Helper()
	dir := t.TempDir()
	ctxFile := filepath.Join(dir, "ctx.json")
	if err := os.WriteFile(ctxFile, []byte(c.Ctx), 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"--left-delim", "<%", "--right-delim", "%>"}
	if c.Missing != "" {
		args = append(args, "--missing-key", c.Missing)
	}
	args = append(args, "-c", ".="+ctxFile+"?type=application/json", "-i", c.Tmpl)
	cmd := exec.Command(gomplateBin(), args...)
	wd, _ := os.Getwd()
	cmd.Dir = wd
	cmd.Env = os.Environ()
	for k, v := range testEnv {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	c.OK, c.Out, c.Msg = err == nil, "", ""
	if err == nil {
		c.Out = stdout.String()
		return c
	}
	c.Msg = cliMessage(stderr.String())
	return c
}

var callAfter = regexp.MustCompile(`error calling [A-Za-z.]*: `)

// cliMessage returns the text after the last "error calling X: " of the err
// field of the gomplate log line, or "" if there is none.
func cliMessage(stderr string) string {
	m := errLine.FindStringSubmatch(stderr)
	if m == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal([]byte(`"`+m[1]+`"`), &s); err != nil {
		return ""
	}
	locs := callAfter.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return ""
	}
	return s[locs[len(locs)-1][1]:]
}

func goMessage(err error) string {
	s := err.Error()
	locs := callAfter.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return ""
	}
	return s[locs[len(locs)-1][1]:]
}

func checkCase(t *testing.T, want fnCase) {
	t.Helper()
	if want.Ctx == baseCtx {
		want.Ctx = "(base)"
	}
	got, err := runGo(withCtx(want))
	if want.OK {
		if err != nil {
			t.Errorf("%s\n  ctx: %s\n  want %q\n  got error: %v", want.Tmpl, want.Ctx, want.Out, err)
		} else if string(got) != want.Out {
			t.Errorf("%s\n  ctx: %s\n  missing-key: %q\n  want %q\n  got  %q", want.Tmpl, want.Ctx, want.Missing, want.Out, got)
		}
		return
	}
	if err == nil {
		t.Errorf("%s\n  ctx: %s\n  want an error (%q), got output %q", want.Tmpl, want.Ctx, want.Msg, got)
		return
	}
	if m := goMessage(err); m != want.Msg {
		t.Errorf("%s\n  ctx: %s\n  want message %q\n  got  message %q (%v)", want.Tmpl, want.Ctx, want.Msg, m, err)
	}
}

func TestFuncsGolden(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		bin := gomplateBin()
		if bin == "" {
			t.Fatal("UPDATE_GOLDEN=1 needs the gomplate CLI")
		}
		cases := funcCases()
		out := make([]fnCase, len(cases))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 8)
		for i, c := range cases {
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				out[i] = runCLI(t, c)
			}()
		}
		wg.Wait()
		b, err := json.MarshalIndent(out, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d cases", len(out))
	}
	b, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("%v (run UPDATE_GOLDEN=1 go test ./internal/tmpl)", err)
	}
	var cases []fnCase
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no golden cases")
	}
	for _, c := range cases {
		checkCase(t, c)
	}
	t.Logf("%d function cases equal to gomplate", len(cases))
}

// TestFuncsGoldenIsCurrent makes sure that the golden file lists the cases of
// funcCases, and that it is still the CLI output when the CLI is present.
func TestFuncsGoldenIsCurrent(t *testing.T) {
	b, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Skip("no golden file")
	}
	var golden []fnCase
	if err := json.Unmarshal(b, &golden); err != nil {
		t.Fatal(err)
	}
	cases := funcCases()
	if len(cases) != len(golden) {
		t.Fatalf("golden file has %d cases, funcCases has %d: run UPDATE_GOLDEN=1", len(golden), len(cases))
	}
	for i := range cases {
		if cases[i].Tmpl != golden[i].Tmpl || cases[i].Ctx != golden[i].Ctx || cases[i].Missing != golden[i].Missing {
			t.Fatalf("case %d differs from the golden file: run UPDATE_GOLDEN=1", i)
		}
	}
	if gomplateBin() == "" || os.Getenv("TMPL_NO_LIVE") == "1" {
		t.Skip("gomplate CLI not found")
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := range golden {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			live := runCLI(t, cases[i])
			if live.OK != golden[i].OK || live.Out != golden[i].Out || live.Msg != golden[i].Msg {
				t.Errorf("golden case %d is not the CLI output: %s", i, cases[i].Tmpl)
			}
		}()
	}
	wg.Wait()
}

// withCtx restores the base context that checkCase shortens.
func withCtx(c fnCase) fnCase {
	if c.Ctx == "(base)" {
		c.Ctx = baseCtx
	}
	return c
}
