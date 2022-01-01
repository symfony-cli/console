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

import "sort"

type CommandCategories interface {
	// AddCommand adds a command to a category, creating a new category if necessary.
	AddCommand(category string, command *Command)
	// Categories returns a copy of the category slice
	Categories() []CommandCategory
}

type commandCategories []*commandCategory

func newCommandCategories() CommandCategories {
	ret := commandCategories([]*commandCategory{})
	return &ret
}

func (c *commandCategories) Less(i, j int) bool {
	return (*c)[i].Name() < (*c)[j].Name()
}

func (c *commandCategories) Len() int {
	return len(*c)
}

func (c *commandCategories) Swap(i, j int) {
	(*c)[i], (*c)[j] = (*c)[j], (*c)[i]
}

func (c *commandCategories) AddCommand(category string, command *Command) {
	for _, commandCategory := range []*commandCategory(*c) {
		if commandCategory.name == category {
			commandCategory.commands = append(commandCategory.commands, command)
			return
		}
	}
	newVal := commandCategories(append(*c,
		&commandCategory{name: category, commands: []*Command{command}}))
	(*c) = newVal
}

func (c *commandCategories) Categories() []CommandCategory {
	ret := make([]CommandCategory, len(*c))
	for i, cat := range *c {
		ret[i] = cat
	}
	return ret
}

// CommandCategory is a category containing commands.
type CommandCategory interface {
	// Name returns the category name string
	Name() string
	// VisibleCommands returns a slice of the Commands with Hidden=false
	VisibleCommands() []*Command
}

type commandCategory struct {
	name     string
	commands []*Command
}

func (c *commandCategory) Name() string {
	return c.name
}

func (c *commandCategory) VisibleCommands() []*Command {
	if c.commands == nil {
		return []*Command{}
	}

	ret := []*Command{}
	for _, command := range c.commands {
		if command.Hidden == nil || !command.Hidden() {
			ret = append(ret, command)
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})

	return ret
}
