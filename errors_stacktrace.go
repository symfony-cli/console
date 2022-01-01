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
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/symfony-cli/terminal"
)

type WrappedPanic struct {
	msg   interface{}
	stack []uintptr
}

func WrapPanic(msg interface{}) error {
	if err, ok := msg.(error); ok {
		if _, hasStackTrace := err.(stackTracer); hasStackTrace {
			return err
		}
	}

	const depth = 100
	var pcs [depth]uintptr
	// 5 is the number of call functions made after a panic to reach the recover with the call this function :)
	n := runtime.Callers(0, pcs[:])

	return WrappedPanic{
		msg:   msg,
		stack: skipPanicsFromStacktrace(pcs[0:n]),
	}
}

func (p WrappedPanic) Error() string {
	return fmt.Sprintf("panic: %v", p.msg)
}

func (p WrappedPanic) Cause() error {
	if err, ok := p.msg.(error); ok {
		return err
	}

	return nil
}

func (p WrappedPanic) StackTrace() errors.StackTrace {
	f := make([]errors.Frame, len(p.stack))
	for i, n := 0, len(f); i < n; i++ {
		f[i] = errors.Frame(p.stack[i])
	}
	return f
}

func skipPanicsFromStacktrace(stack []uintptr) []uintptr {
	newStack := make([]uintptr, len(stack))
	pos := 0
	for i, n := 0, len(newStack); i < n; i++ {
		f := stack[i]

		pc := uintptr(f) - 1
		fn := runtime.FuncForPC(pc)
		// we found a panic call, let's strip the previous frames
		if fn.Name() == "runtime.gopanic" {
			pos = 0
			continue
		}

		newStack[pos] = f
		pos++
	}

	return newStack[:pos]
}

type causer interface {
	Cause() error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func pc(f errors.Frame) uintptr { return uintptr(f) - 1 }

func FormatErrorChain(buf *bytes.Buffer, err error, trimPaths bool) bool {
	var parent error

	// Go up in the error tree following causes.
	// Each new cause is kept until we don't have new ones
	// or we find one with a stacktrace, in this case this
	// one must be treated on its own.
	for cause := err; cause != nil; {
		errWithCause, ok := cause.(causer)
		if !ok {
			break
		}
		cause = errWithCause.Cause()
		if _, newClauseHasStackTrace := cause.(stackTracer); newClauseHasStackTrace {
			parent = cause
			break
		}
	}

	var st errors.StackTrace
	if errWithStackTrace, hasStackTrace := err.(stackTracer); hasStackTrace {
		st = errWithStackTrace.StackTrace()
	} else if parent != nil {
		if errWithStackTrace, hasStackTrace := parent.(stackTracer); hasStackTrace {
			st = errWithStackTrace.StackTrace()

			if errWithCause, ok := parent.(causer); ok {
				parent = errWithCause.Cause()
			} else {
				parent = nil
			}
		}
	} else {
		return false
	}

	msg := err.Error()
	if parent != nil {
		msg = strings.TrimSuffix(msg, fmt.Sprintf(": %s", parent.Error()))
	}

	buf.WriteString(terminal.FormatBlockMessage("error", msg))

	for _, f := range st {
		buf.WriteString("\n")
		pc := pc(f)
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			buf.WriteString("unknown")
		} else {
			file, line := fn.FileLine(pc)
			if trimPaths {
				file = trimGOPATH(fn.Name(), file)
			}
			buf.WriteString(fmt.Sprintf("%s\n\t<info>%s:%d</>", fn.Name(), file, line))
		}
	}

	buf.WriteByte('\n')
	if parent != nil {
		buf.WriteByte('\n')
		buf.WriteString("Previous error:\n")
		FormatErrorChain(buf, parent, trimPaths)
	}

	return true
}

func trimGOPATH(name, file string) string {
	// Here we want to get the source file path relative to the compile time
	// GOPATH. As of Go 1.6.x there is no direct way to know the compiled
	// GOPATH at runtime, but we can infer the number of path segments in the
	// GOPATH. We note that fn.Name() returns the function name qualified by
	// the import path, which does not include the GOPATH. Thus we can trim
	// segments from the beginning of the file path until the number of path
	// separators remaining is one more than the number of path separators in
	// the function name. For example, given:
	//
	//    GOPATH     /home/user
	//    file       /home/user/src/pkg/sub/file.go
	//    fn.Name()  pkg/sub.Type.Method
	//
	// We want to produce:
	//
	//    pkg/sub/file.go
	//
	// From this we can easily see that fn.Name() has one less path separator
	// than our desired output. We count separators from the end of the file
	// path until it finds two more than in the function name and then move
	// one character forward to preserve the initial path segment without a
	// leading separator.
	const sep = "/"
	goal := strings.Count(name, sep) + 2
	i := len(file)
	for n := 0; n < goal; n++ {
		i = strings.LastIndex(file[:i], sep)
		if i == -1 {
			// not enough separators found, set i so that the slice expression
			// below leaves file unmodified
			i = -len(sep)
			break
		}
	}
	// get back to 0 or trim the leading separator
	file = file[i+len(sep):]
	return file
}
