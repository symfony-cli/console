//go:build !darwin && !linux && !freebsd && !openbsd
// +build !darwin,!linux,!freebsd,!openbsd

package console

const HasAutocompleteSupport = false

func IsAutocomplete(c *Command) bool {
	return false
}

func registerAutocompleteCommands(a *Application) {
}
