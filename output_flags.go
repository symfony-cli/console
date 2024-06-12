/*
 * Copyright (c) 2021-present Fabien Potencier <fabien@symfony.com>
 *
 * This file is part of Symfony CLI project
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package console

import (
	"flag"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/posener/complete"
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
		terminal.Stdout = terminal.NewBufferedConsoleOutput(io.Discard, io.Discard)
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

func (f *quietFlag) PredictArgs(c *Context, a complete.Args) []string {
	return []string{"true", "false", ""}
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
	if IsAutocomplete(c.Command) {
		terminal.DefaultStdout.SetDecorated(false)
		terminal.Stdin.SetInteractive(false)
		return
	}

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
