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
	"fmt"

	"github.com/symfony-cli/terminal"
	. "gopkg.in/check.v1"
)

type CliEnhancementSuite struct{}

var _ = Suite(&CliEnhancementSuite{})

var (
	testAppUploadFlags = []Flag{
		&IntFlag{Name: "reference", Aliases: []string{"r"}},
		&IntFlag{Name: "samples", Aliases: []string{"s"}},
		&BoolFlag{Name: "test", Aliases: []string{"t"}},
	}

	curlCmd   = &Command{Name: "curl", Flags: testAppUploadFlags}
	uploadCmd = &Command{Name: "upload", Flags: testAppUploadFlags}
	fooCmd    = &Command{Name: "foo", Flags: testAppUploadFlags, FlagParsing: FlagParsingSkipped}
	runCmd    = &Command{Name: "run", Flags: testAppUploadFlags, FlagParsing: FlagParsingSkippedAfterFirstArg}

	testApp = Application{
		Flags: []Flag{
			&IntFlag{Name: "v", DefaultValue: 1},
			&StringFlag{Name: "server-id"},
			&StringFlag{Name: "server-token"},
			&StringFlag{Name: "config"},
			&BoolFlag{Name: "quiet", Aliases: []string{"q"}},
		},
		Commands: []*Command{
			{Name: "agent"},
			curlCmd,
			uploadCmd,
			fooCmd,
			runCmd,
		},
	}
)

func (ts *CliEnhancementSuite) TestFixAndParseArgsApplication(c *C) {
	var (
		args     []string
		expected []string
		sorted   []string

		ctx          *Context
		fs           *flag.FlagSet
		argsExpected []string
	)

	args = []string{"-reference=4", "--v=3", "-q", "upload", "file1", "file2"}
	expected = []string{"--v=3", "-quiet", "upload", "-reference=4", "file1", "file2"}
	argsExpected = []string{"upload", "-reference=4", "file1", "file2"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, err := testApp.parseArgs(args)
	c.Assert(err, IsNil)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 3)
	c.Check(ctx.Bool("quiet"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"-reference", "4", "--v=3", "-q", "upload", "file1", "file2"}
	expected = []string{"--v=3", "-quiet", "upload", "-reference", "4", "file1", "file2"}
	argsExpected = []string{"upload", "-reference", "4", "file1", "file2"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 3)
	c.Check(ctx.Bool("quiet"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"upload", "-reference=4", "-v=3", "-q", "file1", "file2"}
	expected = []string{"-v=3", "-quiet", "upload", "-reference=4", "file1", "file2"}
	argsExpected = []string{"upload", "-reference=4", "file1", "file2"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 3)
	c.Check(ctx.Bool("quiet"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"upload", "-reference=4", "-v=3", "-q", "upload", "file1", "file2"}
	expected = []string{"-v=3", "-quiet", "upload", "-reference=4", "upload", "file1", "file2"}
	argsExpected = []string{"upload", "-reference=4", "upload", "file1", "file2"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 3)
	c.Check(ctx.Bool("quiet"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"curl", "-reference=4", "-v=3", "-q", "-X", "POST", "http://blackfire.io"}
	expected = []string{"-v=3", "-quiet", "curl", "-reference=4", "-X", "POST", "http://blackfire.io"}
	argsExpected = []string{"curl", "-reference=4", "-X", "POST", "http://blackfire.io"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 3)
	c.Check(ctx.Bool("quiet"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"curl"}
	expected = []string{"curl"}
	argsExpected = []string{"curl"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 1)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"-server-id=75299154-8b63-4632-9b04-1e10bb19c144", "-server-token=f30f10d62f6f577e90e1be4e218a638ec3d16a0e0454bd69b2459bb046588c6f", "agent"}
	expected = []string{"-server-id=75299154-8b63-4632-9b04-1e10bb19c144", "-server-token=f30f10d62f6f577e90e1be4e218a638ec3d16a0e0454bd69b2459bb046588c6f", "agent"}
	argsExpected = []string{"agent"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 1)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"-server-id=75299154-8b63-4632-9b04-1e10bb19c144", "-server-token=f30f10d62f6f577e90e1be4e218a638ec3d16a0e0454bd69b2459bb046588c6f", "agent", "-v=4"}
	expected = []string{"-server-id=75299154-8b63-4632-9b04-1e10bb19c144", "-server-token=f30f10d62f6f577e90e1be4e218a638ec3d16a0e0454bd69b2459bb046588c6f", "-v=4", "agent"}
	argsExpected = []string{"agent"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 4)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"run", "-v=4", "--reference", "8", "php", "vd.php"}
	expected = []string{"-v=4", "run", "--reference", "8", "php", "vd.php"}
	argsExpected = []string{"run", "--reference", "8", "php", "vd.php"}
	sorted = testApp.fixArgs(args)
	c.Check(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 4)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"run", "--", "-v=4", "--reference", "8", "php", "vd.php"}
	expected = []string{"run", "--", "-v=4", "--reference", "8", "php", "vd.php"}
	argsExpected = []string{"run", "-v=4", "--reference", "8", "php", "vd.php"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 1)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"-v=4", "foo", "--reference", "8", "php", "vd.php"}
	expected = []string{"-v=4", "foo", "--reference", "8", "php", "vd.php"}
	argsExpected = []string{"foo", "--reference", "8", "php", "vd.php"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 4)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"-config", "/Users/marc/.blackfire-d1.ini", "-reference=19", "upload", "profiler/README.md"}
	expected = []string{"-config", "/Users/marc/.blackfire-d1.ini", "upload", "-reference=19", "profiler/README.md"}
	argsExpected = []string{"upload", "-reference=19", "profiler/README.md"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 1)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"curl", "-v=4", "-reference=4", "-samples=4", "http://labomedia.org"}
	expected = []string{"-v=4", "curl", "-reference=4", "-samples=4", "http://labomedia.org"}
	argsExpected = []string{"curl", "-reference=4", "-samples=4", "http://labomedia.org"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 4)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)

	args = []string{"run", "-v=4", "--reference", "8", "php", "vd.php", "--config=foo", "--foo", "bar"}
	expected = []string{"-v=4", "run", "--reference", "8", "php", "vd.php", "--config=foo", "--foo", "bar"}
	argsExpected = []string{"run", "--reference", "8", "php", "vd.php", "--config=foo", "--foo", "bar"}
	sorted = testApp.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = testApp.parseArgs(args)
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("v"), Equals, 4)
	c.Check(ctx.Bool("quiet"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, argsExpected)
}

