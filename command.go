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
	"regexp"
	"strings"
)

type Alias struct {
	Name   string
	Hidden bool
}

func (a *Alias) String() string {
	return a.Name
}

// Command is a subcommand for a console.App.
type Command struct {
	// The name of the command
	Name string
	// A list of aliases for the command
	Aliases []*Alias
	// A short description of the usage of this command
	Usage string
	// A longer explanation of how the command works
	Description string
	// or a function responsible to render the description
	DescriptionFunc DescriptionFunc
	// The category the command is part of
	Category string
	// The function to call when checking for shell command completions
	ShellComplete ShellCompleteFunc
	// An action to execute before any sub-subcommands are run, but after the context is ready
	// If a non-nil error is returned, no sub-subcommands are run
	Before BeforeFunc
	// An action to execute after any subcommands are run, but after the subcommand has finished
	// It is run even if Action() panics
	After AfterFunc
	// The function to call when this command is invoked
	Action ActionFunc
	// List of flags to parse
	Flags []Flag
	// List of args to parse
	Args ArgDefinition
	// Treat all flags as normal arguments if true
	FlagParsing FlagParsingMode
	// Boolean to hide this command from help
	Hidden func() bool
	// Full name of command for help, defaults to full command name, including parent commands.
	HelpName string
	// The name used on the CLI by the user
	UserName string
}

func Hide() bool {
	return true
}

func (c *Command) normalizeCommandNames() {
	c.Category = strings.ToLower(c.Category)
	c.Name = strings.ToLower(c.Name)
	c.HelpName = strings.ToLower(c.HelpName)
	for _, alias := range c.Aliases {
		alias.Name = strings.ToLower(alias.Name)
	}
}

// FullName returns the full name of the command.
// For subcommands this ensures that parent commands are part of the command path
func (c *Command) FullName() string {
	if c.Category != "" {
		return strings.Join([]string{c.Category, c.Name}, ":")
	}
	return c.Name
}

func (c *Command) PreferredName() string {
	name := c.FullName()
	if name == "" && len(c.Aliases) > 0 {
		names := []string{}
		for _, a := range c.Aliases {
			if name := a.String(); name != "" {
				names = append(names, a.String())
			}
		}
		return strings.Join(names, ", ")
	}
	return name
}

// Run invokes the command given the context, parses ctx.Args() to generate command-specific flags
func (c *Command) Run(ctx *Context) (err error) {
	if HelpFlag != nil {
		// append help to flags
		if !hasFlag(c.Flags, HelpFlag) {
			c.Flags = append(c.Flags, HelpFlag)
		}
	}

	set, err := c.parseArgs(ctx.rawArgs().Tail(), ctx.App.FlagEnvPrefix)
	context := NewContext(ctx.App, set, ctx)
	context.Command = c
	if err == nil {
		err = checkFlagsValidity(c.Flags, set, context)
	}
	if err == nil {
		err = checkRequiredArgs(c, context)
	}
	if err != nil {
		ShowCommandHelp(ctx, c.FullName())
		fmt.Fprintln(ctx.App.Writer)
		return IncorrectUsageError{err}
	}

	if checkCommandHelp(context, c.FullName()) {
		return nil
	}

	if c.After != nil {
		defer func() {
			afterErr := c.After(context)
			if afterErr != nil {
				HandleExitCoder(err)
				if err != nil {
					err = newMultiError(err, afterErr)
				} else {
					err = afterErr
				}
			}
		}()
	}

	if c.Before != nil {
		err = c.Before(context)
		if err != nil {
			ShowCommandHelp(ctx, c.FullName())
			HandleExitCoder(err)
			return err
		}
	}

	err = c.Action(context)
	if err != nil {
		HandleExitCoder(err)
	}
	return err
}

// Names returns the names including short names and aliases.
func (c *Command) Names() []string {
	names := []string{}
	if name := c.FullName(); name != "" {
		names = append(names, name)
	}
	for _, a := range c.Aliases {
		if a.Hidden {
			continue
		}
		if name := a.String(); name != "" {
			names = append(names, a.String())
		}
	}

	return names
}

// HasName returns true if Command.Name matches given name
func (c *Command) HasName(name string, exact bool) bool {
	possibilities := []string{}
	if c.Category != "" {
		possibilities = append(possibilities, strings.Join([]string{c.Category, c.Name}, ":"))
	} else {
		possibilities = append(possibilities, c.Name)
	}
	for _, alias := range c.Aliases {
		possibilities = append(possibilities, alias.String())
	}
	for _, p := range possibilities {
		if p == name {
			return true
		}
	}
	if exact {
		return false
	}

	parts := strings.Split(name, ":")
	for i, part := range parts {
		parts[i] = regexp.QuoteMeta(part)
	}
	re := regexp.MustCompile("^" + strings.Join(parts, "[^:]*:") + "[^:]*$")
	for _, p := range possibilities {
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

// Arguments returns a slice of the Arguments
func (c *Command) Arguments() ArgDefinition {
	return ArgDefinition(c.Args)
}

// VisibleFlags returns a slice of the Flags with Hidden=false
func (c *Command) VisibleFlags() []Flag {
	return visibleFlags(c.Flags)
}
