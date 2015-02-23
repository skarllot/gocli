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

type Command struct {
	Name   string
	Short  string
	Long   string
	Run    func(cmd *Command, args []string)
	childs []*Command
	parent *Command
	exit   *Command
	help   *Command
}

func (c *Command) AddChild(cmds ...*Command) {
	for _, v := range cmds {
		if v == c {
			panic("Command can't be a child of itself")
		}

		v.parent = c
		c.childs = append(c.childs, v)
	}
}

func (c *Command) Execute() error {
	if c.exit == nil {
		return errors.New("Cannot execute a command without an exit command")
	}
	if c.Run == nil && len(c.childs) == 0 {
		return errors.New("Current command has no action and childs either")
	}
	if c.Run != nil && len(c.childs) != 0 {
		return errors.New("Current command should not define an action and has childs")
	}

	for {
		if c.parent == nil {
			fmt.Printf("%s>", c.Name)
		} else {
			prompt := fmt.Sprintf("%s)>", c.Name)
			parent := c.parent
			for {
				if parent.parent == nil {
					prompt = fmt.Sprintf("%s(%s", parent.Name, prompt)
					break
				}
				prompt = fmt.Sprintf("%s/%s", parent.Name, prompt)
				parent = parent.parent
			}
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
		if selCmd.Run == nil && len(selCmd.childs) == 0 {
			fmt.Printf("Missing action for %s command\n", selCmd.Name)
			break
		}
		if selCmd.Run == nil && len(args) > 1 {
			fmt.Println("Invalid arguments")
			continue
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

func ExitCommand(parent *Command) *Command {
	cmd := &Command{
		Name:  "exit",
		Short: "Exit from this program",
	}
	parent.exit = cmd
	parent.AddChild(cmd)
	return cmd
}

func (c *Command) Find(name string) *Command {
	for _, v := range c.childs {
		if v.Name == name {
			return v
		}
	}

	return nil
}

func HelpCommand(parent *Command) *Command {
	cmd := &Command{
		Name:  "help",
		Short: "help about this program",
	}
	parent.help = cmd
	parent.AddChild(cmd)
	return cmd
}

func (c *Command) ShowHelp() {
	for _, v := range c.childs {
		fmt.Printf("%s\t\t%s\n", v.Name, v.Short)
	}
}
