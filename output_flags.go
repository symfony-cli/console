package console

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/symfony-cli/terminal"
)

type quietValue struct {
	app *Application
}

func (r *quietValue) Set(s string) error {
	quiet, err := strconv.ParseBool(s)
	if err != nil {
		return errors.WithStack(err)
	}
	if quiet {
		terminal.Stdout = terminal.NewBufferedConsoleOutput(ioutil.Discard, ioutil.Discard)
	} else {
		terminal.Stdout = terminal.DefaultStdout
	}
	terminal.Stdin.SetInteractive(!quiet)
	terminal.Stderr = terminal.Stdout.Stderr

	if r.app != nil {
		r.app.Writer = terminal.Stdout
		r.app.ErrWriter = terminal.Stderr
	}

	return nil
}

func (r *quietValue) Get() interface{} {
	return terminal.Stdout.Stdout.IsQuiet()
}

func (r *quietValue) String() string {
	return strconv.FormatBool(r.Get().(bool))
}

func (r *quietValue) IsBoolFlag() bool {
	return true
}

var QuietFlag = newQuietFlag("quiet", "q")

type quietFlag struct {
	Name    string
	Aliases []string
	Usage   string
	Hidden  bool

	app *Application
}

func newQuietFlag(name string, aliases ...string) *quietFlag {
	return &quietFlag{
		Name:    name,
		Aliases: aliases,
		Usage:   "Do not output any message",
	}
}

func (f *quietFlag) ForApp(app *Application) *quietFlag {
	return &quietFlag{
		Name:    f.Name,
		Aliases: f.Aliases,
		Usage:   f.Usage,
		Hidden:  f.Hidden,
		app:     app,
	}
}

func (f *quietFlag) Validate(c *Context) error {
	return nil
}

func (f *quietFlag) Apply(set *flag.FlagSet) {
	set.Var(&quietValue{f.app}, f.Name, f.Usage)
}

// Names returns the names of the flag
func (f *quietFlag) Names() []string {
	return flagNames(f)
}

// String returns a readable representation of this value (for usage defaults)
func (f *quietFlag) String() string {
	return stringifyFlag(f)
}

var (
	NoInteractionFlag = &BoolFlag{
		Name:  "no-interaction",
		Usage: "Disable all interactions",
	}
	NoAnsiFlag = &BoolFlag{
		Name:  "no-ansi",
		Usage: "Disable ANSI output",
	}
	AnsiFlag = &BoolFlag{
		Name:  "ansi",
		Usage: "Force ANSI output",
	}
)

func (app *Application) configureIO(c *Context) {
	if c.IsSet(AnsiFlag.Name) {
		terminal.DefaultStdout.SetDecorated(c.Bool(AnsiFlag.Name))
	} else if c.IsSet(NoAnsiFlag.Name) {
		terminal.DefaultStdout.SetDecorated(!c.Bool(NoAnsiFlag.Name))
	} else if _, isPresent := os.LookupEnv("NO_COLOR"); isPresent {
		terminal.DefaultStdout.SetDecorated(false)
	}

	if c.IsSet(NoInteractionFlag.Name) {
		terminal.Stdin.SetInteractive(!c.Bool(NoInteractionFlag.Name))
	} else if !terminal.IsInteractive(terminal.Stdin) {
		terminal.Stdin.SetInteractive(false)
	}
}
