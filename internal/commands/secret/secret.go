package secret

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"github.com/yourusername/lockbox/internal/crypto"
	"github.com/yourusername/lockbox/internal/git"
	"github.com/yourusername/lockbox/internal/prompt"
	"path/filepath"
	"strings"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "secret",
		Usage: "Manage encrypted secrets",
		Subcommands: []*cli.Command{
			encryptCommand(),
			decryptCommand(),
		},
	}
}

func encryptCommand() *cli.Command {
	return &cli.Command{
		Name:  "encrypt",
		Usage: "Encrypt a file using team members' public keys",
		Action: func(c *cli.Context) error {
			// Get file path
			inputPath, err := prompt.Input("Enter path to file to encrypt")
			if err != nil {
				return err
			}

			// Confirm output path
			defaultOutput := inputPath + ".encrypted"
			outputPath, err := prompt.Input(fmt.Sprintf("Enter output path [%s]", defaultOutput))
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = defaultOutput
			}

			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

			// Show team members who will be able to decrypt
			identities, err := km.ListTeamKeys()
			if err != nil {
				return err
			}

			fmt.Println("\nThe following team members will be able to decrypt:")
			for _, identity := range identities {
				fmt.Printf("- %s\n", identity.Name)
			}

			// Confirm encryption
			confirmed, err := prompt.Confirm("\nProceed with encryption?")
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Operation cancelled")
				return nil
			}

			if err := km.EncryptFile(inputPath, outputPath); err != nil {
				return err
			}

			fmt.Printf("Successfully encrypted %s -> %s\n", inputPath, outputPath)
			return nil
		},
	}
}

func decryptCommand() *cli.Command {
	return &cli.Command{
		Name:  "decrypt",
		Usage: "Decrypt a file using your private key",
		Action: func(c *cli.Context) error {
			// Get file path
			inputPath, err := prompt.Input("Enter path to encrypted file")
			if err != nil {
				return err
			}

			// Get output path
			defaultOutput := strings.TrimSuffix(inputPath, ".encrypted")
			if defaultOutput == inputPath {
				defaultOutput += ".decrypted"
			}
			outputPath, err := prompt.Input(fmt.Sprintf("Enter output path [%s]", defaultOutput))
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = defaultOutput
			}

			gitRoot, err := git.FindRoot()
			if err != nil {
				return err
			}

			km, err := crypto.NewKeyManager()
			if err != nil {
				return err
			}
			km.SetLocalDir(filepath.Join(gitRoot, ".lockbox"))

			// Select key to use for decryption
			identities, err := km.ListPersonalKeys()
			if err != nil {
				return err
			}

			if len(identities) == 0 {
				return fmt.Errorf("no personal keys found. Create one with 'lockbox key add'")
			}

			var options []string
			for _, id := range identities {
				options = append(options, id.Name)
			}

			selected, err := prompt.SelectFromList("Select key to decrypt with", options)
			if err != nil {
				return err
			}

			if err := km.DecryptFile(inputPath, outputPath, selected); err != nil {
				return err
			}

			fmt.Printf("Successfully decrypted %s -> %s\n", inputPath, outputPath)
			return nil
		},
	}
} 