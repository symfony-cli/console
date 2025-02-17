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
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/symfony-cli/terminal"
)

// Application is the main structure of a cli application.
type Application struct {
	// The name of the program. Defaults to filepath.Base(os.Executable())
	Name string
	// Full name of command for help, defaults to Name
	HelpName string
	// Description of the program.
	Usage string
	// Version of the program
	Version string
	// Channel of the program (dev, beta, stable, ...)
	Channel string
	// Description of the program
	Description string
	// List of commands to execute
	Commands []*Command
	// List of flags to parse
	Flags []Flag
	// Prefix used to automatically find flag in environment
	FlagEnvPrefix []string
	// Categories contains the categorized commands and is populated on app startup
	Categories CommandCategories
	// An action to execute before any subcommands are run, but after the context is ready
	// If a non-nil error is returned, no subcommands are run
	Before BeforeFunc
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFunc
	// The action to execute when no subcommands are specified
	Action ActionFunc
	// Build date
	BuildDate string
	// Copyright of the binary if any
	Copyright string
	// Writer writer to write output to
	Writer io.Writer
	// ErrWriter writes error output
	ErrWriter io.Writer

	setupOnce sync.Once
}

// Run is the entry point to the cli app. Parses the arguments slice and routes
// to the proper flag/args combination
func (a *Application) Run(arguments []string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			HandleExitCoder(WrapPanic(e))
		}
	}()

	a.setupOnce.Do(func() {
		a.setup()
	})

	context := NewContext(a, nil, nil)
	context.flagSet, err = a.parseArgs(arguments[1:])

	a.configureIO(context)

	if err := checkFlagsValidity(a.Flags, context.flagSet, context); err != nil {
		return err
	}

	if err != nil {
		err = IncorrectUsageError{err}
		ShowAppHelp(context)
		fmt.Fprintln(a.Writer)
		HandleExitCoder(err)
		return err
	}

	defer func() {
		if a.After != nil {
			if afterErr := a.After(context); afterErr != nil {
				if err != nil {
					err = newMultiError(err, afterErr)
				} else {
					err = afterErr
				}
			}
		}
	}()

	args := context.Args()
	if args.Present() {
		name := args.first()
		context.Command = a.BestCommand(name)
	}

	if a.Before != nil {
		beforeErr := a.Before(context)
		if beforeErr != nil {
			fmt.Fprintf(a.Writer, "%v\n\n", beforeErr)
			ShowAppHelp(context)
			HandleExitCoder(beforeErr)
			err = beforeErr
			return err
		}
	}

	if checkHelp(context) {
		err := ShowAppHelpAction(context)
		HandleExitCoder(err)
		return err
	}

	if checkVersion(context) {
		ShowVersion(context)
		return nil
	}

	if c := context.Command; c != nil {
		err = c.Run(context)
	} else {
		err = a.Action(context)
	}
	HandleExitCoder(err)
	return err
}

// Command returns the named command on App. Returns nil if the command does not
// exist
func (a *Application) Command(name string) *Command {
	for _, c := range a.Commands {
		if c.HasName(name, true) {
			c.UserName = name
			return c
		}
	}
	return nil
}

// BestCommand returns the named command on App or a command fuzzy matching if
// there is only one. Returns nil if the command does not exist of if the fuzzy
// matching find more than one.
func (a *Application) BestCommand(name string) *Command {
	name = strings.ToLower(name)
	if c := a.Command(name); c != nil {
		return c
	}

	// fuzzy match?
	var matches []*Command
	for _, c := range a.Commands {
		if c.HasName(name, false) {
			matches = append(matches, c)
		}
	}
	if len(matches) == 1 {
		matches[0].UserName = name
		return matches[0]
	}
	return nil
}

// Category returns the named CommandCategory on App. Returns nil if the category does not exist
func (a *Application) Category(name string) *CommandCategory {
	name = strings.ToLower(name)
	if a.Categories == nil {
		return nil
	}

	for _, c := range a.Categories.Categories() {
		if c.Name() == name {
			return &c
		}
	}

	return nil
}

