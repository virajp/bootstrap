package format

import (
	"bytes"
	"os"
	"sync"

	yqcmd "github.com/mikefarah/yq/v4/cmd"
)

// The engine runs the command of the yq CLI (package cmd of yq v4) in the same
// process, so the output has the bytes of `yq`, with its comments. The package
// cmd keeps its flags in variables of the package, so only one call runs at a
// time.
var yqMu sync.Mutex

// yq runs `yq <sub> -M <args>`. <sub> is `eval` or `ea` (eval-all). <env> is a
// list of KEY=VALUE pairs for the call, as `KEY=VALUE yq ...` in bash. It gives
// the output and the error (nil when the exit code is 0).
func yq(env []string, sub string, args ...string) (res yqResult, err error) {
	yqMu.Lock()
	defer yqMu.Unlock()
	for _, kv := range env {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				old, had := os.LookupEnv(kv[:i])
				_ = os.Setenv(kv[:i], kv[i+1:])
				defer func(k, o string, had bool) {
					if had {
						_ = os.Setenv(k, o)
					} else {
						_ = os.Unsetenv(k)
					}
				}(kv[:i], old, had)
				break
			}
		}
	}
	var out, errb, all bytes.Buffer
	c := yqcmd.New()
	// The package cmd sets its variable unwrapScalar to true for a YAML output
	// and never sets it back, so each call sets the flag as the CLI would have.
	unwrap := "-r=false"
	for i, a := range args {
		if a == "-o" && i+1 < len(args) && args[i+1] == "yaml" {
			unwrap = "-r=true"
		}
	}
	c.SetArgs(append([]string{sub, "-M", unwrap}, args...))
	c.SetOut(&tee{&out, &all})
	c.SetErr(&tee{&errb, &all})
	err = c.Execute()
	return yqResult{out.String(), errb.String(), all.String()}, err
}

// yqResult is the output of one call: stdout, stderr, and both together in
// the order of writing.
type yqResult struct{ stdout, stderr, both string }

// tee writes to two buffers.
type tee struct{ a, b *bytes.Buffer }

func (t *tee) Write(p []byte) (int, error) {
	t.a.Write(p)
	return t.b.Write(p)
}
