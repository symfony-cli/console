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
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/symfony-cli/terminal"
	. "gopkg.in/check.v1"
)

type ApplicationSuite struct{}

var _ = Suite(&ApplicationSuite{})

func TestApplication(t *testing.T) { TestingT(t) }

var (
	stdin  = bytes.NewBuffer(nil)
	stdout = bytes.NewBuffer(nil)
	stderr = bytes.NewBuffer(nil)
)

func init() {
	terminal.Stdin.SetReader(stdin)
	terminal.Stdout = terminal.NewBufferedConsoleOutput(stdout, stderr)
	terminal.Stderr = terminal.Stdout.Stderr
}

func resetOutputsInput() {
	stdin.Reset()
	stdout.Reset()
	stderr.Reset()
}

var (
	lastExitCode = 0
	fakeOsExiter = func(rc int) {
		lastExitCode = rc
	}
)

func init() {
	OsExiter = fakeOsExiter
	terminal.Stdout = terminal.NewBufferedConsoleOutput(io.Discard, io.Discard)
	terminal.Stderr = terminal.Stdout.Stderr
}

type opCounts struct {
	Total, Before, Action, After int
}

func ExampleApplication_Run() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	app := &Application{
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
		Name:      "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Printf("Hello %v\n", c.String("name"))
			return nil
		},
	}

	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}

func ExampleApplication_Run_quiet() {
	// set args for examples sake
	os.Args = []string{"greet", "-q", "--name", "Jeremy"}

	app := &Application{
		Writer:    os.Stdout,
		ErrWriter: os.Stdout,
		Name:      "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Fprintf(c.App.Writer, "Hello %v\n", c.String("name"))
			fmt.Fprintf(c.App.ErrWriter, "Byebye %v\n", c.String("name"))
			return nil
		},
	}

	app.Run(os.Args)
	// Output:
}

func ExampleApplication_Run_quietDisabled() {
	terminal.DefaultStdout = terminal.NewBufferedConsoleOutput(os.Stdout, os.Stdout)
	// set args for examples sake
	os.Args = []string{"greet", "--quiet=false", "--name", "Jeremy"}

	app := &Application{
		Writer:    os.Stdout,
		ErrWriter: os.Stdout,
		Name:      "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Fprintf(c.App.Writer, "Hello %v\n", c.String("name"))
			fmt.Fprintf(c.App.ErrWriter, "Byebye %v\n", c.String("name"))
			return nil
		},
	}

	app.Run(os.Args)
	// Output:
	// Hello Jeremy
	// Byebye Jeremy
}

func (ts *ApplicationSuite) ExampleApplication_Run_quietInvalid(c *C) {
	resetOutputsInput()

	// set args for examples sake
	os.Args = []string{"greet", "--quiet=foo", "--name", "Jeremy"}

	app := &Application{
		Writer:    os.Stdout,
		ErrWriter: os.Stdout,
		Name:      "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Action: func(c *Context) error {
			fmt.Fprintf(c.App.Writer, "Hello %v\n", c.String("name"))
			fmt.Fprintf(c.App.ErrWriter, "Byebye %v\n", c.String("name"))
			return nil
		},
	}

	app.Run(os.Args)
	c.Assert(stdout.String(), Equals, `Output:
<info>greet</> version <comment>0.0.0</>
A new cli application

<comment>Usage</>:
  app [first_arg] [second_arg]

<comment>Global options:</>
  <info>--help, -h</>                         Show help <comment>[default: false]</>
  <info>--quiet, -q</>                        Do not output any message
  <info>-v|vv|vvv, --verbose, --log-level</>  Increase the verbosity of messages: 1 for normal output, 2 and 3 for more verbose outputs and 4 for debug <comment>[default: 1]</>
  <info>-V</>                                 Print the version <comment>[default: false]</>
  <info>--name=value</>                       a name to say <comment>[default: "bob"]</>

<comment>Available commands:</>
 <comment>self</>           
  <info>self:help, help, list</>  Display help for a command or a category of commands
  <info>self:version, version</>  Display the application version
`)
}