func (ts *CliEnhancementSuite) TestFixAndParseArgsApplicationVerbosityFlag(c *C) {
	defaultLogLevel := terminal.GetLogLevel()
	defer terminal.SetLogLevel(defaultLogLevel)

	testApp := Application{
		Flags: []Flag{
			VerbosityFlag("log-level", "verbose", "v"),
		},
		Commands: []*Command{
			{
				Name: "envs",
				Flags: []Flag{
					&StringFlag{
						Name:    "project",
						Aliases: []string{"p"},
					},
				},
			},
		},
	}

	cases := []struct {
		arg           string
		expectedLevel int
	}{
		{"--log-level=5", 5},
		{"--verbose", 3},
		{"-vvv", 4},
		{"-vv", 3},
		{"-v", 2},
		{"-v=3", 3},
	}

	for _, tt := range cases {
		args := []string{tt.arg, "-p", "agb6vnth4arfo", "envs"}
		expected := []string{tt.arg, "envs", "-p", "agb6vnth4arfo"}
		sorted := testApp.fixArgs(args)
		c.Assert(sorted, DeepEquals, expected)
		fs, _ := testApp.parseArgs(args)
		ctx := NewContext(&testApp, fs, nil)

		c.Check(terminal.GetLogLevel(), Equals, tt.expectedLevel)
		c.Check(ctx.IsSet("log-level"), Equals, true)

		cmd := testApp.Command(ctx.Args().first())
		fs, _ = cmd.parseArgs(ctx.Args().Tail(), []string{})
		ctx = NewContext(&testApp, fs, nil)

		c.Check(ctx.String("project"), Equals, "agb6vnth4arfo")
	}
}

