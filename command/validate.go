package command

import (
	"context"
	"fmt"

	"vauban/config"

	"github.com/urfave/cli/v3"
)

func Validate() *Command {
	return New().
		WithName("validate").
		WithUsage("Validate the vauban configuration file").
		WithAction(func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadAndValidate(config.DefaultPath)
			if err != nil {
				return err
			}

			fmt.Printf("Configuration is valid: %d job(s) loaded from %s\n", len(cfg.Jobs), config.DefaultPath)
			return nil
		})
}
