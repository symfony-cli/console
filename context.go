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
	"fmt"
)

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific args and
// parsed command-line options.
type Context struct {
	App     *Application
	Command *Command

	flagSet       *flag.FlagSet
	args          *args
	parentContext *Context
}

// NewContext creates a new context. For use in when invoking an App or Command action.
func NewContext(app *Application, set *flag.FlagSet, parentCtx *Context) *Context {
	return &Context{App: app, flagSet: set, parentContext: parentCtx}
}

// Set assigns a value to a context flag.
func (c *Context) Set(name, value string) error {
	if fs := lookupFlagSet(name, c); fs != nil {
		return fs.Set(name, value)
	}

	return fmt.Errorf("no such flag -%v", name)
}

// IsSet determines if the flag was actually set
func (c *Context) IsSet(name string) bool {
	if fs := lookupFlagSet(name, c); fs != nil {
		isSet := false
		fs.Visit(func(f *flag.Flag) {
			if f.Name == name {
				isSet = true
			}
		})
		if isSet {
			return true
		}
	}

	return false
}

// HasFlag determines if a flag is defined in this context and all of its parent
// contexts.
func (c *Context) HasFlag(name string) bool {
	return lookupFlag(name, c) != nil
}

// Lineage returns *this* context and all of its ancestor contexts in order from
// child to parent
func (c *Context) Lineage() []*Context {
	lineage := []*Context{}

	for cur := c; cur != nil; cur = cur.parentContext {
		lineage = append(lineage, cur)
	}

	return lineage
}

// Args returns the command line arguments associated with the context.
func (c *Context) rawArgs() Args {
	v := args{
		values: c.flagSet.Args(),
	}
	return &v
}

func (c *Context) Args() Args {
	// cache args fetch
	if c.args != nil {
		return c.args
	}

	argsValue := make([]string, 0, c.flagSet.NArg())
	for _, arg := range c.flagSet.Args() {
		if arg == "--" {
			continue
		}

		argsValue = append(argsValue, arg)
	}

	c.args = &args{
		values:  argsValue,
		command: c.Command,
	}
	return c.args
}

// NArg returns the number of the command line arguments.
func (c *Context) NArg() int {
	return c.Args().Len()
}

func lookupFlag(name string, ctx *Context) Flag {
	for _, c := range ctx.Lineage() {
		if c.Command == nil {
			continue
		}

		for _, f := range c.Command.Flags {
			for _, n := range f.Names() {
				if n == name {
					return f
				}
			}
		}
	}

	if ctx.App != nil {
		for _, f := range ctx.App.Flags {
			for _, n := range f.Names() {
				if n == name {
					return f
				}
			}
		}
	}

	return nil
}

func lookupFlagSet(name string, ctx *Context) *flag.FlagSet {
	for _, c := range ctx.Lineage() {
		if c.Command != nil {
			name = expandShortcut(c.Command.Flags, name)
		}
		if c.App != nil {
			name = expandShortcut(c.App.Flags, name)
		}
		if f := c.flagSet.Lookup(name); f != nil {
			return c.flagSet
		}
	}

	return nil
}

func lookupRawFlag(name string, ctx *Context) *flag.Flag {
	for _, c := range ctx.Lineage() {
		if c.Command != nil {
			name = expandShortcut(c.Command.Flags, name)
		}
		if c.App != nil {
			name = expandShortcut(c.App.Flags, name)
		}
		if f := c.flagSet.Lookup(name); f != nil {
			return f
		}
	}

	return nil
}