// VisibleCategories returns a slice of categories and commands that are
// Hidden=false
func (a *Application) VisibleCategories() []CommandCategory {
	ret := []CommandCategory{}
	for _, category := range a.Categories.Categories() {
		if len(category.VisibleCommands()) > 0 {
			ret = append(ret, category)
		}
	}
	return ret
}

// VisibleCommands returns a slice of the Commands with Hidden=false
func (a *Application) VisibleCommands() []*Command {
	ret := []*Command{}
	for _, command := range a.Commands {
		if command.Hidden == nil || !command.Hidden() {
			ret = append(ret, command)
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})

	return ret
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (a *Application) VisibleFlags() []Flag {
	return visibleFlags(a.Flags)
}

// setup runs initialization code to ensure all data structures are ready for
// `Run` or inspection prior to `Run`.
func (a *Application) setup() {
	if a.BuildDate == "" {
		a.BuildDate = time.Now().Format(time.RFC3339)
	}

	if a.Name == "" {
		a.Name = CurrentBinaryName()
	}

	if a.HelpName == "" {
		a.HelpName = CurrentBinaryName()
	}

	if a.Usage == "" {
		a.Usage = "A new cli application"
	}

	if a.Version == "" {
		a.Version = "0.0.0"
	}

	if a.Channel == "" {
		a.Channel = "dev"
	}

	if a.Action == nil {
		a.Action = helpCommand.Action
	}

	if a.Writer == nil {
		a.Writer = terminal.Stdout
	}
	if a.ErrWriter == nil {
		a.ErrWriter = terminal.Stderr
	}

	a.prependFlag(VersionFlag)

	if LogLevelFlag != nil && LogLevelFlag.Name != "" {
		a.prependFlag(LogLevelFlag)
	}

	if QuietFlag != nil && QuietFlag.Name != "" {
		a.prependFlag(QuietFlag.ForApp(a))
	}

	if NoInteractionFlag != nil && NoInteractionFlag.Name != "" {
		a.prependFlag(NoInteractionFlag)
	}

	if AnsiFlag != nil {
		a.prependFlag(AnsiFlag)
	}

	if NoAnsiFlag != nil && NoAnsiFlag.Name != "" {
		a.prependFlag(NoAnsiFlag)
	}

	if a.Command(helpCommand.Name) == nil && (helpCommand.Hidden == nil || !helpCommand.Hidden()) {
		a.Commands = append([]*Command{helpCommand}, a.Commands...)
		// This command is global and as such is mutated by tests so we reset
		// the flags to ensure a consistent behaviour
		helpCommand.Flags = nil
	}

	if a.Command(versionCommand.Name) == nil && (versionCommand.Hidden == nil || !versionCommand.Hidden()) {
		a.Commands = append([]*Command{versionCommand}, a.Commands...)
		// This command is global and as such is mutated by tests so we reset
		// the flags to ensure a consistent behaviour
		helpCommand.Flags = nil
	}

	if HelpFlag != nil {
		a.prependFlag(HelpFlag)
	}

	registerAutocompleteCommands(a)

	for _, c := range a.Commands {
		c.normalizeCommandNames()
		if c.HelpName == "" {
			c.HelpName = fmt.Sprintf("%s %s", a.HelpName, c.FullName())
		}
		checkFlagsUnicity(a.Flags, c.Flags, c.FullName())
		checkArgsModes(c.Args)
	}

	a.Categories = newCommandCategories()
	for _, command := range a.Commands {
		a.Categories.AddCommand(command.Category, command)
	}
	sort.Sort(a.Categories.(*commandCategories))
}

func (a *Application) prependFlag(fl Flag) {
	if !hasFlag(a.Flags, fl) {
		a.Flags = append([]Flag{fl}, a.Flags...)
	}
}
