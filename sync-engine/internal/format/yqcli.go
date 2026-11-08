package format

// YQ runs `yq <args>` (the eval command of the yq CLI) in the same process,
// as `_se_yq` of `common` runs it. It gives stdout, stderr and the error (nil
// when the exit code is 0). Only one call runs at a time.
func YQ(args ...string) (stdout, stderr string, err error) {
	r, err := yq(nil, "eval", args...)
	return r.stdout, r.stderr, err
}
