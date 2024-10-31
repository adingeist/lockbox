package team

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/git"
	"github.com/yourusername/lockbox/internal/crypto"
	"github.com/yourusername/lockbox/internal/prompt"
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
			showKeyCommand(),
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

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

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
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

			// Ask if adding from personal keys or file
			source, err := prompt.SelectFromList(
				"Add key from",
				[]string{"Personal keys", "Public key file"},
			)
			if err != nil {
				return err
			}

			var identity *crypto.Identity
			if source == "Personal keys" {
				identities, err := km.ListPersonalKeys()
				if err != nil {
					return err
				}

				if len(identities) == 0 {
					return fmt.Errorf("no personal keys found. Create one with 'lockbox key add'")
				}

				var options []string
				idMap := make(map[string]crypto.Identity)
				for _, id := range identities {
					options = append(options, fmt.Sprintf("%s (%s)", id.Name, id.PublicKey))
					idMap[options[len(options)-1]] = id
				}

				selected, err := prompt.SelectFromList("Select a key to add", options)
				if err != nil {
					return err
				}

				id := idMap[selected]
				identity = &id
			} else {
				filePath, err := prompt.Input("Enter path to public key file")
				if err != nil {
					return err
				}

				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read key file: %w", err)
				}

				name, err := prompt.Input("Enter name for this key")
				if err != nil {
					return err
				}

				identity = &crypto.Identity{
					Name:      name,
					PublicKey: string(data),
				}
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
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

			// List team members
			identities, err := km.ListTeamKeys()
			if err != nil {
				return err
			}

			if len(identities) == 0 {
				return fmt.Errorf("no team members found")
			}

			// Create options for selection
			var options []string
			idMap := make(map[string]crypto.Identity)
			for _, id := range identities {
				display := fmt.Sprintf("%s (%s)", id.Name, id.PublicKey)
				options = append(options, display)
				idMap[display] = id
			}

			// Select member to remove
			selected, err := prompt.SelectFromList("Select team member to remove", options)
			if err != nil {
				return err
			}

			identity := idMap[selected]

			// Confirm removal
			confirmed, err := prompt.Confirm(fmt.Sprintf("Are you sure you want to remove %s from the team?", identity.Name))
			if err != nil {
				return err
			}

			if !confirmed {
				fmt.Println("Operation cancelled")
				return nil
			}

			if err := km.RemoveTeamKey(identity.PublicKey); err != nil {
				return err
			}

			fmt.Printf("Successfully removed %s from the team\n", identity.Name)
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

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

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

func showKeyCommand() *cli.Command {
	return &cli.Command{
		Name:  "show-key",
		Usage: "Show your public key to share with others",
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

			identity, err := km.LoadPrivateKey()
			if err != nil {
				return err
			}
			if identity == nil {
				return fmt.Errorf("no identity found. Run 'lockbox team init' first")
			}

			fmt.Printf("Your public key:\n%s\n", identity.PublicKey)
			return nil
		},
	}
}