func (ts *CliEnhancementSuite) TestFixAndParseArgsCommand(c *C) {
	var (
		args     = []string{"-reference=4", "--samples=10", "-t", "file1", "-s=", "5", "-H='Host: foo'", "foo"}
		expected []string

		ctx    *Context
		err    error
		sorted []string
		fs     *flag.FlagSet
	)

	expected = []string{"-reference=4", "--samples=10", "-test", "-samples", "5", "-H='Host: foo'", "--", "file1", "foo"}
	sorted = curlCmd.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = curlCmd.parseArgs(args, []string{})
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 4)
	c.Check(ctx.Int("samples"), Equals, 5)
	c.Check(ctx.Bool("test"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, []string{"file1", "foo"})

	expected = []string{"-reference=4", "--samples=10", "-test", "-samples", "5", "-H='Host: foo'", "--", "file1", "foo"}
	sorted = uploadCmd.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = uploadCmd.parseArgs(args, []string{})
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 4)
	c.Check(ctx.Int("samples"), Equals, 5)
	c.Check(ctx.Bool("test"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, []string{"file1", "foo"})

	expected = append([]string{"--"}, args...)
	sorted = fooCmd.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = fooCmd.parseArgs(args, []string{})
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 0)
	c.Check(ctx.Int("samples"), Equals, 0)
	c.Check(ctx.Bool("test"), Equals, false)
	c.Check(ctx.Args().Slice(), DeepEquals, args)

	expected = []string{"-reference=4", "--samples=10", "-test", "--", "file1", "-s=", "5", "-H='Host: foo'", "foo"}
	sorted = runCmd.fixArgs(args)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = runCmd.parseArgs(args, []string{})
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 4)
	c.Check(ctx.Int("samples"), Equals, 10)
	c.Check(ctx.Bool("test"), Equals, true)
	c.Check(ctx.Args().Slice(), DeepEquals, []string{"file1", "-s=", "5", "-H='Host: foo'", "foo"})

	dashDashArgs := []string{"-reference=4", "-s=", "5", "--", "--samples=10", "file1", "-f=", "3", "foo"}
	expected = []string{"-reference=4", "-samples", "5", "--", "--samples=10", "file1", "-f=", "3", "foo"}
	sorted = curlCmd.fixArgs(dashDashArgs)
	c.Assert(sorted, DeepEquals, expected)
	fs, _ = curlCmd.parseArgs(dashDashArgs, []string{})
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 4)
	c.Check(ctx.Int("samples"), Equals, 5)
	c.Check(ctx.Args().Slice(), DeepEquals, []string{"--samples=10", "file1", "-f=", "3", "foo"})

	weirdArgs := []string{"-reference=4", "--unknown", "-r=", "-s=", "5", "--samples=10", "file1", "-f=", "3", "foo"}
	expected = []string{"-reference=4", "--unknown", "-reference", "-samples", "5", "--samples=10", "-f=", "3", "--", "file1", "foo"}
	sorted = curlCmd.fixArgs(weirdArgs)
	c.Assert(sorted, DeepEquals, expected)
	fs, err = curlCmd.parseArgs(weirdArgs, []string{})
	c.Check(err, Not(IsNil))
	ctx = NewContext(&testApp, fs, nil)
	c.Check(ctx.Int("reference"), Equals, 4)
	c.Check(ctx.Int("samples"), Equals, 0)
	c.Check(ctx.Args().Slice(), DeepEquals, []string{"-reference", "-samples", "5", "--samples=10", "-f=", "3", "file1", "foo"})
}

func (ts *CliEnhancementSuite) TestCheckRequiredFlagsSuccess(c *C) {
	flags := []Flag{
		&StringFlag{
			Name:     "required",
			Required: true,
		},
		&StringFlag{
			Name: "optional",
		},
	}

	set := flag.NewFlagSet("test", 0)
	for _, f := range flags {
		f.Apply(set)
	}

	e := set.Parse([]string{"--required", "foo"})
	c.Assert(e, IsNil)

	err := checkRequiredFlags(flags, set)
	c.Assert(err, IsNil)
}

func (ts *CliEnhancementSuite) TestCheckRequiredFlagsFailure(c *C) {
	flags := []Flag{
		&StringFlag{
			Name:     "required",
			Required: true,
		},
		&StringFlag{
			Name: "optional",
		},
	}

	set := flag.NewFlagSet("test", 0)
	for _, f := range flags {
		f.Apply(set)
	}

	e := set.Parse([]string{"--optional", "foo"})
	c.Assert(e, IsNil)

	err := checkRequiredFlags(flags, set)
	c.Assert(err, Not(IsNil))
}

func (ts *CliEnhancementSuite) TestFlagsValidation(c *C) {
	validatorHasBeenCalled, subValidatorHasBeenCalled := false, false

	app := Application{
		Flags: []Flag{
			&StringFlag{
				Name: "foo",
				Validator: func(context *Context, s string) error {
					validatorHasBeenCalled = true

					return nil
				},
			},
			&StringFlag{
				Name: "bar",
				Validator: func(context *Context, s string) error {
					if s != "bar" {
						return errors.New("invalid")
					}
					return nil
				},
			},
		},
		Commands: []*Command{
			{
				Name: "test",
				Flags: []Flag{
					&StringFlag{
						Name: "sub-foo",
						Validator: func(context *Context, s string) error {
							subValidatorHasBeenCalled = true

							return nil
						},
					},
					&StringFlag{
						Name: "sub-bar",
						Validator: func(context *Context, s string) error {
							if s != "bar" {
								return errors.New("invalid")
							}
							return nil
						},
					},
				},
				Action: func(c *Context) error {
					fmt.Println("sub-foo:", c.String("sub-foo"))
					return nil
				},
			},
		},
	}

	c.Assert(app.Run([]string{"app", "--foo=bar"}), IsNil)
	c.Assert(validatorHasBeenCalled, Equals, true)
	c.Assert(app.Run([]string{"app", "--bar=bar"}), IsNil)
	c.Assert(app.Run([]string{"app", "--bar=toto"}), ErrorMatches, "invalid value for flag \"bar\".*")

	c.Assert(app.Run([]string{"app", "test", "--sub-foo=bar"}), IsNil)
	c.Assert(subValidatorHasBeenCalled, Equals, true)
	c.Assert(app.Run([]string{"app", "test", "--sub-bar=bar"}), IsNil)
	c.Assert(app.Run([]string{"app", "test", "--sub-bar=toto"}), ErrorMatches, ".*invalid value for flag \"sub-bar\".*")
}
