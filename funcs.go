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

import "github.com/posener/complete"

// ShellCompleteFunc is an action to execute when the shell completion flag is set
type ShellCompleteFunc func(*Context, complete.Args) []string

// BeforeFunc is an action to execute before any subcommands are run, but after
// the context is ready if a non-nil error is returned, no subcommands are run
type BeforeFunc func(*Context) error

// AfterFunc is an action to execute after any subcommands are run, but after the
// subcommand has finished it is run even if Action() panics
type AfterFunc func(*Context) error

// ActionFunc is the action to execute when no subcommands are specified
type ActionFunc func(*Context) error

// CommandNotFoundFunc is executed if the proper command cannot be found
type CommandNotFoundFunc func(*Context, string) error

// DescriptionFunc is used by the help generation to display a description when
// its computation is intensive or needs runtime information
type DescriptionFunc func(*Command, *Application) string

// FlagStringFunc is used by the help generation to display a flag, which is
// expected to be a single line.
type FlagStringFunc func(Flag) string
