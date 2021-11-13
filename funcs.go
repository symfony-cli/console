package console

// BeforeFunc is an action to execute before any subcommands are run, but after
// the context is ready if a non-nil error is returned, no subcommands are run
type BeforeFunc func(*Context) error

// AfterFunc is an action to execute after any subcommands are run, but after the
// subcommand has finished it is run even if Action() panics
type AfterFunc func(*Context) error

// ActionFunc is the action to execute when no subcommands are specified
type ActionFunc func(*Context) error

// CommandNotFoundFunc is executed if the proper command cannot be found
type CommandNotFoundFunc func(*Context, string) error

// FlagStringFunc is used by the help generation to display a flag, which is
// expected to be a single line.
type FlagStringFunc func(Flag) string