func (ts *ApplicationSuite) ExampleApplication_Run_appHelp(c *C) {
	resetOutputsInput()

	// set args for examples sake
	os.Args = []string{"greet", "help"}

	app := &Application{
		Writer:      os.Stdout,
		ErrWriter:   os.Stderr,
		Name:        "greet",
		Version:     "0.1.0",
		Description: "This is how we describe greet the app",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Commands: []*Command{
			{
				Name:    "describeit",
				Aliases: []*Alias{{Name: "d"}},
				Args: ArgDefinition{
					&Arg{Name: "id"},
				},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(c *Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}
	app.Run(os.Args)
	c.Assert(stdout.String(), Equals, `Output:
<info>greet</> version <comment>0.1.0</>
A new cli application
<comment>Usage</>:
  greet [global options] <command> [command options] [arguments...]
This is how we describe greet the app
<comment>Global options:</>
  <info>--help, -h</>                         Show help <comment>[default: false]</>
  <info>--quiet, -q</>                        Do not output any message
  <info>-v|vv|vvv, --verbose, --log-level</>  Increase the verbosity of messages: 1 for normal output, 2 and 3 for more verbose outputs and 4 for debug <comment>[default: 1]</>
  <info>-V</>                                 Print the version <comment>[default: false]</>
  <info>--name=value</>                       a name to say <comment>[default: "bob"]</>
<comment>Available commands:</>
  <info>describeit, d</>    use it to see a description
 <comment>self</>           
  <info>self:help, help, list</>  Display help for a command or a category of commands
`)
}

func ExampleApplication_Run_commandHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "help", "describeit"}

	app := &Application{
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
		Name:      "greet",
		HelpName:  "greet",
		Flags: []Flag{
			&StringFlag{Name: "name", DefaultValue: "bob", Usage: "a name to say"},
		},
		Commands: []*Command{
			{
				Name:        "describeit",
				Aliases:     []*Alias{{Name: "d"}},
				Usage:       "use it to see a description",
				Description: "This is how we describe describeit the function",
				Action: func(c *Context) error {
					fmt.Printf("i like to describe things")
					return nil
				},
			},
		},
	}
	app.Run(os.Args)
	// Output:
	// <comment>Description:</>
	//   use it to see a description
	//
	// <comment>Usage:</>
	//   greet describeit
	//
	//
	// <comment>Help:</>
	//
	//   This is how we describe describeit the function
}

func (ts *ApplicationSuite) TestApp_Run(c *C) {
	s := ""

	app := &Application{
		Action: func(c *Context) error {
			s += c.Args().first()
			return nil
		},
	}

	err := app.Run([]string{"command", "foo"})
	c.Assert(err, Equals, nil)
	err = app.Run([]string{"command", "bar"})
	c.Assert(err, Equals, nil)
	c.Assert(s, Equals, "foobar")
}

var commandAppTests = []struct {
	name     string
	expected bool
}{
	{"foobar", true},
	{"batbaz", true},
	{"b", true},
	{"f", true},
	{"bata", false},
	{"nothing", false},
}

func (ts *ApplicationSuite) TestApp_Command(c *C) {
	app := &Application{}
	fooCommand := &Command{Name: "foobar", Aliases: []*Alias{{Name: "f"}}}
	batCommand := &Command{Name: "batbaz", Aliases: []*Alias{{Name: "b"}}}
	app.Commands = []*Command{
		fooCommand,
		batCommand,
	}

	for _, test := range commandAppTests {
		c.Assert(app.Command(test.name) != nil, Equals, test.expected)
	}
}

func (ts *ApplicationSuite) TestApp_CommandWithFlagBeforeTerminator(c *C) {
	var parsedOption string
	var args Args

	app := &Application{}
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&StringFlag{Name: "option", DefaultValue: "", Usage: "some option"},
		},
		Args: ArgDefinition{
			&Arg{Name: "first"},
			&Arg{Name: "second"},
		},
		Action: func(c *Context) error {
			parsedOption = c.String("option")
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "--option", "my-option", "my-arg", "--", "--notARealFlag"})

	c.Assert(parsedOption, Equals, "my-option")
	c.Assert(args, NotNil)
	c.Assert(args.Get("first"), Equals, "my-arg")
	c.Assert(args.Get("second"), Equals, "--notARealFlag")
}

