package key

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/crypto"
	"github.com/yourusername/lockbox/internal/prompt"
	"github.com/yourusername/lockbox/internal/output"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "key",
		Usage: "Manage your personal keys",
		Subcommands: []*cli.Command{
			addCommand(),
			removeCommand(),
			listCommand(),
		},
	}
}

func addCommand() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Create a new personal key",
		Action: func(c *cli.Context) error {
			name, err := prompt.Input("Enter a name for the key")
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}

			identity, err := km.GenerateKeyPair(name)
			if err != nil {
				return err
			}

			output.Successf("Created new key for %s", identity.Name)
			output.Infof("Public key: %s", identity.PublicKey)
			return nil
		},
	}
}

func removeCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove a personal key",
		Action: func(c *cli.Context) error {
			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}

			identities, err := km.ListPersonalKeys()
			if err != nil {
				return err
			}

			if len(identities) == 0 {
				return fmt.Errorf("no keys found")
			}

			var options []string
			for _, id := range identities {
				options = append(options, id.Name)
			}

			selected, err := prompt.SelectFromList("Select a key to remove", options)
			if err != nil {
				return err
			}

			confirmed, err := prompt.Confirm(fmt.Sprintf("Are you sure you want to remove key '%s'?", selected))
			if err != nil {
				return err
			}

			if !confirmed {
				fmt.Println("Operation cancelled")
				return nil
			}

			if err := km.RemovePersonalKey(selected); err != nil {
				return err
			}

			fmt.Printf("Removed key %s\n", selected)
			return nil
		},
	}
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all personal keys",
		Action: func(c *cli.Context) error {
			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}

			identities, err := km.ListPersonalKeys()
			if err != nil {
				return err
			}

			if len(identities) == 0 {
				output.Infof("No personal keys found")
				return nil
			}

			output.Section("Your keys")
			for _, identity := range identities {
				output.ListItem(fmt.Sprintf("%s: %s", identity.Name, identity.PublicKey))
			}

			return nil
		},
	}
} 