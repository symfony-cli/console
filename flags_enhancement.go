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
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/symfony-cli/terminal"
)

// FlagParsingMode defines how arguments and flags parsing is done.
// Three different values: FlagParsingNormal, FlagParsingSkipped and
// FlagParsingSkippedAfterFirstArg.
type FlagParsingMode int

const (
	// FlagParsingNormal sets parsing to a normal mode, complete parsing of all
	// flags found after command name
	FlagParsingNormal FlagParsingMode = iota
	// FlagParsingSkipped sets parsing to a mode where parsing stops just after
	// the command name.
	FlagParsingSkipped
	// FlagParsingSkippedAfterFirstArg sets parsing to a hybrid mode where parsing continues
	// after the command name until an argument is found.
	// This for example allows usage like `blackfire -v=4 run --samples=2 php foo.php`
	FlagParsingSkippedAfterFirstArg
)

func (mode FlagParsingMode) IsPostfix() bool {
	return mode == FlagParsingNormal
}

func (mode FlagParsingMode) IsPrefix() bool {
	return mode != FlagParsingNormal
}

func (app *Application) parseArgs(arguments []string) (*flag.FlagSet, error) {
	fs, err := parseArgs(app.fixArgs(arguments), flagSet(app.Name, app.Flags))
	if err != nil {
		return fs, errors.WithStack(err)
	}

	parseFlagsFromEnv(app.FlagEnvPrefix, app.Flags, fs)

	// We expand "~" for each provided string flag
	fs.Visit(expandHomeInFlagsValues)

	err = errors.WithStack(checkRequiredFlags(app.Flags, fs))

	return fs, err
}

func (app *Application) fixArgs(args []string) []string {
	return fixArgs(args, app.Flags, app.Commands, FlagParsingNormal, "")
}

func (c *Command) parseArgs(arguments []string, prefixes []string) (*flag.FlagSet, error) {
	fs, err := parseArgs(c.fixArgs(arguments), flagSet(c.Name, c.Flags))
	if err != nil {
		return fs, errors.WithStack(err)
	}

	parseFlagsFromEnv(prefixes, c.Flags, fs)

	// We expand "~" for each provided string flag
	fs.Visit(expandHomeInFlagsValues)

	err = errors.WithStack(checkRequiredFlags(c.Flags, fs))

	return fs, err
}

func (c *Command) fixArgs(args []string) []string {
	return fixArgs(args, c.Flags, nil, c.FlagParsing, "--")
}

func parseArgs(arguments []string, fs *flag.FlagSet) (*flag.FlagSet, error) {
	fs.SetOutput(io.Discard)
	err := errors.WithStack(fs.Parse(arguments))
	if err != nil {
		return fs, err
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.WithStack(e.(error))
		}
	}()

	fs.Visit(func(f *flag.Flag) {
		terminal.Logger.Trace().Msgf("Using CLI flags for '%s' configuration entry.\n", f.Name)
	})

	return fs, err
}

func parseFlagsFromEnv(prefixes []string, flags []Flag, fs *flag.FlagSet) {
	definedFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		definedFlags[f.Name] = true
	})

	for _, f := range flags {
		fName := flagName(f)

		// flags given on the CLI overrides environment values
		if _, alreadyThere := definedFlags[fName]; alreadyThere {
			continue
		}

		envVariableNames := flagStringSliceField(f, "EnvVars")

		for _, prefix := range prefixes {
			envVariableNames = append(envVariableNames, strings.ToUpper(strings.Replace(fmt.Sprintf("%s_%s", prefix, fName), "-", "_", -1)))
		}

		// reverse slice order
		for i := len(envVariableNames)/2 - 1; i >= 0; i-- {
			opp := len(envVariableNames) - 1 - i
			envVariableNames[i], envVariableNames[opp] = envVariableNames[opp], envVariableNames[i]
		}

		for _, name := range envVariableNames {
			val := os.Getenv(name)
			if val == "" {
				continue
			}

			terminal.Logger.Trace().Msgf("Using %s from ENV for '%s' configuration entry.\n", name, fName)
			if err := fs.Set(fName, val); err != nil {
				panic(errors.Errorf("Failed to set flag %s with value %s", fName, val))
			}
		}
	}
}

