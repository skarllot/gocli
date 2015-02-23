/*
* Copyright 2015 Fabrício Godoy
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package gocli

import (
	"errors"
	"fmt"
	"strings"
)

// A Command represents a CLI command.
type Command struct {
	// Name of command.
	Name string
	// Short help for command.
	Short string
	// Long help for command.
	Long string
	// Run is a function that is executed at user call.
	Run func(cmd *Command, args []string)
	// Load is a function that is executed when a parent command is called.
	Load func(cmd *Command)

	childs Commands
	parent *Command
	exit   *Command
	help   *Command
}

// A Commands is a slice of Command pointers
type Commands []*Command

// RecurseGet represents an interaction from RecurseParent.
// The successful returns true whether the value was got successfully.
type RecurseGet func(cmd *Command, first, last bool) (successful bool)

// AddChild adds one or more child commands.
func (c *Command) AddChild(cmds ...*Command) {
	for _, v := range cmds {
		if v == c {
			panic("Command can't be a child of itself")
		}

		v.parent = c
		c.childs = append(c.childs, v)
	}
}

// Execute an interactive CLI until users exits.
func (c *Command) Execute() error {
	if c.Run != nil && len(c.childs) != 0 {
		return errors.New("Current command should not define an action and has childs")
	}
	if c.Load != nil && c.Run != nil {
		return errors.New("Only Run or Load functions should be defined, not both")
	}

	if c.Load != nil {
		c.Load(c)
	}
	if c.Run == nil && len(c.childs) == 0 {
		return errors.New("Current command has neither action and childs")
	}

	if !RecurseParents(c, func(cmd *Command, first, last bool) bool {
		if cmd.exit != nil {
			c.exit = cmd.exit
			return true
		}
		return false
	}) {
		ExitCommand(c)
	}
	if !RecurseParents(c, func(cmd *Command, first, last bool) bool {
		if cmd.help != nil {
			c.help = cmd.help
			return true
		}
		return false
	}) {
		HelpCommand(c)
	}

	for {
		if c.parent == nil {
			fmt.Printf("%s>", c.Name)
		} else {
			prompt := ""
			RecurseParents(c, func(cmd *Command, first, last bool) bool {
				if first {
					prompt = fmt.Sprintf("%s)>", c.Name)
					return false
				}
				if last {
					prompt = fmt.Sprintf("%s(%s", parent.Name, prompt)
					return false
				}
				prompt = fmt.Sprintf("%s/%s", parent.Name, prompt)
				return false
			})
			fmt.Print(prompt)
		}

		input, err := readString()
		if err != nil {
			fmt.Println("error:", err.Error())
			continue
		}

		input = strings.Trim(input, " ")
		if len(input) == 0 {
			continue
		}

		args := strings.Split(input, " ")
		selCmd := c.Find(args[0])
		if selCmd == nil {
			fmt.Printf(
				"Invalid command, type %s for available commands\n",
				c.help.Name)
			continue
		}
		if selCmd == c.exit {
			break
		}
		if selCmd == c.help {
			c.ShowHelp()
			continue
		}
		if selCmd.Run == nil &&
			selCmd.Load == nil &&
			len(selCmd.childs) == 0 {
			fmt.Printf("Missing action for %s command\n", selCmd.Name)
			break
		}
		if selCmd.Run == nil {
			err = selCmd.Execute()
			if err != nil {
				return err
			}
		}

		selCmd.Run(selCmd, args[1:])
	}

	return nil
}

// ExitCommand creates a new exit command and adds it to selected parent
// command.
func ExitCommand(parent *Command) *Command {
	cmd := &Command{
		Name:  "exit",
		Short: "Exit from this program",
	}
	parent.exit = cmd
	parent.AddChild(cmd)
	return cmd
}

// Find finds a command by its name.
func (c *Command) Find(name string) *Command {
	for _, v := range c.childs {
		if v.Name == name {
			return v
		}
	}

	return nil
}

// HelpCommand creates a new help command and adds it to selected parent
// command.
func HelpCommand(parent *Command) *Command {
	cmd := &Command{
		Name:  "help",
		Short: "help about this program",
	}
	parent.help = cmd
	parent.AddChild(cmd)
	return cmd
}

// RecurseParents interates from current commands to all its parents until
// the function returns true.
func RecurseParents(c *Command, f RecurseGet) (successful bool) {
	if f(c, true, false) {
		return true
	}

	parent := c.parent
	for {
		if parent.parent == nil {
			return f(parent, false, true)
		}
		if f(parent, false, false) {
			return true
		}
		parent = parent.parent
	}

	return false
}

// ShowHelp prints out short help of all child commands.
func (c *Command) ShowHelp() {
	for _, v := range c.childs {
		fmt.Printf("%s\t\t%s\n", v.Name, v.Short)
	}
}
