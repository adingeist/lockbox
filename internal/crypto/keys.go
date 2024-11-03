package crypto

import (
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
	PrivateKey string // Only set for personal keys
}

type KeyManager struct {
	globalDir string // ~/.lockbox
	localDir  string // ./.lockbox
}

func NewKeyManager() (*KeyManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	globalDir := filepath.Join(homeDir, ".lockbox")
	if err := os.MkdirAll(globalDir, 0700); err != nil {
	return nil, fmt.Errorf("failed to create global key directory: %w", err)
	}

	return &KeyManager{
		globalDir: globalDir,
	}, nil
}

func (km *KeyManager) SetLocalDir(dir string) {
	km.localDir = dir
}

// Personal key management
func (km *KeyManager) GenerateKeyPair(name string) (*Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	id := &Identity{
		Name:       name,
		PublicKey:  identity.Recipient().String(),
		PrivateKey: identity.String(),
	}

	if err := km.savePersonalKey(id); err != nil {
		return nil, err
	}

	return id, nil
}

func (km *KeyManager) savePersonalKey(identity *Identity) error {
	keysDir := filepath.Join(km.globalDir, "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return fmt.Errorf("failed to create keys directory: %w", err)
	}

	data, err := json.Marshal(identity)
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}

	keyPath := filepath.Join(keysDir, fmt.Sprintf("%s.json", identity.Name))
	return os.WriteFile(keyPath, data, 0600)
}

func (km *KeyManager) getPersonalKey(name string) (*Identity, error) {
	keyPath := filepath.Join(km.globalDir, "keys", fmt.Sprintf("%s.json", name))
	data, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	var identity Identity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity: %w", err)
	}

	return &identity, nil
}

func (km *KeyManager) ListPersonalKeys() ([]Identity, error) {
	keysDir := filepath.Join(km.globalDir, "keys")
	entries, err := os.ReadDir(keysDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read keys directory: %w", err)
	}

	var identities []Identity
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			data, err := os.ReadFile(filepath.Clean(filepath.Join(keysDir, entry.Name())))
			if err != nil {
				continue
			}
			var id Identity
			if err := json.Unmarshal(data, &id); err != nil {
				continue
			}
			identities = append(identities, id)
		}
	}

	return identities, nil
}

func (km *KeyManager) RemovePersonalKey(name string) error {
	keyPath := filepath.Join(km.globalDir, "keys", fmt.Sprintf("%s.json", name))
	return os.Remove(keyPath)
}

// Team key management
func (km *KeyManager) SaveTeamKey(identity *Identity) error {
	if km.localDir == "" {
		return fmt.Errorf("no local directory set")
	}

	keysFile := filepath.Join(km.localDir, "team-keys.txt")
	f, err := os.OpenFile(keysFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open team keys file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "# %s\n%s\n", identity.Name, identity.PublicKey)
	return err
}

func (km *KeyManager) ListTeamKeys() ([]Identity, error) {
	if km.localDir == "" {
		return nil, fmt.Errorf("no local directory set")
	}

	keysFile := filepath.Join(km.localDir, "team-keys.txt")
	data, err := os.ReadFile(keysFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read team keys file: %w", err)
	}

	var identities []Identity
	var currentName string

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else {
			identities = append(identities, Identity{
				Name:      currentName,
				PublicKey: line,
			})
		}
	}

	return identities, nil
}

// File encryption/decryption
func (km *KeyManager) EncryptFile(inputPath string, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	encrypted, err := km.Encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, encrypted, 0644)
}

func (km *KeyManager) DecryptFile(inputPath string, outputPath string, keyName string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	identity, err := km.getPersonalKey(keyName)
	if err != nil {
		return err
	}

	decrypted, err := km.Decrypt(data, identity.PrivateKey)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, decrypted, 0644)
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
	keyFile := filepath.Join(km.localDir, "private.key")
	data, err := json.Marshal(identity)
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}
	return os.WriteFile(keyFile, data, 0600)
}

// LoadPrivateKey loads the user's private key
func (km *KeyManager) LoadPrivateKey() (*Identity, error) {
	keyFile := filepath.Join(km.localDir, "private.key")
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

func (km *KeyManager) RemoveTeamKey(publicKey string) error {
	if km.localDir == "" {
		return fmt.Errorf("no local directory set")
	}

	keysFile := filepath.Join(km.localDir, "team-keys.txt")
	data, err := os.ReadFile(keysFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read team keys file: %w", err)
	}

	var lines []string
	var skipNext bool

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if !skipNext {
				lines = append(lines, line)
			}
			skipNext = false
		} else {
			if line == publicKey {
				skipNext = true
				continue
			}
			if !skipNext {
				lines = append(lines, line)
			}
		}
	}

	if len(lines) == 0 {
		return os.Remove(keysFile)
	}

	var buf bytes.Buffer
	for _, line := range lines {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	return os.WriteFile(keysFile, buf.Bytes(), 0644)
} 