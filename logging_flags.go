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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/posener/complete"
	"github.com/symfony-cli/terminal"
)

var LogLevelFlag = VerbosityFlag("log-level", "verbose", "v")

type logLevelValue struct{}

func (r *logLevelValue) Set(s string) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return errors.WithStack(err)
	}
	return terminal.SetLogLevel(v)
}

func (r *logLevelValue) Get() interface{} {
	return terminal.GetLogLevel()
}

func (r *logLevelValue) String() string {
	return terminal.Logger.GetLevel().String()
}

type logLevelShortcutValue struct {
	set      *flag.FlagSet
	target   string
	logLevel string
}

func newLogLevelShortcutValue(set *flag.FlagSet, target string, val int) *logLevelShortcutValue {
	return &logLevelShortcutValue{
		set:      set,
		target:   target,
		logLevel: strconv.Itoa(val),
	}
}

func (r *logLevelShortcutValue) IsBoolFlag() bool { return true }

func (r *logLevelShortcutValue) Set(s string) error {
	if s != "" && s != "true" {
		return r.set.Set(r.target, s)
	}

	return r.set.Set(r.target, r.logLevel)
}

func (r *logLevelShortcutValue) String() string {
	return ""
}

type verbosityFlag struct {
	Name         string
	Aliases      []string
	ShortAlias   string
	Usage        string
	DefaultValue int
	DefaultText  string
	Hidden       bool
	EnvVars      []string
	Destination  *logLevelValue
}

func VerbosityFlag(name, alias, shortAlias string) *verbosityFlag {
	return &verbosityFlag{
		Name:        name,
		Aliases:     []string{alias},
		ShortAlias:  shortAlias,
		DefaultText: "",
		Usage:       "Increase the verbosity of messages: 1 for normal output, 2 and 3 for more verbose outputs and 4 for debug",
	}
}

func (f *verbosityFlag) PredictArgs(c *Context, a complete.Args) []string {
	return []string{"1", "2", "3", "4"}
}

func (f *verbosityFlag) Validate(c *Context) error {
	return nil
}

func (f *verbosityFlag) Apply(set *flag.FlagSet) {
	f.DefaultValue = terminal.GetLogLevel()
	f.Destination = &logLevelValue{}

	if f.Name != "" {
		set.Var(f.Destination, f.Name, f.Usage)
	}

	for _, alias := range f.Aliases {
		set.Var(newLogLevelShortcutValue(set, f.Name, 3), alias, "")
	}
	for i := 1; i <= len(terminal.LogLevels)-2; i++ {
		set.Var(newLogLevelShortcutValue(set, f.Name, i+1), strings.Repeat(f.ShortAlias, i), "")
	}
}

// Names returns the names of the flag
func (f *verbosityFlag) Names() []string {
	names := make([]string, 0, len(f.Aliases)+len(terminal.LogLevels)-2)

	if f.Name != "" {
		names = append(names, f.Name)
	}

	names = append(names, f.Aliases...)
	for i := 1; i <= len(terminal.LogLevels)-2; i++ {
		names = append(names, strings.Repeat(f.ShortAlias, i))
	}

	return names
}

// String returns a readable representation of this value (for usage defaults)
func (f *verbosityFlag) String() string {
	_, usage := unquoteUsage(f.Usage)
	names := ""

	for i, n := 1, len(terminal.LogLevels)-2; i <= n; i++ {
		if i == 1 {
			names += prefixFor(f.ShortAlias)
		} else {
			names += "|"
		}
		names += strings.Repeat(f.ShortAlias, i)
	}

	for _, alias := range f.Aliases {
		if alias != "" {
			names += ", " + prefixFor(alias) + alias
		}
	}

	if f.Name != "" {
		names += fmt.Sprintf(", %s%s", prefixFor(f.Name), f.Name)
	}

	return fmt.Sprintf("<info>%s</>\t%s", names, strings.TrimSpace(usage))
}

func (f *verbosityFlag) addToPosenerFlags(c *Context, flags complete.Flags) {
	for i, n := 1, len(terminal.LogLevels)-2; i <= n; i++ {
		name := prefixFor(f.ShortAlias)
		name += strings.Repeat(f.ShortAlias, i)
		flags[name] = complete.PredictFunc(func(a complete.Args) []string {
			return f.PredictArgs(c, a)
		})
	}

	for _, alias := range f.Aliases {
		if alias != "" {
			flags[prefixFor(alias)+alias] = complete.PredictFunc(func(a complete.Args) []string {
				return f.PredictArgs(c, a)
			})
		}
	}

	if f.Name != "" {
		flags[prefixFor(f.Name)+f.Name] = complete.PredictFunc(func(a complete.Args) []string {
			return f.PredictArgs(c, a)
		})
	}
}
