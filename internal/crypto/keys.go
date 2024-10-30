package crypto

import (
	"bufio"
	"bytes"
	"encoding/json"
	"filippo.io/age"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Identity struct {
	Name       string
	PublicKey  string
	PrivateKey string // Only set for the user's own key
}

type KeyManager struct {
	lockboxDir string
}

func NewKeyManager(lockboxDir string) *KeyManager {
	return &KeyManager{
		lockboxDir: lockboxDir,
	}
}

// GenerateKeyPair creates a new key pair
func (km *KeyManager) GenerateKeyPair(name string) (*Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Identity{
		Name:       name,
		PublicKey:  identity.Recipient().String(),
		PrivateKey: identity.String(),
	}, nil
}

// SaveTeamKey saves a public key to the team keyring
func (km *KeyManager) SaveTeamKey(identity *Identity) error {
	keysFile := filepath.Join(km.lockboxDir, "keys.txt")
	f, err := os.OpenFile(keysFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open keys file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "# %s\n%s\n", identity.Name, identity.PublicKey)
	return err
}

// RemoveTeamKey removes a public key from the team keyring
func (km *KeyManager) RemoveTeamKey(publicKey string) error {
	keysFile := filepath.Join(km.lockboxDir, "keys.txt")
	input, err := os.ReadFile(keysFile)
	if err != nil {
		return fmt.Errorf("failed to read keys file: %w", err)
	}

	var output []string
	scanner := bufio.NewScanner(bytes.NewReader(input))
	skip := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			if skip {
				skip = false
				continue
			}
			output = append(output, line)
		} else {
			if line == publicKey {
				skip = true
				continue
			}
			if !skip {
				output = append(output, line)
			}
		}
	}

	return os.WriteFile(keysFile, []byte(strings.Join(output, "\n")+"\n"), 0644)
}

// ListTeamKeys returns all public keys in the team keyring
func (km *KeyManager) ListTeamKeys() ([]Identity, error) {
	keysFile := filepath.Join(km.lockboxDir, "keys.txt")
	f, err := os.Open(keysFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open keys file: %w", err)
	}
	defer f.Close()

	var identities []Identity
	var currentName string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else if line != "" {
			identities = append(identities, Identity{
				Name:      currentName,
				PublicKey: line,
			})
		}
	}

	return identities, nil
}

// Encrypt encrypts data for all team members
func (km *KeyManager) Encrypt(data []byte) ([]byte, error) {
	identities, err := km.ListTeamKeys()
	if err != nil {
		return nil, err
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no team members found")
	}

	var recipients []age.Recipient
	for _, identity := range identities {
		recipient, err := age.ParseX25519Recipient(identity.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("invalid public key for %s: %w", identity.Name, err)
		}
		recipients = append(recipients, recipient)
	}

	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, recipients...)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption writer: %w", err)
	}

	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize encryption: %w", err)
	}

	return buf.Bytes(), nil
}

// Decrypt decrypts data using the user's private key
func (km *KeyManager) Decrypt(data []byte, privateKey string) ([]byte, error) {
	identity, err := age.ParseX25519Identity(privateKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	r, err := age.Decrypt(bytes.NewReader(data), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return buf.Bytes(), nil
}

// SavePrivateKey saves the user's private key
func (km *KeyManager) SavePrivateKey(identity *Identity) error {
	keyFile := filepath.Join(km.lockboxDir, "private.key")
	data, err := json.Marshal(identity)
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}
	return os.WriteFile(keyFile, data, 0600)
}

// LoadPrivateKey loads the user's private key
func (km *KeyManager) LoadPrivateKey() (*Identity, error) {
	keyFile := filepath.Join(km.lockboxDir, "private.key")
	data, err := os.ReadFile(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	var identity Identity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity: %w", err)
	}

	return &identity, nil
} 