func (ts *ApplicationSuite) TestApp_CommandWithDash(c *C) {
	var args Args

	app := &Application{}
	command := &Command{
		Name: "cmd",
		Args: ArgDefinition{
			&Arg{Name: "first"},
			&Arg{Name: "second"},
		},
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "my-arg", "-"})
	c.Assert(args, NotNil)
	c.Assert(args.Len(), Equals, 2)
	c.Assert(args.Get("first"), Equals, "my-arg")
	c.Assert(args.Get("second"), Equals, "-")
}

func (ts *ApplicationSuite) TestApp_CommandWithNoFlagBeforeTerminator(c *C) {
	var args Args

	app := &Application{}
	command := &Command{
		Name: "cmd",
		Args: ArgDefinition{
			&Arg{Name: "first"},
			&Arg{Name: "second"},
		},
		Action: func(c *Context) error {
			args = c.Args()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

	c.Assert(args.Get("first"), Equals, "my-arg")
	c.Assert(args.Get("second"), Equals, "notAFlagAtAll")
}

func (ts *ApplicationSuite) TestApp_VisibleCommands(c *C) {
	frob := &Command{
		Name:     "frob",
		HelpName: "foo frob",
		Action:   func(_ *Context) error { return nil },
	}
	app := &Application{
		Commands: []*Command{
			frob,
			{
				Name:     "frib",
				HelpName: "foo frib",
				Hidden:   Hide,
				Action:   func(_ *Context) error { return nil },
			},
		},
	}

	helpCommand.Hidden = Hide
	versionCommand.Hidden = Hide
	defer func() {
		helpCommand.Hidden = nil
		versionCommand.Hidden = nil
	}()

	app.setup()
	expected := []*Command{
		frob,
	}
	actual := app.VisibleCommands()
	c.Assert(len(actual), Equals, len(expected))
	for i, actualCommand := range actual {
		expectedCommand := expected[i]

		if expectedCommand.Action != nil {
			// comparing func addresses is OK!
			c.Assert(fmt.Sprintf("%p", actualCommand.Action), Equals, fmt.Sprintf("%p", expectedCommand.Action))
		}

		func() {
			// nil out funcs, as they cannot be compared
			// (https://github.com/golang/go/issues/8554)
			expectedAction := expectedCommand.Action
			actualAction := actualCommand.Action
			defer func() {
				expectedCommand.Action = expectedAction
				actualCommand.Action = actualAction
			}()
			expectedCommand.Action = nil
			actualCommand.Action = nil

			c.Assert(expectedCommand, DeepEquals, actualCommand)
		}()
	}
}

func (ts *ApplicationSuite) TestApp_Float64Flag(c *C) {
	var meters float64

	app := &Application{
		Flags: []Flag{
			&Float64Flag{Name: "height", DefaultValue: 1.5, Usage: "Set the height, in meters"},
		},
		Action: func(c *Context) error {
			meters = c.Float64("height")
			return nil
		},
	}

	app.Run([]string{"", "--height", "1.93"})
	c.Assert(meters, Equals, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var firstArg string
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := &Application{}
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&IntSliceFlag{Name: "p", Destination: NewIntSlice(), Usage: "set one or more ip addr"},
			&StringSliceFlag{Name: "ip", Destination: NewStringSlice(), Usage: "set one or more ports to open"},
		},
		Args: []*Arg{
			{Name: "my-arg"},
		},
		Action: func(c *Context) error {
			parsedIntSlice = c.IntSlice("p")
			parsedStringSlice = c.StringSlice("ip")
			firstArg = c.Args().first()
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4", "my-first-arg"})

	IntsEquals := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	StrsEquals := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}
	var expectedIntSlice = []int{22, 80}
	var expectedStringSlice = []string{"8.8.8.8", "8.8.4.4"}

	if !IntsEquals(parsedIntSlice, expectedIntSlice) {
		t.Errorf("%v does not match %v", parsedIntSlice, expectedIntSlice)
	}

	if !StrsEquals(parsedStringSlice, expectedStringSlice) {
		t.Errorf("%v does not match %v", parsedStringSlice, expectedStringSlice)
	}

	if firstArg != "my-first-arg" {
		t.Errorf("%v does not match %v", firstArg, "my-first-arg")
	}
}

func TestApp_ParseSliceFlagsWithMissingValue(t *testing.T) {
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := &Application{}
	command := &Command{
		Name: "cmd",
		Flags: []Flag{
			&IntSliceFlag{Name: "a", Usage: "set numbers"},
			&StringSliceFlag{Name: "str", Usage: "set strings"},
		},
		Args: []*Arg{
			{Name: "my-arg"},
		},
		Action: func(c *Context) error {
			parsedIntSlice = c.IntSlice("a")
			parsedStringSlice = c.StringSlice("str")
			return nil
		},
	}
	app.Commands = []*Command{command}

	app.Run([]string{"", "cmd", "-a", "2", "-str", "A", "my-arg"})

	var expectedIntSlice = []int{2}
	var expectedStringSlice = []string{"A"}

	if len(parsedIntSlice) != len(expectedIntSlice) {
		t.Fatalf("%v does not match %v", len(parsedIntSlice), len(expectedIntSlice))
	}
	if parsedIntSlice[0] != expectedIntSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}

	if len(parsedStringSlice) != len(expectedStringSlice) {
		t.Fatalf("%v does not match %v", len(parsedStringSlice), len(expectedStringSlice))
	}
	if parsedStringSlice[0] != expectedStringSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}
}

