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
	"errors"
	"flag"
	"io"
	"reflect"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

type CommandSuite struct{}

var _ = Suite(&CommandSuite{})

func (cs *CommandSuite) TestCommandFlagParsing(c *C) {
	cases := []struct {
		testArgs        []string
		skipFlagParsing bool
		expectedErr     string
	}{
		// Test normal "not ignoring flags" flow
		{[]string{"test-cmd", "-break", "blah", "blah"}, false, "Incorrect usage: flag provided but not defined: -break"},

		{[]string{"test-cmd", "blah", "blah"}, true, ""},   // Test SkipFlagParsing without any args that look like flags
		{[]string{"test-cmd", "blah", "-break"}, true, ""}, // Test SkipFlagParsing with random flag arg
		{[]string{"test-cmd", "blah", "-help"}, true, ""},  // Test SkipFlagParsing with "special" help flag arg
	}

	for _, ca := range cases {
		app := &Application{}
		app.setup()
		set := flag.NewFlagSet("test", 0)
		set.Parse(ca.testArgs)

		context := NewContext(app, set, nil)

		flagParsingMode := FlagParsingNormal
		if ca.skipFlagParsing {
			flagParsingMode = FlagParsingSkipped
		}

		command := Command{
			Name:        "test-cmd",
			Aliases:     []*Alias{{Name: "tc"}},
			Usage:       "this is for testing",
			Description: "testing",
			Action:      func(_ *Context) error { return nil },
			FlagParsing: flagParsingMode,
			Args: []*Arg{
				{Name: "my-arg", Slice: true},
			},
		}

		err := command.Run(context)

		if ca.expectedErr == "" {
			c.Assert(err, Equals, nil)
		} else {
			c.Assert(err, ErrorMatches, ca.expectedErr)
		}
		c.Assert(context.Args().Slice(), DeepEquals, ca.testArgs)
	}
}

func TestCommand_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &Application{
		Commands: []*Command{
			{
				Name: "bar",
				Before: func(c *Context) error {
					return errors.New("before error")
				},
				After: func(c *Context) error {
					return errors.New("after error")
				},
			},
		},
	}

	err := app.Run([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected to receive error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}

func TestCaseInsensitiveCommandNames(t *testing.T) {
	app := Application{}
	app.ErrWriter = io.Discard
	projectList := &Command{Name: "project:LIST", Aliases: []*Alias{{Name: "FOO"}}}
	projectLink := &Command{Name: "PROJECT:link"}
	app.Commands = []*Command{
		projectList,
		projectLink,
	}

	app.setup()

	if c := app.BestCommand("project:list"); c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	if c := app.BestCommand("Project:lISt"); c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	if c := app.BestCommand("project:link"); c != projectLink {
		t.Fatalf("expected project:link, got %v", c)
	}
	if c := app.BestCommand("project:Link"); c != projectLink {
		t.Fatalf("expected project:link, got %v", c)
	}
	if c := app.BestCommand("foo"); c != projectList {
		t.Fatalf("expected project:link, got %v", c)
	}
	if c := app.BestCommand("FoO"); c != projectList {
		t.Fatalf("expected project:link, got %v", c)
	}
}

func TestFuzzyCommandNames(t *testing.T) {
	app := Application{}
	app.ErrWriter = io.Discard
	projectList := &Command{Name: "project:list"}
	projectLink := &Command{Name: "project:link"}
	app.Commands = []*Command{
		projectList,
		projectLink,
	}

	c := app.BestCommand("project:list")
	if c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	c = app.BestCommand("project:link")
	if c != projectLink {
		t.Fatalf("expected project:link, got %v", c)
	}
	c = app.BestCommand("pro:list")
	if c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	c = app.BestCommand("pro:lis")
	if c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	c = app.BestCommand("p:lis")
	if c != projectList {
		t.Fatalf("expected project:list, got %v", c)
	}
	c = app.BestCommand("p:li")
	if c != nil {
		t.Fatalf("expected no matches, got %v", c)
	}
}

func TestCommandWithNoNames(t *testing.T) {
	c := Command{
		Aliases: []*Alias{
			{},
			{Name: "foo"},
			{Name: "bar"},
		},
	}

	if got, expected := c.Names(), []string{"foo", "bar"}; len(got) != 2 {
		t.Fatalf("expected two names, got %v", len(got))
	} else if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}

	if name := c.PreferredName(); name != "foo, bar" {
		t.Fatalf(`expected "foo, bar", got "%v"`, name)
	}
}
