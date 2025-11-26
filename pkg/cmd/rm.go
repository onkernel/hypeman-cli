package cmd

import (
	"context"
	"fmt"

	"github.com/onkernel/hypeman-go"
	"github.com/onkernel/hypeman-go/option"
	"github.com/urfave/cli/v3"
)

var rmCmd = cli.Command{
	Name:      "rm",
	Usage:     "Remove one or more instances",
	ArgsUsage: "<instance> [instance...]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "Force removal of running instances",
		},
	},
	Action:          handleRm,
	HideHelpCommand: true,
}

func handleRm(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf("instance ID required\nUsage: hypeman rm [flags] <instance> [instance...]")
	}

	force := cmd.Bool("force")
	client := hypeman.NewClient(getDefaultRequestOptions(cmd)...)

	var lastErr error
	for _, identifier := range args {
		// Resolve instance by ID, partial ID, or name
		instanceID, err := ResolveInstance(ctx, &client, identifier)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			lastErr = err
			continue
		}

		// Check instance state if not forcing
		if !force {
			inst, err := client.Instances.Get(
				ctx,
				instanceID,
				option.WithMiddleware(debugMiddleware(cmd.Root().Bool("debug"))),
			)
			if err != nil {
				fmt.Printf("Error: failed to get instance %s: %v\n", instanceID, err)
				lastErr = err
				continue
			}

			if inst.State == "Running" {
				fmt.Printf("Error: cannot remove running instance %s. Stop it first or use --force\n", instanceID)
				lastErr = fmt.Errorf("instance is running")
				continue
			}
		}

		// Delete the instance
		err = client.Instances.Delete(
			ctx,
			instanceID,
			option.WithMiddleware(debugMiddleware(cmd.Root().Bool("debug"))),
		)
		if err != nil {
			fmt.Printf("Error: failed to remove instance %s: %v\n", instanceID, err)
			lastErr = err
			continue
		}

		fmt.Println(instanceID)
	}

	return lastErr
}

