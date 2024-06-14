//go:build darwin || linux || freebsd || openbsd

package console

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/posener/complete"
	"github.com/rs/zerolog"
	"github.com/symfony-cli/terminal"
)

func init() {
	for _, key := range []string{"COMP_LINE", "COMP_POINT", "COMP_DEBUG"} {
		if _, hasEnv := os.LookupEnv(key); hasEnv {
			// Disable Garbage collection for faster autocompletion
			debug.SetGCPercent(-1)
			return
		}
	}
}

var autoCompleteCommand = &Command{
	Category:    "self",
	Name:        "autocomplete",
	Description: "Internal command to provide shell completion suggestions",
	Hidden:      Hide,
	FlagParsing: FlagParsingSkippedAfterFirstArg,
	Args: ArgDefinition{
		&Arg{
			Slice:    true,
			Optional: true,
		},
	},
	Action: AutocompleteAppAction,
}

func registerAutocompleteCommands(a *Application) {
	if IsGoRun() {
		return
	}

	a.Commands = append(
		[]*Command{shellAutoCompleteInstallCommand, autoCompleteCommand},
		a.Commands...,
	)
}

func AutocompleteAppAction(c *Context) error {
	// connect posener/complete logger to our logging facilities
	logger := terminal.Logger.WithLevel(zerolog.DebugLevel)
	complete.Log = func(format string, args ...interface{}) {
		logger.Msgf("completion | "+format, args...)
	}

	cmd := complete.Command{
		GlobalFlags: make(complete.Flags),
		Sub:         make(complete.Commands),
	}

	// transpose registered commands and flags to posener/complete equivalence
	for _, command := range c.App.Commands {
		subCmd := command.convertToPosenerCompleteCommand(c)

		if command.Hidden == nil || !command.Hidden() {
			cmd.Sub[command.FullName()] = subCmd
		}
		for _, alias := range command.Aliases {
			if !alias.Hidden {
				cmd.Sub[alias.String()] = subCmd
			}
		}
	}

	for _, f := range c.App.VisibleFlags() {
		if vf, ok := f.(*verbosityFlag); ok {
			vf.addToPosenerFlags(c, cmd.GlobalFlags)
			continue
		}

		predictor := ContextPredictor{f, c}

		for _, name := range f.Names() {
			name = fmt.Sprintf("%s%s", prefixFor(name), name)
			cmd.GlobalFlags[name] = predictor
		}
	}

	if !complete.New(c.App.HelpName, cmd).Complete() {
		return errors.New("Could not run auto-completion")
	}

	return nil
}

func (c *Command) convertToPosenerCompleteCommand(ctx *Context) complete.Command {
	command := complete.Command{
		Flags: make(complete.Flags, 0),
	}

	for _, f := range c.VisibleFlags() {
		for _, name := range f.Names() {
			name = fmt.Sprintf("%s%s", prefixFor(name), name)
			command.Flags[name] = ContextPredictor{f, ctx}
		}
	}

	if len(c.Args) > 0 || c.ShellComplete != nil {
		command.Args = ContextPredictor{c, ctx}
	}

	return command
}

func (c *Command) PredictArgs(ctx *Context, a complete.Args) []string {
	if c.ShellComplete != nil {
		return c.ShellComplete(ctx, a)
	}

	return nil
}

type Predictor interface {
	PredictArgs(*Context, complete.Args) []string
}

// ContextPredictor determines what terms can follow a command or a flag
// It is used for autocompletion, given the last word in the already completed
// command line, what words can complete it.
type ContextPredictor struct {
	predictor Predictor
	ctx       *Context
}

// Predict invokes the predict function and implements the Predictor interface
func (p ContextPredictor) Predict(a complete.Args) []string {
	return p.predictor.PredictArgs(p.ctx, a)
}
