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
	"text/tabwriter"
	"text/template"

	"github.com/agext/levenshtein"
	"github.com/rs/zerolog"
)

// AppHelpTemplate is the text template for the Default help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var AppHelpTemplate = `<info>{{.Name}}</>{{if .Version}} version <comment>{{.Version}}</>{{end}}{{if .Copyright}} {{.Copyright}}{{end}}
{{.Usage}}

<comment>Usage</>:
  {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} <command> [command options]{{end}} [arguments...]{{if .Description}}

{{.Description}}{{end}}{{if .VisibleFlags}}

<comment>Global options:</>
  {{range $index, $option := .VisibleFlags}}{{if $index}}
  {{end}}{{$option}}{{end}}{{end}}{{if .VisibleCommands}}

<comment>Available commands:</>{{range .VisibleCategories}}{{if .Name}}
 <comment>{{.Name}}</>{{"\t"}}{{end}}{{range .VisibleCommands}}
  <info>{{join .Names ", "}}</>{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}
`

// CategoryHelpTemplate is the text template for the category help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CategoryHelpTemplate = `{{with .App }}<info>{{.Name}}</>{{if .Version}} version <comment>{{.Version}}</>{{end}}{{if .Copyright}} {{.Copyright}}{{end}}
{{.Usage}}

<comment>Usage</>:
  {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} <command> [command options]{{end}} [arguments...]{{if .Description}}

{{.Description}}{{end}}{{if .VisibleFlags}}

<comment>Global options:</>
  {{range $index, $option := .VisibleFlags}}{{if $index}}
  {{end}}{{$option}}{{end}}{{end}}{{end}}{{ range .Categories }}

<comment>Available commands for the "{{.Name}}" namespace:</>{{range .VisibleCommands}}
 <info>{{join .Names ", "}}</>{{"\t"}}{{.Usage}}{{end}}{{end}}
`

// CommandHelpTemplate is the text template for the command help topic.
// cli.go uses text/template to render templates. You can
// render custom help text by setting this variable.
var CommandHelpTemplate = `{{if .Usage}}<comment>Description:</>
  {{.Usage}}

{{end}}<comment>Usage:</>
  {{.HelpName}}{{if .VisibleFlags}} [options]{{end}}{{.Arguments.Usage}}{{if .Arguments}}

<comment>Arguments:</>
  {{range .Arguments}}{{.}}
  {{end}}{{end}}{{if .VisibleFlags}}

<comment>Options:</>
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}{{if .Description}}

