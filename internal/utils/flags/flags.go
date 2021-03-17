package flags

import "github.com/spf13/pflag"

// MarkHidden marks the specified flag as hidden from the provided flag set
// TODO(REALMC-8369): this method should go away if/when we can get
// golangci-lint to play nicely with errcheck and our exclude .errcheck file
// For now, we use this to isolate and minimize the nolint directives in this repo
func MarkHidden(fs *pflag.FlagSet, name string) {
	fs.MarkHidden(name) //nolint: errcheck
}
