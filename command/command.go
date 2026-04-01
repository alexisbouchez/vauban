package command

import "github.com/urfave/cli/v3"

type Command struct {
	*cli.Command
}

func New() *Command {
	return &Command{
		Command: &cli.Command{},
	}
}

func (c *Command) WithName(name string) *Command {
	c.Name = name
	return c
}

func (c *Command) WithUsage(usage string) *Command {
	c.Usage = usage
	return c
}

func (c *Command) WithAction(action cli.ActionFunc) *Command {
	c.Action = action
	return c
}

func (c *Command) WithCommands(commands ...*Command) *Command {
	c.Commands = make([]*cli.Command, 0, len(commands))
	for _, command := range commands {
		if command == nil {
			continue
		}
		c.Commands = append(c.Commands, command.Command)
	}
	return c
}
