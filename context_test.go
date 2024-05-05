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
	"time"

	"github.com/symfony-cli/terminal"
	. "gopkg.in/check.v1"
)

type ContextSuite struct{}

var _ = Suite(&ContextSuite{})

func (cs *ContextSuite) TestNewContext(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	set.Int64("myflagInt64", int64(12), "doc")
	set.Uint("myflagUint", uint(93), "doc")
	set.Uint64("myflagUint64", uint64(93), "doc")
	set.Float64("myflag64", float64(17), "doc")
	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Int("myflag", 42, "doc")
	globalSet.Int64("myflagInt64", int64(42), "doc")
	globalSet.Uint("myflagUint", uint(33), "doc")
	globalSet.Uint64("myflagUint64", uint64(33), "doc")
	globalSet.Float64("myflag64", float64(47), "doc")
	globalCtx := NewContext(nil, globalSet, nil)
	command := &Command{Name: "mycommand"}
	ctx := NewContext(nil, set, globalCtx)
	ctx.Command = command
	c.Assert(ctx.Int("myflag"), Equals, 12)
	c.Assert(ctx.Int64("myflagInt64"), Equals, int64(12))
	c.Assert(ctx.Uint("myflagUint"), Equals, uint(93))
	c.Assert(ctx.Uint64("myflagUint64"), Equals, uint64(93))
	c.Assert(ctx.Float64("myflag64"), Equals, float64(17))
	c.Assert(ctx.Command.Name, Equals, "mycommand")
}

func (cs *ContextSuite) TestContext_Int(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Int("myflag", 12, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Int("top-flag", 13, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Int("myflag"), Equals, 12)
	c.Assert(ctx.Int("top-flag"), Equals, 13)
}

func (cs *ContextSuite) TestContext_Int64(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Int64("myflagInt64", 12, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Int64("top-flag", 13, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Int64("myflagInt64"), Equals, int64(12))
	c.Assert(ctx.Int64("top-flag"), Equals, int64(13))
}

func (cs *ContextSuite) TestContext_Uint(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Uint("myflagUint", uint(13), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Uint("top-flag", uint(14), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Uint("myflagUint"), Equals, uint(13))
	c.Assert(ctx.Uint("top-flag"), Equals, uint(14))
}

func (cs *ContextSuite) TestContext_Uint64(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Uint64("myflagUint64", uint64(9), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Uint64("top-flag", uint64(10), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Uint64("myflagUint64"), Equals, uint64(9))
	c.Assert(ctx.Uint64("top-flag"), Equals, uint64(10))
}

func (cs *ContextSuite) TestContext_Float64(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Float64("myflag", float64(17), "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Float64("top-flag", float64(18), "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Float64("myflag"), Equals, float64(17))
	c.Assert(ctx.Float64("top-flag"), Equals, float64(18))
}

func (cs *ContextSuite) TestContext_Duration(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Duration("myflag", 12*time.Second, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Duration("top-flag", 13*time.Second, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Duration("myflag"), Equals, 12*time.Second)
	c.Assert(ctx.Duration("top-flag"), Equals, 13*time.Second)
}

func (cs *ContextSuite) TestContext_String(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.String("myflag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.String("top-flag", "hai veld", "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.String("myflag"), Equals, "hello world")
	c.Assert(ctx.String("top-flag"), Equals, "hai veld")
}

func (cs *ContextSuite) TestContext_Bool(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	c.Assert(ctx.Bool("myflag"), Equals, false)
	c.Assert(ctx.Bool("top-flag"), Equals, true)
}

func (cs *ContextSuite) TestContext_Args(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	ctx := NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	c.Assert(ctx.Args().Len(), Equals, 2)
	c.Assert(ctx.Bool("myflag"), Equals, true)
}

func (cs *ContextSuite) TestContext_NArg(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("myflag", false, "doc")
	ctx := NewContext(nil, set, nil)
	set.Parse([]string{"--myflag", "bat", "baz"})
	c.Assert(ctx.NArg(), Equals, 2)
}

func (cs *ContextSuite) TestContext_HasFlag(c *C) {
	app := &Application{
		Flags: []Flag{
			&StringFlag{Name: "top-flag"},
		},
		Commands: []*Command{
			{
				Name:    "hello",
				Aliases: []*Alias{{Name: "hi"}},
				Flags: []Flag{
					&StringFlag{Name: "one-flag"},
				},
			},
		},
	}
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(app, parentSet, nil)
	ctx := NewContext(app, set, parentCtx)

	c.Assert(parentCtx.HasFlag("top-flag"), Equals, true)
	c.Assert(parentCtx.HasFlag("one-flag"), Equals, false)
	c.Assert(parentCtx.HasFlag("bogus"), Equals, false)

	parentCtx.Command = app.Commands[0]

	c.Assert(ctx.HasFlag("top-flag"), Equals, true)
	c.Assert(ctx.HasFlag("one-flag"), Equals, true)
	c.Assert(ctx.HasFlag("bogus"), Equals, false)
}

func (cs *ContextSuite) TestContext_IsSet(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("one-flag", false, "doc")
	set.Bool("two-flag", false, "doc")
	set.String("three-flag", "hello world", "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)

	set.Parse([]string{"--one-flag", "--two-flag", "frob"})
	parentSet.Parse([]string{"--top-flag"})

	c.Assert(ctx.IsSet("one-flag"), Equals, true)
	c.Assert(ctx.IsSet("two-flag"), Equals, true)
	c.Assert(ctx.IsSet("three-flag"), Equals, false)
	c.Assert(ctx.IsSet("top-flag"), Equals, true)
	c.Assert(ctx.IsSet("bogus"), Equals, false)
}

func (cs *ContextSuite) TestContext_Set(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Int("int", 5, "an int")
	ctx := NewContext(nil, set, nil)

	ctx.Set("int", "1")
	c.Assert(ctx.Int("int"), Equals, 1)
}

func (cs *ContextSuite) TestContext_Set_AppFlags(c *C) {
	defer terminal.SetLogLevel(1)

	app := &Application{
		Commands: []*Command{
			{
				Name: "foo",
				Action: func(ctx *Context) error {
					err := ctx.Set("log-level", "4")
					c.Assert(err, IsNil)

					return nil
				},
			},
		},
	}
	app.Run([]string{"cmd", "foo"})
}

func (cs *ContextSuite) TestContext_Lineage(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--local-flag"})
	parentSet.Parse([]string{"--top-flag"})

	lineage := ctx.Lineage()
	c.Assert(len(lineage), Equals, 2)
	c.Assert(lineage[0], Equals, ctx)
	c.Assert(lineage[1], Equals, parentCtx)
}

func (cs *ContextSuite) TestContext_lookupFlagSet(c *C) {
	set := flag.NewFlagSet("test", 0)
	set.Bool("local-flag", false, "doc")
	parentSet := flag.NewFlagSet("test", 0)
	parentSet.Bool("top-flag", true, "doc")
	parentCtx := NewContext(nil, parentSet, nil)
	ctx := NewContext(nil, set, parentCtx)
	set.Parse([]string{"--local-flag"})
	parentSet.Parse([]string{"--top-flag"})

	fs := lookupFlagSet("top-flag", ctx)
	c.Assert(fs, Equals, parentCtx.flagSet)

	fs = lookupFlagSet("local-flag", ctx)
	c.Assert(fs, Equals, ctx.flagSet)

	if fs := lookupFlagSet("frob", ctx); fs != nil {
		c.Fail()
	}
}