func TestApp_DefaultStdout(t *testing.T) {
	app := &Application{}
	app.setup()

	if app.Writer == nil {
		t.Error("Default output writer not set.")
	}
}

type mockWriter struct {
	written []byte
}

func (fw *mockWriter) Write(p []byte) (n int, err error) {
	if fw.written == nil {
		fw.written = p
	} else {
		fw.written = append(fw.written, p...)
	}

	return len(p), nil
}

func (fw *mockWriter) GetWritten() (b []byte) {
	return fw.written
}

func TestApp_SetStdout(t *testing.T) {
	w := &mockWriter{}

	app := &Application{
		Name:   "test",
		Writer: w,
	}

	err := app.Run([]string{"help"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if len(w.written) == 0 {
		t.Error("App did not write output to desired writer.")
	}
}

func TestApp_BeforeFunc(t *testing.T) {
	counts := &opCounts{}
	beforeError := errors.New("fail")
	var err error

	app := &Application{
		Before: func(c *Context) error {
			counts.Total++
			counts.Before = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return beforeError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(c *Context) error {
					counts.Total++
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
	}

	// run with the Before() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != beforeError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}

	// reset
	counts = &opCounts{}

	afterError := errors.New("fail again")
	app.After = func(_ *Context) error {
		return afterError
	}

	// run with the Before() func failing, wrapped by After()
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if _, ok := err.(MultiError); !ok {
		t.Errorf("MultiError expected, but not received")
	}

	if counts.Before != 1 {
		t.Errorf("Before() not executed when expected")
	}
}

func TestApp_AfterFunc(t *testing.T) {
	counts := &opCounts{}
	afterError := errors.New("fail")
	var err error

	app := &Application{
		After: func(c *Context) error {
			counts.Total++
			counts.After = counts.Total
			s := c.String("opt")
			if s == "fail" {
				return afterError
			}

			return nil
		},
		Commands: []*Command{
			{
				Name: "sub",
				Action: func(c *Context) error {
					counts.Total++
					return nil
				},
			},
		},
		Flags: []Flag{
			&StringFlag{Name: "opt"},
		},
	}

	// run with the After() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}

	// reset
	counts = &opCounts{}

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != afterError {
		t.Errorf("Run error expected, but not received")
	}

	if counts.After != 2 {
		t.Errorf("After() not executed when expected")
	}
}

func TestAppNoHelpFlag(t *testing.T) {
	oldFlag := HelpFlag
	defer func() {
		HelpFlag = oldFlag
	}()

	HelpFlag = nil

	app := &Application{Writer: io.Discard}
	err := app.Run([]string{"test", "-h"})

	if !strings.Contains(err.Error(), flag.ErrHelp.Error()) {
		t.Errorf("expected error about missing help flag, but got: %q (%T)", err, err)
	}
}

func TestAppHelpPrinter(t *testing.T) {
	oldPrinter := HelpPrinter
	defer func() {
		HelpPrinter = oldPrinter
	}()

	var wasCalled = false
	HelpPrinter = func(w io.Writer, template string, data interface{}) {
		wasCalled = true
	}

	app := &Application{}
	app.Run([]string{"-h"})

	if wasCalled == false {
		t.Errorf("Help printer expected to be called, but was not")
	}
}

func TestApp_VersionPrinter(t *testing.T) {
	oldPrinter := VersionPrinter
	defer func() {
		VersionPrinter = oldPrinter
	}()

	var wasCalled = false
	VersionPrinter = func(c *Context) {
		wasCalled = true
	}

	app := &Application{}
	ctx := NewContext(app, nil, nil)
	ShowVersion(ctx)

	if wasCalled == false {
		t.Errorf("Version printer expected to be called, but was not")
	}
}

func (ts *ApplicationSuite) TestApp_OrderOfOperations(c *C) {
	counts := &opCounts{}
	resetCounts := func() { counts = &opCounts{} }

	app := &Application{}

	beforeNoError := func(c *Context) error {
		counts.Total++
		counts.Before = counts.Total
		return nil
	}

	beforeError := func(c *Context) error {
		counts.Total++
		counts.Before = counts.Total
		return errors.New("hay Before")
	}

	app.Before = beforeNoError

	afterNoError := func(c *Context) error {
		counts.Total++
		counts.After = counts.Total
		return nil
	}

	afterError := func(c *Context) error {
		counts.Total++
		counts.After = counts.Total
		return errors.New("hay After")
	}

	app.After = afterNoError
	app.Commands = []*Command{
		{
			Name: "bar",
			Action: func(c *Context) error {
				counts.Total++
				return nil
			},
		},
	}

	app.Action = func(c *Context) error {
		counts.Total++
		counts.Action = counts.Total
		return nil
	}

	_ = app.Run([]string{"command", "--nope"})
	c.Assert(counts.Total, Equals, 0)

	resetCounts()

	_ = app.Run([]string{"command", "--nope"})
	c.Assert(counts.Total, Equals, 0)

	resetCounts()

	_ = app.Run([]string{"command", "foo"})
	c.Assert(counts.Before, Equals, 1)
	c.Assert(counts.Action, Equals, 2)
	c.Assert(counts.After, Equals, 3)
	c.Assert(counts.Total, Equals, 3)

	resetCounts()

	app.Before = beforeError
	_ = app.Run([]string{"command", "bar"})
	c.Assert(counts.Before, Equals, 1)
	c.Assert(counts.After, Equals, 2)
	c.Assert(counts.Total, Equals, 2)
	app.Before = beforeNoError

	resetCounts()

	app.After = nil
	_ = app.Run([]string{"command", "bar"})
	c.Assert(counts.Before, Equals, 1)
	c.Assert(counts.Total, Equals, 2)
	app.After = afterNoError

	resetCounts()

	app.After = afterError
	err := app.Run([]string{"command", "bar"})
	c.Assert(err, NotNil)
	c.Assert(counts.Before, Equals, 1)
	c.Assert(counts.After, Equals, 3)
	c.Assert(counts.Total, Equals, 3)
	app.After = afterNoError

	resetCounts()

	oldCommands := app.Commands
	app.Commands = nil
	_ = app.Run([]string{"command"})
	c.Assert(counts.Before, Equals, 1)
	c.Assert(counts.Action, Equals, 2)
	c.Assert(counts.After, Equals, 3)
	c.Assert(counts.Total, Equals, 3)
	app.Commands = oldCommands
}

func TestApp_Run_CommandHelpName(t *testing.T) {
	t.SkipNow()
	app := &Application{}
	buf := new(bytes.Buffer)
	app.Writer = buf
	app.Name = "command"
	cmd := &Command{
		Name:        "foo",
		HelpName:    "custom",
		Description: "foo commands",
	}
	app.Commands = []*Command{cmd}

	err := app.Run([]string{"command", "foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()

	expected := "command foo bar - does bar things"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}

	expected = "command foo bar [command options] [arguments...]"
	if !strings.Contains(output, expected) {
		t.Errorf("expected %q in output: %s", expected, output)
	}
}

func TestApp_Run_Help(t *testing.T) {
	var helpArguments = [][]string{{"boom", "--help"}, {"boom", "-h"}, {"boom", "help"}}

	for _, args := range helpArguments {
		buf := new(bytes.Buffer)

		app := &Application{
			Name:   "boom",
			Usage:  "make an explosive entrance",
			Writer: buf,
			Action: func(c *Context) error {
				buf.WriteString("boom I say!")
				return nil
			},
		}

		err := app.Run(args)
		if err != nil {
			t.Error(err)
		}

		output := buf.String()
		if !strings.Contains(output, "<info>boom</> version") || !strings.Contains(output, "\nmake an explosive entrance") {
			t.Logf("==> checking with arguments %v", args)
			t.Logf("output: %q\n", buf.Bytes())
			t.Errorf("want help to contain %q, did not: \n%q", "boom version [...] make an explosive entrance", output)
		}
	}
}

func TestApp_Run_Version(t *testing.T) {
	var versionArguments = [][]string{{"boom", "-V"}}

	for _, args := range versionArguments {
		buf := new(bytes.Buffer)

		app := &Application{
			Name:    "boom",
			Usage:   "make an explosive entrance",
			Version: "0.1.0",
			Writer:  buf,
			Action: func(c *Context) error {
				buf.WriteString("boom I say!")
				return nil
			},
		}

		err := app.Run(args)
		if err != nil {
			t.Error(err)
		}

		output := buf.String()

		if !strings.Contains(output, "0.1.0") {
			t.Logf("==> checking with arguments %v", args)
			t.Logf("output: %q\n", buf.Bytes())
			t.Errorf("want version to contain %q, did not: \n%q", "0.1.0", output)
		}
	}
}

func TestApp_Run_Categories(t *testing.T) {
	buf := new(bytes.Buffer)

	app := &Application{
		Name: "categories",
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
			},
			{
				Name:     "command2",
				Category: "1",
			},
			{
				Name:     "command3",
				Category: "2",
			},
		},
		Writer: buf,
	}
	helpCommand.Hidden = Hide
	versionCommand.Hidden = Hide
	defer func() {
		helpCommand.Hidden = nil
		versionCommand.Hidden = nil
	}()

	app.Run([]string{"categories"})

	expect := commandCategories([]*commandCategory{
		{
			name: "1",
			commands: []*Command{
				app.Commands[0],
				app.Commands[1],
			},
		},
		{
			name: "2",
			commands: []*Command{
				app.Commands[2],
			},
		},
	})

	if !reflect.DeepEqual(app.Categories, &expect) {
		t.Fatalf("expected categories %#v, to equal %#v", app.Categories, &expect)
	}

	output := buf.String()

	if !strings.Contains(output, "<comment>1</>         \n  <info>1:command1</>") {
		t.Logf("output: %q\n", buf.Bytes())
		t.Errorf("want buffer to include category %q, did not: \n%q", "<comment>1</>\n  <info>1:command1</>", output)
	}
}

func (ts *ApplicationSuite) TestApp_VisibleCategories(c *C) {
	app := &Application{
		Name: "visible-categories",
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   Hide,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
			},
		},
	}

	helpCommand.Hidden = Hide
	versionCommand.Hidden = Hide
	defer func() {
		helpCommand.Hidden = nil
		versionCommand.Hidden = nil
	}()

	expected := []CommandCategory{
		&commandCategory{
			name: "2",
			commands: []*Command{
				app.Commands[1],
			},
		},
		&commandCategory{
			name: "3",
			commands: []*Command{
				app.Commands[2],
			},
		},
	}

	app.setup()
	c.Assert(app.VisibleCategories(), DeepEquals, expected)

	app = &Application{
		Name: "visible-categories",
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   Hide,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
				Hidden:   Hide,
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
			},
		},
	}

	expected = []CommandCategory{
		&commandCategory{
			name: "3",
			commands: []*Command{
				app.Commands[2],
			},
		},
	}

	app.setup()
	c.Assert(app.VisibleCategories(), DeepEquals, expected)

	app = &Application{
		Name: "visible-categories",
		Commands: []*Command{
			{
				Name:     "command1",
				Category: "1",
				HelpName: "foo command1",
				Hidden:   Hide,
			},
			{
				Name:     "command2",
				Category: "2",
				HelpName: "foo command2",
				Hidden:   Hide,
			},
			{
				Name:     "command3",
				Category: "3",
				HelpName: "foo command3",
				Hidden:   Hide,
			},
		},
	}

	app.setup()
	c.Assert(app.VisibleCategories(), DeepEquals, []CommandCategory{})
}

func TestApp_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := &Application{
		Action: func(c *Context) error { return nil },
		Before: func(c *Context) error { return errors.New("before error") },
		After:  func(c *Context) error { return errors.New("after error") },
	}

	err := app.Run([]string{"foo"})
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
