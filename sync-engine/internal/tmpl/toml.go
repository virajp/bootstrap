package tmpl

import "github.com/BurntSushi/toml"

// DecodeTOML decodes a TOML text as gomplate decodes a TOML datasource (the
// flag -c name=file.toml): a table is a map[string]any, an integer is an
// int64, and an array of tables is a []any of maps.
func DecodeTOML(data []byte) (any, error) {
	var m map[string]any
	if _, err := toml.Decode(string(data), &m); err != nil {
		return nil, err
	}
	return tomlTypes(m), nil
}

// tomlTypes changes the types of the TOML decoder to the types of gomplate.
func tomlTypes(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			x[k] = tomlTypes(e)
		}
		return x
	case []map[string]any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = tomlTypes(e)
		}
		return out
	case []any:
		for i, e := range x {
			x[i] = tomlTypes(e)
		}
		return x
	}
	return v
}
