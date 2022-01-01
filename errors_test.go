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
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/symfony-cli/terminal"
	. "gopkg.in/check.v1"
)

type ErrorsSuite struct{}

var _ = Suite(&ErrorsSuite{})

func mockOsExiter(fn func(int)) func(int) {
	once := &sync.Once{}
	return func(rc int) {
		once.Do(func() {
			fn(rc)
		})
	}
}

func (es *ErrorsSuite) TestHandleExitCoder_nil(c *C) {
	exitCode := 0
	called := false

	OsExiter = mockOsExiter(func(rc int) {
		exitCode = rc
		called = true
	})

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(nil)

	c.Assert(exitCode, Equals, 0)
	c.Assert(called, Equals, false)
}

func (es *ErrorsSuite) TestHandleExitCoder_ExitCoder(c *C) {
	exitCode := 0
	called := false

	OsExiter = mockOsExiter(func(rc int) {
		exitCode = rc
		called = true
	})

	defer func() { OsExiter = fakeOsExiter }()

	HandleExitCoder(Exit("galactic perimeter breach", 9))

	c.Assert(exitCode, Equals, 9)
	c.Assert(called, Equals, true)
}

func (es *ErrorsSuite) TestHandleExitCoder_MultiErrorWithExitCoder(c *C) {
	exitCode := 0
	called := false

	OsExiter = mockOsExiter(func(rc int) {
		exitCode = rc
		called = true
	})

	defer func() { OsExiter = fakeOsExiter }()

	exitErr := Exit("galactic perimeter breach", 9)
	err := newMultiError(errors.New("wowsa"), exitErr, errors.New("egad"))
	HandleExitCoder(err)

	c.Assert(exitCode, Equals, 9)
	c.Assert(called, Equals, true)
}

func (es *ErrorsSuite) TestHandleExitCoder_MultiErrorWithoutExitCoder(c *C) {
	exitCode := 0
	called := false

	OsExiter = func(rc int) {
		if !called {
			exitCode = rc
			called = true
		}
	}

	defer func() { OsExiter = fakeOsExiter }()

	err := newMultiError(errors.New("wowsa"), errors.New("egad"))
	HandleExitCoder(err)

	c.Assert(exitCode, Equals, 1)
	c.Assert(called, Equals, true)
}

func (es *ErrorsSuite) TestHandleExitCoder_ErrorWithMessage(c *C) {
	exitCode := 0
	called := false

	OsExiter = mockOsExiter(func(rc int) {
		exitCode = rc
		called = true
	})
	previousStderr := terminal.Stderr
	defer func() {
		OsExiter = fakeOsExiter
		terminal.Stderr = previousStderr
	}()

	bufferStderr := new(bytes.Buffer)
	formatter := terminal.NewFormatter()
	terminal.Stderr = terminal.NewOutput(bufferStderr, formatter)

	err := errors.New("gourd havens")
	HandleExitCoder(err)

	c.Assert(exitCode, Equals, 1)
	c.Assert(called, Equals, true)
	c.Assert(strings.Contains(bufferStderr.String(), "gourd havens"), Equals, true)
}

func (es *ErrorsSuite) TestHandleExitCoder_ErrorWithoutMessage(c *C) {
	exitCode := 0
	called := false

	OsExiter = mockOsExiter(func(rc int) {
		exitCode = rc
		called = true
	})
	previousStderr := terminal.Stderr

	defer func() {
		OsExiter = fakeOsExiter
		terminal.Stderr = previousStderr
	}()

	bufferStderr := new(bytes.Buffer)
	formatter := terminal.NewFormatter()
	terminal.Stderr = terminal.NewOutput(bufferStderr, formatter)

	err := errors.New("")
	HandleExitCoder(err)

	c.Assert(exitCode, Equals, 1)
	c.Assert(called, Equals, true)
	c.Assert(bufferStderr.String(), Equals, "")
}
