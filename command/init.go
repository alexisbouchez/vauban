package command

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
)

func Init() *Command {
	return New().
		WithName("init").
		WithUsage("Initialize a Vauban project").
		WithAction(func(ctx context.Context, cmd *cli.Command) error {
			return errors.New("not implemented yet")
		})
}
