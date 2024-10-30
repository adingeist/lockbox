package team

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/git"
	"github.com/yourusername/lockbox/internal/gpg"
)

// Command returns the team command
func Command() *cli.Command {
	return &cli.Command{
		Name:  "team",
		Usage: "Manage team members' GPG keys",
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
		Usage: "Add a team member's GPG key",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "me",
				Usage: "Add your own public key",
			},
			&cli.StringFlag{
				Name:  "file",
				Usage: "Import key from file",
			},
			&cli.StringFlag{
				Name:  "id",
				Usage: "Import key by ID",
			},
			&cli.StringFlag{
				Name:  "fingerprint",
				Usage: "Import key by fingerprint",
			},
		},
		Action: func(c *cli.Context) error {
			// Only one flag should be set
			flags := []string{"me", "file", "id", "fingerprint"}
			set := 0
			for _, flag := range flags {
				if c.IsSet(flag) {
					set++
				}
			}
			if set > 1 {
				return fmt.Errorf("only one of --me, --file, --id, or --fingerprint can be used")
			}

			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			userGPG, err := gpg.New("")
			if err != nil {
				return err
			}

			lockboxGPG, err := gpg.New(filepath.Join(gitRoot, ".lockbox"))
			if err != nil {
				return err
			}

			var keyID string

			switch {
			case c.Bool("me"):
				keys, err := userGPG.ListKeys(true)
				if err != nil || len(keys) == 0 {
					return fmt.Errorf("no private keys found in your keyring")
				}
				keyID = keys[0].KeyID

			case c.String("file") != "":
				data, err := os.ReadFile(c.String("file"))
				if err != nil {
					return fmt.Errorf("failed to read key file: %w", err)
				}
				return lockboxGPG.ImportKey(string(data))

			case c.String("id") != "":
				keyID = c.String("id")

			case c.String("fingerprint") != "":
				keyID = c.String("fingerprint")

			default:
				// Interactive mode
				keys, err := userGPG.ListKeys(false)
				if err != nil || len(keys) == 0 {
					return fmt.Errorf("no public keys found in your keyring")
				}

				fmt.Print("Enter name or email to search (or press Enter to list all): ")
				reader := bufio.NewReader(os.Stdin)
				search, _ := reader.ReadString('\n')
				search = strings.TrimSpace(search)

				var matching []gpg.Key
				for _, key := range keys {
					if search == "" {
						matching = append(matching, key)
					} else {
						for _, uid := range key.UIDs {
							if strings.Contains(strings.ToLower(uid), strings.ToLower(search)) {
								matching = append(matching, key)
								break
							}
						}
					}
				}

				if len(matching) == 0 {
					return fmt.Errorf("no matching keys found")
				}

				if len(matching) == 1 {
					keyID = matching[0].KeyID
				} else {
					fmt.Println("\nMatching keys:")
					for i, key := range matching {
						uid := "No UID"
						if len(key.UIDs) > 0 {
							uid = key.UIDs[0]
						}
						fmt.Printf("%d. %s (%s)\n", i+1, uid, key.KeyID)
					}

					var choice int
					fmt.Print("\nEnter the number of the key to add: ")
					fmt.Scanf("%d", &choice)
					if choice < 1 || choice > len(matching) {
						return fmt.Errorf("invalid choice")
					}
					keyID = matching[choice-1].KeyID
				}
			}

			// Export and import the selected key
			if c.String("file") == "" {
				keyData, err := userGPG.ExportKey(keyID)
				if err != nil {
					return fmt.Errorf("failed to export key: %w", err)
				}

				if err := lockboxGPG.ImportKey(keyData); err != nil {
					return fmt.Errorf("failed to import key: %w", err)
				}

				fmt.Printf("Successfully added key %s\n", keyID)
			}

			return nil
		},
	}
}

