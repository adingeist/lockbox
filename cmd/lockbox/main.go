package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/git"
	"github.com/yourusername/lockbox/internal/version"
	"github.com/yourusername/lockbox/internal/commands/team"
	"github.com/yourusername/lockbox/internal/commands/key"
	"github.com/yourusername/lockbox/internal/commands/secret"
)

func main() {
	app := &cli.App{
		Name:    "lockbox",
		Usage:   "Secure team secret management",
		Version: version.Version,
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize lockbox in the current git repository",
				Action: func(c *cli.Context) error {
					gitRoot, err := git.FindRoot()
					if err != nil {
						return err
					}

					lockboxDir := filepath.Join(gitRoot, ".lockbox")
					if err := os.MkdirAll(lockboxDir, 0755); err != nil {
						return fmt.Errorf("failed to create lockbox directory: %w", err)
					}

					fmt.Printf("Created Lockbox directory at '%s'\n", lockboxDir)
					return nil
				},
			},
			key.Command(),
			team.Command(),
			secret.Command(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
	}
} 