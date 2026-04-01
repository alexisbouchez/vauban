package main

import (
	"context"
	"fmt"
	"os"
	"vauban/command"
)

func main() {
	cmd := command.New().
		WithName("vauban").
		WithUsage("Vauban CLI").
		WithCommands(command.Init(), command.Validate(), command.Run())

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
