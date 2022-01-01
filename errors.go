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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/symfony-cli/terminal"
)

// OsExiter is the function used when the app exits. If not set defaults to os.Exit.
var OsExiter = os.Exit

// MultiError is an error that wraps multiple errors.
type MultiError interface {
	error
	// Errors returns a copy of the errors slice
	Errors() []error
}

// newMultiError creates a new MultiError. Pass in one or more errors.
func newMultiError(err ...error) MultiError {
	ret := multiError(err)
	return &ret
}

type multiError []error

// Error implements the error interface.
func (m *multiError) Error() string {
	errs := make([]string, len(*m))
	for i, err := range *m {
		errs[i] = err.Error()
	}

	return strings.Join(errs, "\n")
}

// Errors returns a copy of the errors slice
func (m *multiError) Errors() []error {
	errs := make([]error, len(*m))
	for _, err := range *m {
		errs = append(errs, err)
	}
	return errs
}

// ExitCoder is the interface checked by `App` and `Command` for a custom exit
// code
type ExitCoder interface {
	error
	ExitCode() int
}

type exitError struct {
	exitCode int
	message  string
}

// Exit wraps a message and exit code into an ExitCoder suitable for handling by
// HandleExitCoder
func Exit(message string, exitCode int) ExitCoder {
	return &exitError{
		exitCode: exitCode,
		message:  message,
	}
}

func (ee *exitError) Error() string {
	return ee.message
}

func (ee *exitError) ExitCode() int {
	return ee.exitCode
}

// HandleExitCoder checks if the error fulfills the ExitCoder interface, and if
// so prints the error to stderr (if it is non-empty) and calls OsExiter with the
// given exit code.  If the given error is a MultiError, then this func is
// called on all members of the Errors slice.
func HandleExitCoder(err error) {
	if err == nil {
		return
	}

	HandleError(err)
	OsExiter(handleExitCode(err))
}

func HandleError(err error) {
	if err == nil {
		return
	}

	if multiErr, ok := err.(MultiError); ok {
		for _, merr := range multiErr.Errors() {
			HandleError(merr)
		}
		return
	}

	if msg := err.Error(); msg != "" {
		var buf bytes.Buffer

		if terminal.IsVerbose() && isGoRun() {
			msg = fmt.Sprintf("[%s]\n%s", reflect.TypeOf(err), err)
		}

		buf.WriteString(terminal.FormatBlockMessage("error", msg))

		if terminal.IsVerbose() {
			var traceBuf bytes.Buffer
			if FormatErrorChain(&traceBuf, err, !isGoRun()) {
				buf.WriteString("\n<comment>Error trace:</>\n")
				buf.Write(traceBuf.Bytes())
			}
		}

		terminal.Eprint(buf.String())
	}
}

func handleExitCode(err error) int {
	if exitErr, ok := err.(ExitCoder); ok {
		return exitErr.ExitCode()
	}

	if multiErr, ok := err.(MultiError); ok {
		for _, merr := range multiErr.Errors() {
			if exitErr, ok := merr.(ExitCoder); ok {
				return exitErr.ExitCode()
			}
		}
	}

	return 1
}

type IncorrectUsageError struct {
	ParentError error
}

func (e IncorrectUsageError) Cause() error {
	return e.ParentError
}

func (e IncorrectUsageError) Error() string {
	return fmt.Sprintf("Incorrect usage: %s", e.ParentError.Error())
}

func isGoRun() bool {
	// Unfortunately, Golang does not expose that we are currently using go run
	// So we detect the main binary is (or used to be ;)) "go" and then the
	// current binary is within a temp "go-build" directory.
	_, exe := filepath.Split(os.Getenv("_"))
	argv0, _ := os.Executable()

	return exe == "go" && strings.Contains(argv0, "go-build")
}
