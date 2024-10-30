package team

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/git"
	"github.com/yourusername/lockbox/internal/crypto"
)

// Command returns the team command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "team",
		Usage: "Manage team members' keys",
		Subcommands: []*cli.Command{
			addCommand(),
			removeCommand(),
			listCommand(),
			initCommand(),
		},
	}
}

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize your identity for this repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Your name",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km := crypto.NewKeyManager(filepath.Join(gitRoot, ".lockbox"))

			// Check if user already has a key
			if identity, err := km.LoadPrivateKey(); err != nil {
				return err
			} else if identity != nil {
				return fmt.Errorf("you already have a key pair initialized")
			}

			// Generate new key pair
			identity, err := km.GenerateKeyPair(c.String("name"))
			if err != nil {
				return err
			}

			// Save private key
			if err := km.SavePrivateKey(identity); err != nil {
				return err
			}

			// Add public key to team
			if err := km.SaveTeamKey(identity); err != nil {
				return err
			}

			fmt.Printf("Initialized identity for %s\n", identity.Name)
			return nil
		},
	}
}

func addCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a team member's public key",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Team member's name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "key",
				Usage:    "Public key string",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km := crypto.NewKeyManager(filepath.Join(gitRoot, ".lockbox"))

			identity := &crypto.Identity{
				Name:      c.String("name"),
				PublicKey: c.String("key"),
			}

			if err := km.SaveTeamKey(identity); err != nil {
				return err
			}

			fmt.Printf("Added %s to the team\n", identity.Name)
			return nil
		},
	}
}

func removeCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove a team member",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Usage:    "Public key to remove",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km := crypto.NewKeyManager(filepath.Join(gitRoot, ".lockbox"))

			if err := km.RemoveTeamKey(c.String("key")); err != nil {
				return err
			}

			fmt.Println("Team member removed")
			return nil
		},
	}
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all team members",
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km := crypto.NewKeyManager(filepath.Join(gitRoot, ".lockbox"))

			identities, err := km.ListTeamKeys()
			if err != nil {
				return err
			}

			if len(identities) == 0 {
				fmt.Println("No team members found")
				return nil
			}

			fmt.Println("Team members:")
			for _, identity := range identities {
				fmt.Printf("- %s: %s\n", identity.Name, identity.PublicKey)
			}

			return nil
		},
	}
}