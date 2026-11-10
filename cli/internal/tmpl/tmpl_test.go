package tmpl

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecodeJSONTypes(t *testing.T) {
	v, err := DecodeJSON([]byte(`{"i":3,"f":1.5,"w":3.0,"u":18446744073709551615,"s":"x","l":[1,{"k":2}],"n":null,"t":true,"big":1e400}`))
	if err != nil {
		t.Fatal(err)
	}
	m := v.(map[string]any)
	want := map[string]any{
		"i": 3, "f": 1.5, "w": 3.0, "u": uint64(18446744073709551615), "s": "x",
		"l": []any{1, map[string]any{"k": 2}}, "n": nil, "t": true, "big": "1e400",
	}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("got %#v", m)
	}
	if _, err := DecodeJSON([]byte(`{} {}`)); err == nil {
		t.Error("two values must fail")
	}
	if _, err := DecodeJSON([]byte(`{`)); err == nil {
		t.Error("bad JSON must fail")
	}
}

func TestRenderOptions(t *testing.T) {
	out, err := Render("t", []byte(`<% .a %>|{{ .b }}|<% .b %>`), map[string]any{"a": 1}, nil)
	if err == nil {
		t.Fatalf("a missing key must fail by default, got %q", out)
	}
	out, err = Render("t", []byte(`<% .a %>|{{ x }}|[[ c ]]|<% .zz %>`), map[string]any{"a": 1}, &Options{MissingKey: "default"})
	if err != nil || string(out) != "1|{{ x }}|[[ c ]]|<no value>" {
		t.Errorf("got %q, %v", out, err)
	}
	out, err = Render("t", []byte(`<% env.Getenv "K" %>`), nil, &Options{Env: func(string) string { return "v" }})
	if err != nil || string(out) != "v" {
		t.Errorf("got %q, %v", out, err)
	}
	if _, err := Render("t", []byte(`<% if %>`), nil, nil); err == nil {
		t.Error("a parse error must fail")
	}
}

func TestMessage(t *testing.T) {
	_, err := Render("t", []byte(`<% required "set it" "" %>`), map[string]any{}, nil)
	if err == nil || Message(err) != "set it" {
		t.Errorf("got %v, message %q", err, Message(err))
	}
	_, err = Render("t", []byte(`<% test.Fail "boom" %>`), map[string]any{}, nil)
	if got := Message(err); got != "template generation failed: boom" {
		t.Errorf("got %q", got)
	}
	_, err = Render("t", []byte(`<% .x %>`), map[string]any{}, nil)
	if got := Message(err); !strings.Contains(got, `map has no entry for key "x"`) {
		t.Errorf("got %q", got)
	}
}