// fixArgs fixes command lines arguments for them to be parsed.
// Examples:
// upload -slot=4 --v="4" file1 file2 will return:
// -slot=4 --v="4" upload file1 file2
//
// --config=$HOME/.blackfire.ini run --reference=1 php -n exception.php --config=foo will return:
// --config=$HOME/.blackfire.ini --reference=1 run php -n exception.php --config=foo
//
// Note in the latter example than this function needs to pay attention to be eager
// and not try to "fix" arguments belonging to a possible embedded command as run.
// For this purpose, you have three different FlagParsing modes available.
// See FlagParsingMode for more information.
func fixArgs(args []string, flagDefs []Flag, cmdDefs []*Command, defaultMode FlagParsingMode, defaultCommand string) []string {
	var (
		flags    = make([]string, 0)
		nonFlags = make([]string, 0)

		command                = defaultCommand
		parsingMode            = defaultMode
		previousFlagNeedsValue = false
	)

	var isFlag = func(name string) bool {
		return len(name) > 1 && name[0] == '-'
	}
	var cleanFlag = func(name string) string {
		if index := strings.Index(name, "="); index != -1 {
			name = name[:index]
		}
		name = strings.TrimLeft(name, "-")

		return expandShortcut(flagDefs, name)
	}
	var translateShortcutFlags = func(name string) string {
		twoDashes := (name[1] == '-')
		name = strings.Trim(name, "-")
		if index := strings.Index(name, "="); index != -1 {
			name = expandShortcut(flagDefs, name[:index]) + name[index:]
		} else {
			name = expandShortcut(flagDefs, name)
		}

		if twoDashes {
			name = "--" + name
		} else {
			name = "-" + name
		}

		return name
	}
	var findCommand = func(name string) *Command {
		var matches []*Command
		for _, c := range cmdDefs {
			if c.HasName(name, true) {
				c.UserName = name
				return c
			}
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

	for _, arg := range args {
		if parsingMode == FlagParsingSkipped {
			nonFlags = append(nonFlags, arg)
			continue
		}

		if arg == "--" {
			parsingMode = FlagParsingSkipped
			if arg != command {
				nonFlags = append(nonFlags, arg)
			}
			continue
		}

		// argument is a flag
		if isFlag(arg) {
			cleanedFlag := cleanFlag(arg)

			previousFlagNeedsValue = false
			// and is present in our flags/shortcuts
			// or special case when we are in command mode because we need to
			// parse them to get validation errors
			if flag := findFlag(flagDefs, cleanedFlag); flag != nil {
				arg = translateShortcutFlags(arg)
				equalPos := strings.Index(arg, "=")
				// an equal sign at the end means we want to use the next arg
				// as the value, but is not supported by flag parsing
				if equalPos == len(arg)-1 {
					// so we strip the sign and reset the position as equals is
					// not here anymore
					arg, equalPos = arg[:equalPos], -1
				}
				// no equals sign ...
				if equalPos == -1 {
					// ... and not a boolean flag nor a verbosity one
					_, isBoolFlag := flag.(*BoolFlag)
					_, isVerbosityFlag := flag.(*verbosityFlag)

					if !isBoolFlag && !isVerbosityFlag {
						// we keep information about the previousFlag.
						previousFlagNeedsValue = true
					}
				}
				// finally, we add to the flags
				flags = append(flags, arg)
			} else if defaultCommand == "--" && parsingMode == FlagParsingNormal {
				trimmedFlag := strings.TrimSpace(arg)
				flags = append(flags, arg)
				previousFlagNeedsValue = trimmedFlag[len(trimmedFlag)-1] == '='
			} else {
				// we add to the non-flags and let the command deal with it later
				nonFlags = append(nonFlags, arg)
			}

			continue
		}

		// If previous flag needs the arg as its value
		if previousFlagNeedsValue {
			// we just add it to the flags list as all processing as been
			// done earlier
			flags = append(flags, arg)
			previousFlagNeedsValue = false
			continue
		}

		// let's find the command if none is set yet
		if command == "" {
			if cmd := findCommand(arg); cmd != nil {
				command = arg
				previousFlagNeedsValue = false
				parsingMode = cmd.FlagParsing
				continue
			}
		}

		// not a flag and command found or default was this, let's stop
		// because the command wants everything else as args
		if parsingMode == FlagParsingSkippedAfterFirstArg {
			parsingMode = FlagParsingSkipped
		}

		nonFlags = append(nonFlags, arg)
	}

	if command != "" {
		flags = append(flags, command)
	}

	return append(flags, nonFlags...)
}

func findFlag(flagDefs []Flag, name string) Flag {
	for _, f := range flagDefs {
		for _, n := range f.Names() {
			if n == name {
				return f
			}
		}
	}
	return nil
}

func expandShortcut(flagDefs []Flag, name string) string {
	if f := findFlag(flagDefs, name); f != nil {
		if _, isVerbosity := f.(*verbosityFlag); isVerbosity {
			return name
		}

		return flagName(f)
	}
	return name
}

func expandHomeInFlagsValues(f *flag.Flag) {
	// This is the safest right now
	if reflect.ValueOf(f.Value).Elem().Kind() != reflect.String {
		return
	}
	val := ExpandHome(f.Value.String())
	if e := f.Value.Set(val); e != nil {
		panic(errors.Errorf("Failed to set flag %s with value %s", f.Name, val))
	}
}

func ExpandHome(path string) string {
	if expandedPath, err := homedir.Expand(path); err == nil {
		return expandedPath
	}

	return path
}

func checkFlagsUnicity(appFlags []Flag, cmdFlags []Flag, commandName string) {
	appDefinedFlags := make(map[string]bool)
	for _, f := range appFlags {
		for _, name := range f.Names() {
			appDefinedFlags[name] = true
		}
	}

	for _, f := range cmdFlags {
		canonicalName := flagName(f)
		for _, name := range f.Names() {
			if appDefinedFlags[name] {
				msg := ""
				if name == canonicalName {
					msg = fmt.Sprintf("flag redefined by command %s: %s", commandName, name)
				} else {
					msg = fmt.Sprintf("flag redefined by command %s: %s (alias for %s)", commandName, name, canonicalName)
				}
				panic(msg) // Happens only if flags are declared with identical names
			}
		}
	}
}

func checkRequiredFlags(flags []Flag, set *flag.FlagSet) error {
	visited := make(map[string]bool)
	set.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	for _, f := range flags {
		if flagIsRequired(f) {
			if !visited[flagName(f)] {
				return errors.Errorf(`Required flag "%s" is not set`, flagName(f))
			}
		}
	}
	return nil
}

func checkFlagsValidity(flags []Flag, set *flag.FlagSet, c *Context) error {
	visited := make(map[string]bool)
	set.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	for _, f := range flags {
		if !visited[flagName(f)] {
			continue
		}
		if err := f.Validate(c); err != nil {
			return errors.Wrapf(err, `invalid value for flag "%s"`, flagName(f))
		}
	}
	return nil
}
