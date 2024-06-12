//go:build darwin || linux || freebsd || openbsd
// +build darwin linux freebsd openbsd

package console

const SupportsAutocomplete = true

func IsAutocomplete(c *Command) bool {
	return c == autoCompleteCommand
}
