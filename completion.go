//go:build darwin || linux || freebsd || openbsd

package console

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/posener/complete/v2"
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
	cmd := complete.Command{
		Flags: map[string]complete.Predictor{},
		Sub:   map[string]*complete.Command{},
	}

	// transpose registered commands and flags to posener/complete equivalence
	for _, command := range c.App.VisibleCommands() {
		subCmd := command.convertToPosenerCompleteCommand(c)

		for _, name := range command.Names() {
			cmd.Sub[name] = &subCmd
		}
	}

	for _, f := range c.App.VisibleFlags() {
		if vf, ok := f.(*verbosityFlag); ok {
			vf.addToPosenerFlags(c, cmd.Flags)
			continue
		}

		predictor := ContextPredictor{f, c}

		for _, name := range f.Names() {
			name = fmt.Sprintf("%s%s", prefixFor(name), name)
			cmd.Flags[name] = predictor
		}
	}

	cmd.Complete(c.App.HelpName)
	return nil
}

func (c *Command) convertToPosenerCompleteCommand(ctx *Context) complete.Command {
	command := complete.Command{
		Flags: map[string]complete.Predictor{},
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

func (c *Command) PredictArgs(ctx *Context, prefix string) []string {
	if c.ShellComplete != nil {
		return c.ShellComplete(ctx, prefix)
	}

	return nil
}

type Predictor interface {
	PredictArgs(*Context, string) []string
}

// ContextPredictor determines what terms can follow a command or a flag
// It is used for autocompletion, given the last word in the already completed
// command line, what words can complete it.
type ContextPredictor struct {
	predictor Predictor
	ctx       *Context
}

// Predict invokes the predict function and implements the Predictor interface
func (p ContextPredictor) Predict(prefix string) []string {
	return p.predictor.PredictArgs(p.ctx, prefix)
}
