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
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func CurrentBinaryName() string {
	argv0, err := os.Executable()
	if nil != err {
		return ""
	}

	return filepath.Base(argv0)
}

func CurrentBinaryPath() (string, error) {
	argv0, err := os.Executable()
	if nil != err {
		return argv0, errors.WithStack(err)
	}
	return argv0, nil
}

func CurrentBinaryInvocation() (string, error) {
	if len(os.Args) == 0 || os.Args[0] == "" {
		return "", errors.New("no binary invokation found")
	}

	return os.Args[0], nil
}

func (c *Context) CurrentBinaryPath() string {
	path, err := CurrentBinaryPath()
	if err != nil {
		panic(err)
	}

	return path
}

func (c *Context) CurrentBinaryInvocation() string {
	invocation, err := CurrentBinaryInvocation()
	if err != nil {
		panic(err)
	}

	return invocation
}
