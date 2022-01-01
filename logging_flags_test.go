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

	"github.com/rs/zerolog"
	"github.com/symfony-cli/terminal"
	. "gopkg.in/check.v1"
)

type LoggingFlagsSuite struct{}

var _ = Suite(&LoggingFlagsSuite{})

func (ts *LoggingFlagsSuite) TestLogLevel(c *C) {
	defer terminal.SetLogLevel(1)
	value := &logLevelValue{}
	var err error

	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.ErrorLevel)

	err = value.Set("foo")
	c.Assert(err, Not(IsNil))
	c.Assert(err, ErrorMatches, ".* parsing \"foo\": invalid syntax")
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.ErrorLevel)

	err = value.Set("4")
	c.Assert(err, IsNil)
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.DebugLevel)

	err = value.Set("2")
	c.Assert(err, IsNil)
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.WarnLevel)

	err = value.Set("9")
	c.Assert(err, Not(IsNil))
	c.Assert(err.Error(), Equals, "The provided verbosity level '9' is not in the range [1,4]")
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.WarnLevel)
}

func (ts *LoggingFlagsSuite) TestLogLevelShortcuts(c *C) {
	defer terminal.SetLogLevel(1)
	fs := flag.NewFlagSet("foo", flag.ExitOnError)
	fs.Var(&logLevelValue{}, "log-level", "FooBar")

	value := newLogLevelShortcutValue(fs, "log-level", 3)
	var err error

	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.ErrorLevel)

	err = value.Set("true")
	c.Assert(err, IsNil)
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.InfoLevel)

	err = value.Set("false")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, ".* invalid syntax")

	err = value.Set("2")
	c.Assert(err, IsNil)
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.WarnLevel)

	err = value.Set("9")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "The provided verbosity level '9' is not in the range [1,4]")
	c.Assert(terminal.Logger.GetLevel(), Equals, zerolog.WarnLevel)
}