func removeCommand() *cli.Command {
	return &cli.Command{
		Name:  "remove",
		Usage: "Remove a team member's GPG key",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "me",
				Usage: "Remove your own public key",
			},
			&cli.StringFlag{
				Name:  "id",
				Usage: "Remove key by ID",
			},
			&cli.StringFlag{
				Name:  "fingerprint",
				Usage: "Remove key by fingerprint",
			},
		},
		Action: func(c *cli.Context) error {
			// Only one flag should be set
			flags := []string{"me", "id", "fingerprint"}
			set := 0
			for _, flag := range flags {
				if c.IsSet(flag) {
					set++
				}
			}
			if set > 1 {
				return fmt.Errorf("only one of --me, --id, or --fingerprint can be used")
			}

			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			lockboxGPG, err := gpg.New(filepath.Join(gitRoot, ".lockbox"))
			if err != nil {
				return err
			}

			var keyID string

			switch {
			case c.Bool("me"):
				userGPG, err := gpg.New("")
				if err != nil {
					return err
				}

				userKeys, err := userGPG.ListKeys(false)
				if err != nil || len(userKeys) == 0 {
					return fmt.Errorf("no public keys found in your keyring")
				}

				teamKeys, err := lockboxGPG.ListKeys(false)
				if err != nil {
					return fmt.Errorf("failed to list team keys")
				}

				var matching []gpg.Key
				for _, teamKey := range teamKeys {
					for _, userKey := range userKeys {
						if teamKey.Fingerprint == userKey.Fingerprint {
							matching = append(matching, teamKey)
							break
						}
					}
				}

				if len(matching) == 0 {
					return fmt.Errorf("none of your keys were found in the team")
				}

				if len(matching) == 1 {
					keyID = matching[0].KeyID
				} else {
					fmt.Println("Multiple keys found. Please select one to remove:")
					for i, key := range matching {
						uid := "No UID"
						if len(key.UIDs) > 0 {
							uid = key.UIDs[0]
						}
						fmt.Printf("%d. %s (%s)\n", i+1, uid, key.KeyID)
					}

					var choice int
					fmt.Print("Enter the number of the key to remove: ")
					fmt.Scanf("%d", &choice)
					if choice < 1 || choice > len(matching) {
						return fmt.Errorf("invalid choice")
					}
					keyID = matching[choice-1].KeyID
				}

			case c.String("id") != "":
				keyID = c.String("id")

			case c.String("fingerprint") != "":
				keyID = c.String("fingerprint")

			default:
				// Interactive mode
				keys, err := lockboxGPG.ListKeys(false)
				if err != nil || len(keys) == 0 {
					return fmt.Errorf("no team members found")
				}

				fmt.Println("\nSelect a team member to remove:")
				for i, key := range keys {
					uid := "No UID"
					if len(key.UIDs) > 0 {
						uid = key.UIDs[0]
					}
					fmt.Printf("%d. %s (%s)\n", i+1, uid, key.KeyID)
				}

				var choice int
				fmt.Print("\nEnter the number of the key to remove: ")
				fmt.Scanf("%d", &choice)
				if choice < 1 || choice > len(keys) {
					return fmt.Errorf("invalid choice")
				}
				keyID = keys[choice-1].KeyID
			}

			if err := lockboxGPG.DeleteKey(keyID); err != nil {
				return fmt.Errorf("failed to remove key: %w", err)
			}

			fmt.Printf("Successfully removed key %s\n", keyID)
			return nil
		},
	}
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all team members' GPG keys",
		Action: func(c *cli.Context) error {
			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			lockboxGPG, err := gpg.New(filepath.Join(gitRoot, ".lockbox"))
			if err != nil {
				return err
			}

			keys, err := lockboxGPG.ListKeys(false)
			if err != nil {
				return fmt.Errorf("failed to list keys: %w", err)
			}

			if len(keys) == 0 {
				fmt.Println("No team members found")
				return nil
			}

			fmt.Println("Team members:")
			for _, key := range keys {
				uid := "No UID"
				if len(key.UIDs) > 0 {
					uid = key.UIDs[0]
				}
				fmt.Printf("- %s: %s\n", key.KeyID, uid)
			}

			return nil
		},
	}
}