<comment>Help:</>

  {{.Description}}
{{end}}
`

var helpCommand = &Command{
	Category: "self",
	Name:     "help",
	Aliases:  []*Alias{{Name: "help"}, {Name: "list"}},
	Usage:    "Display help for a command or a category of commands",
	Args: []*Arg{
		{Name: "command", Optional: true},
	},
	Action: ShowAppHelpAction,
}

var versionCommand = &Command{
	Category: "self",
	Name:     "version",
	Aliases:  []*Alias{{Name: "version"}},
	Usage:    "Display the application version",
	Action: func(c *Context) error {
		ShowVersion(c)
		return nil
	},
}

// Prints help for the App or Command
type helpPrinter func(w io.Writer, templ string, data interface{})

// HelpPrinter is a function that writes the help output. If not set a default
// is used. The function signature is:
// func(w io.Writer, templ string, data interface{})
var HelpPrinter helpPrinter = printHelp

// VersionPrinter prints the version for the App
var VersionPrinter = printVersion

// ShowAppHelpAction is an action that displays the global help or for the
// specified command.
func ShowAppHelpAction(c *Context) error {
	args := c.Args()
	if args.Present() {
		// We use `first` here because if we are in a situation of an unknown
		// command, args parsing is not done.
		return ShowCommandHelp(c, args.first())
	}

	ShowAppHelp(c)
	return nil
}

// ShowAppHelp is an action that displays the help.
func ShowAppHelp(c *Context) error {
	HelpPrinter(c.App.Writer, AppHelpTemplate, c.App)
	return nil
}

// ShowCommandHelp prints help for the given command
func ShowCommandHelp(ctx *Context, command string) error {
	if c := ctx.App.BestCommand(command); c != nil {
		if c.DescriptionFunc != nil {
			c.Description = c.DescriptionFunc(c, ctx.App)
		}

		HelpPrinter(ctx.App.Writer, CommandHelpTemplate, c)
		return nil
	}

	categories := []CommandCategory{}
	for _, c := range ctx.App.VisibleCategories() {
		if strings.HasPrefix(c.Name(), command) {
			categories = append(categories, c)
		}
	}
	if len(categories) > 0 {
		HelpPrinter(ctx.App.Writer, CategoryHelpTemplate, struct {
			App        *Application
			Categories []CommandCategory
		}{
			App:        ctx.App,
			Categories: categories,
		})
		return nil
	}

	return &CommandNotFoundError{command, ctx.App}
}

type CommandNotFoundError struct {
	command string
	app     *Application
}

func (e *CommandNotFoundError) Error() string {
	message := fmt.Sprintf("Command %q does not exist.", e.command)
	if alternatives := findAlternatives(e.command, e.app.VisibleCommands()); len(alternatives) == 1 {
		message += "\n\nDid you mean this?\n    " + alternatives[0]
	} else if len(alternatives) > 1 {
		message += "\n\nDid you mean one of these?\n    "
		message += strings.Join(alternatives, "\n    ")
	}

	return message
}

func (e *CommandNotFoundError) ExitCode() int {
	return 3
}

func (e *CommandNotFoundError) GetSeverity() zerolog.Level {
	return zerolog.InfoLevel
}

func findAlternatives(name string, commands []*Command) []string {
	alternatives := []string{}

	for _, command := range commands {
		if command.Category != "" {
			if command.Category == name {
				alternatives = append(alternatives, command.FullName())
				continue
			}

			lev := levenshtein.Distance(name, command.Category, nil)
			if lev <= len(name)/3 {
				alternatives = append(alternatives, command.FullName())
				continue
			}
		}

		for _, cmdName := range command.Names() {
			if strings.HasPrefix(cmdName, name) {
				alternatives = append(alternatives, cmdName)
				continue
			}
			if strings.HasSuffix(cmdName, name) {
				alternatives = append(alternatives, cmdName)
				continue
			}

			lev := levenshtein.Distance(name, cmdName, nil)
			if lev <= len(name)/3 {
				alternatives = append(alternatives, cmdName)
				continue
			}
		}
	}

	sort.Strings(alternatives)

	return alternatives
}

// ShowVersion prints the version number of the App
func ShowVersion(c *Context) {
	VersionPrinter(c)
}

func printVersion(c *Context) {
	HelpPrinter(c.App.Writer, "<info>{{.Name}}</>{{if .Version}} version <comment>{{.Version}}</>{{end}}{{if .Copyright}} {{.Copyright}}{{end}} ({{.BuildDate}} - {{.Channel}})\n", c.App)
}

func printHelp(out io.Writer, templ string, data interface{}) {
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))

	err := t.Execute(w, data)
	if err != nil {
		panic(fmt.Errorf("CLI TEMPLATE ERROR: %#v", err.Error()))
	}
	w.Flush()
}

func checkVersion(c *Context) bool {
	found := false
	if VersionFlag.Name != "" {
		for _, name := range VersionFlag.Names() {
			if c.Bool(name) {
				found = true
			}
		}
	}
	return found
}

func IsHelp(c *Context) bool {
	return checkHelp(c) || c.Command == helpCommand
}

func checkHelp(c *Context) bool {
	for _, name := range HelpFlag.Names() {
		if c.Bool(name) {
			return true
		}
	}

	return false
}

func checkCommandHelp(c *Context, name string) bool {
	if c.Bool("h") || c.Bool("help") {
		ShowCommandHelp(c, name)
		return true
	}

	return false
}
