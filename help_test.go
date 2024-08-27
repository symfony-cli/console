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
	"bytes"
	"flag"
	"io"
	"strings"
	"testing"
)

func Test_ShowAppHelp_NoAuthor(t *testing.T) {
	output := new(bytes.Buffer)
	app := &Application{Writer: output}

	c := NewContext(app, nil, nil)

	ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("AUTHOR(S):")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "AUTHOR(S):")
	}
}

func Test_ShowAppHelp_NoVersion(t *testing.T) {
	output := new(bytes.Buffer)
	app := &Application{Writer: output}

	app.Version = ""

	c := NewContext(app, nil, nil)

	ShowAppHelp(c)

	if bytes.Contains(output.Bytes(), []byte("VERSION:")) {
		t.Errorf("expected\n%snot to include %s", output.String(), "VERSION:")
	}
}

func Test_Help_Custom_Flags(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = &BoolFlag{
		Name:    "xxx",
		Aliases: []string{"x"},
		Usage:   "show help",
	}

	app := Application{
		Flags: []Flag{
			&BoolFlag{Name: "help", Aliases: []string{"h"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("help") != true {
				t.Errorf("custom help flag not set")
			}
			return nil
		},
	}
	output := new(bytes.Buffer)
	app.Writer = output
	app.Run([]string{"test", "-h"})
	if output.Len() > 0 {
		t.Errorf("unexpected output: %s", output.String())
	}
}

func Test_Version_Custom_Flags(t *testing.T) {
	oldFlag := VersionFlag
	defer func() {
		VersionFlag = oldFlag
	}()

	VersionFlag = &BoolFlag{
		Name:    "version",
		Aliases: []string{"a"},
		Usage:   "show version",
	}

	app := Application{
		Flags: []Flag{
			&BoolFlag{Name: "foo", Aliases: []string{"V"}},
		},
		Action: func(ctx *Context) error {
			if ctx.Bool("V") != true {
				t.Errorf("custom version flag not set")
			}
			return nil
		},
	}
	output := new(bytes.Buffer)
	app.Writer = output
	app.Run([]string{"test", "-V"})
	if output.Len() > 0 {
		t.Errorf("unexpected output: %s", output.String())
	}
}

func Test_helpCommand_Action_ErrorIfNoTopic(t *testing.T) {
	app := &Application{}
	app.Writer, app.ErrWriter = io.Discard, io.Discard

	set := flag.NewFlagSet("test", 0)
	set.Parse([]string{"foo"})

	c := NewContext(app, set, nil)
	app.setup()

	err := helpCommand.Action(c)

	if err == nil {
		t.Fatalf("expected error from helpCommand.Action(), but got nil")
	}

	exitErr, ok := err.(ExitCoder)
	if !ok {
		t.Fatalf("expected *exitError from helpCommand.Action(), but instead got: %v", err.Error())
	}

	if exitErr.Error() != "Command \"foo\" does not exist." {
		t.Fatalf("expected an command not found error, but got: %q", exitErr.Error())
	}

	if exitErr.ExitCode() != 3 {
		t.Fatalf("expected exit value = 3, got %d instead", exitErr.ExitCode())
	}
}

func Test_helpCommand_InHelpOutput(t *testing.T) {
	app := &Application{}
	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"test", "--help"})

	s := output.String()

	if strings.Contains(s, "\nCOMMANDS:\nGLOBAL OPTIONS:\n") {
		t.Fatalf("empty COMMANDS section detected: %q", s)
	}

	if !strings.Contains(s, "--help, -h") {
		t.Fatalf("missing \"help, h\": %q", s)
	}
}

func Test_helpCategories(t *testing.T) {
	app := &Application{}
	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"help"})

	s := output.String()

	if !strings.Contains(s, "Available commands") {
		t.Fatalf("commands are not listed: %q", s)
	}

	output.Reset()
	app.Run([]string{"help", "self"})
	s = output.String()

	if !strings.Contains(s, "Available commands for the \"self\" namespace:") {
		t.Fatalf("commands from a category are not listed: %q", s)
	}
}

func TestShowAppHelp_CommandAliases(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []*Alias{{Name: "fr"}, {Name: "frob"}, {Name: "not-here", Hidden: true}},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"foo", "--help"})

	if !strings.Contains(output.String(), "<info>frobbly, fr, frob</>") {
		t.Errorf("expected output to include all command aliases; got: %q", output.String())
	}

	if strings.Contains(output.String(), "not-here") {
		t.Errorf("expected output to exclude hidden aliases; got: %q", output.String())
	}
}

func TestShowCommandHelp_CommandAliases(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name:    "frobbly",
				Aliases: []*Alias{{Name: "fr"}, {Name: "frob"}, {Name: "bork"}},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"foo", "help", "fr"})

	if !strings.Contains(output.String(), "frobbly") {
		t.Errorf("expected output to include command name; got: %q", output.String())
	}

	if strings.Contains(output.String(), "bork") {
		t.Errorf("expected output to exclude command aliases; got: %q", output.String())
	}
}

func TestShowCommandHelp_CommandShortcut(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name:     "bar",
				Category: "foo",
				Aliases:  []*Alias{{Name: "fb"}},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"foo", "help", "f:b"})

	if !strings.Contains(output.String(), "foo:bar") {
		t.Errorf("expected output to include command name; got: %q", output.String())
	}
}

func TestShowCommandHelp_DescriptionFunc(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name:        "frobbly",
				Description: "this is not my custom description",
				DescriptionFunc: func(*Command, *Application) string {
					return "this is my custom description"
				},
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"foo", "help", "frobbly"})

	if !strings.Contains(output.String(), "this is my custom description") {
		t.Errorf("expected output to include result of DescriptionFunc; got: %q", output.String())
	}
}

func TestShowAppHelp_HiddenCommand(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name: "frobbly",
				Action: func(ctx *Context) error {
					return nil
				},
			},
			{
				Name:   "secretfrob",
				Hidden: Hide,
				Action: func(ctx *Context) error {
					return nil
				},
			},
		},
	}

	output := &bytes.Buffer{}
	app.Writer = output
	app.Run([]string{"app", "--help"})

	if strings.Contains(output.String(), "secretfrob") {
		t.Errorf("expected output to exclude \"secretfrob\"; got: %q", output.String())
	}

	if !strings.Contains(output.String(), "frobbly") {
		t.Errorf("expected output to include \"frobbly\"; got: %q", output.String())
	}
}
