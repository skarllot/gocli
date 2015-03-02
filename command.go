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
	// Help for command.
	Help string
	// Paremeters accepted by command.
	Parameters []Parameter
	// Run is a function that is executed at user call.
	Run func(cmd *Command, args []string)
	// Load is a function that is executed when a parent command is called.
	Load func(cmd *Command)

	childs Commands
	parent *Command
	exit   *Command
	help   *Command
}

// A Parameter represents a command parameter.
type Parameter struct {
	// Name of parameter.
	Name string
	// Help for parameter.
	Help string
	// Optional defines if current parameter is optional.
	Optional bool
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
func (self *Command) Execute() error {
	if self.Run != nil && len(self.childs) != 0 {
		return errors.New("Current command should not define an action and has childs")
	}
	if self.Load != nil && self.Run != nil {
		return errors.New("Only Run or Load functions should be defined, not both")
	}

	if self.Load != nil {
		self.Load(self)
	}
	if self.Run == nil && len(self.childs) == 0 {
		return errors.New("Current command has neither action and childs")
	}

	if !RecurseParents(self, func(cmd *Command, first, last bool) bool {
		if cmd.help != nil {
			self.help = cmd.help
			self.AddChild(cmd.help)
			return true
		}
		return false
	}) {
		HelpCommand(self)
	}
	if !RecurseParents(self, func(cmd *Command, first, last bool) bool {
		if cmd.exit != nil {
			self.exit = cmd.exit
			self.AddChild(cmd.exit)
			return true
		}
		return false
	}) {
		ExitCommand(self)
	}

	for {
		if self.parent == nil {
			fmt.Printf("%s>", self.Name)
		} else {
			prompt := ""
			RecurseParents(self, func(cmd *Command, first, last bool) bool {
				if first {
					prompt = fmt.Sprintf("%s)>", cmd.Name)
					return false
				}
				if last {
					prompt = fmt.Sprintf("%s(%s", cmd.Name, prompt)
					return false
				}
				prompt = fmt.Sprintf("%s/%s", cmd.Name, prompt)
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

		args := parseArgs(input)
		selCmd := self.Find(args[0])
		if selCmd == nil {
			fmt.Printf(
				"Invalid command, type %s for available commands\n",
				self.help.Name)
			continue
		}
		if selCmd == self.exit {
			break
		}
		if selCmd.Run == nil &&
			selCmd.Load == nil &&
			len(selCmd.childs) == 0 {
			return errors.New(fmt.Sprintf(
				"Missing action for %s command", selCmd.Name))
		}
		if selCmd.Run == nil {
			err = selCmd.Execute()
			if err != nil {
				return err
			}
			continue
		}

		selCmd.Run(selCmd, args[1:])
	}

	return nil
}

// ExitCommand creates a new exit command and adds it to selected parent
// command.
func ExitCommand(parent *Command) *Command {
	cmd := &Command{
		Name: "exit",
		Help: "Exit from this program",
	}
	parent.exit = cmd
	parent.AddChild(cmd)
	return cmd
}

// Find finds a command by its name.
func (self *Command) Find(name string) *Command {
	for _, v := range self.childs {
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
		Name: "help",
		Help: "Show an overview help",
		Parameters: []Parameter{
			Parameter{
				Name:     "command",
				Help:     "Show help of specified command",
				Optional: true},
		},
		Run: DefaultHelp,
	}
	parent.help = cmd
	parent.AddChild(cmd)
	return cmd
}

// RecurseParents interates from current commands to all its parents until
// the function returns true.
func RecurseParents(cmd *Command, f RecurseGet) (successful bool) {
	if f(cmd, true, false) {
		return true
	}

	parent := cmd.parent
	if parent == nil {
		return false
	}

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

// DefaultHelp defines a default output for help command.
func DefaultHelp(cmd *Command, args []string) {
	parent := cmd.parent
	if len(args) > 1 {
		fmt.Println("The help command cannot take more than 1 parameter")
		return
	}
	if len(args) == 0 {
		maxLen := 0
		for _, v := range parent.childs {
			if len(v.Name) > maxLen {
				maxLen = len(v.Name)
			}
		}
		fmt.Println(parent.Help, "\n")
		fmt.Println("Available commands:")
		fmtStr := fmt.Sprintf("  %%-%ds  %%s\n", maxLen)
		for _, v := range parent.childs {
			fmt.Printf(fmtStr, v.Name, v.Help)
		}
	} else {
		selCmd := parent.Find(args[0])
		if selCmd == nil {
			fmt.Printf("The command %s cannot be found", args[0])
			return
		}

		fmt.Println(selCmd.Help)
		if len(selCmd.Parameters) == 0 {
			return
		}
		maxLen := 0
		for _, v := range selCmd.Parameters {
			if len(v.Name) > maxLen {
				maxLen = len(v.Name)
			}
		}
		fmt.Println("\nAvailable parameters:")
		fmtStr := fmt.Sprintf("  %%-%ds  %%s\n", maxLen)
		for _, v := range selCmd.Parameters {
			fmt.Printf(fmtStr, v.Name, v.Help)
		}
	}
}
