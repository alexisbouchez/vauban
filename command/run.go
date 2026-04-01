package command

import (
	"context"
	"errors"
	"fmt"
	"os"

	"vauban/config"
	"vauban/runner"

	"github.com/urfave/cli/v3"
)

func Run() *Command {
	return New().
		WithName("run").
		WithUsage("Run jobs from vauban.json").
		WithAction(func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() > 1 {
				return errors.New("usage: vauban run [job]")
			}

			cfg, err := config.LoadAndValidate(config.DefaultPath)
			if err != nil {
				return err
			}

			r := runner.New()

			if cmd.Args().Len() == 1 {
				jobName := cmd.Args().First()
				if _, err := fmt.Fprintf(os.Stdout, "running job %s\n", jobName); err != nil {
					return err
				}

				return r.RunJob(ctx, cfg, jobName, os.Stdout, os.Stderr)
			}

			jobNames, err := cfg.JobNames()
			if err != nil {
				return err
			}

			for _, jobName := range jobNames {
				if _, err := fmt.Fprintf(os.Stdout, "running job %s\n", jobName); err != nil {
					return err
				}

				if err := r.RunJob(ctx, cfg, jobName, os.Stdout, os.Stderr); err != nil {
					return err
				}
			}

			return nil
		})
}
