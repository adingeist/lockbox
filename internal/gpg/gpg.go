package gpg

import (
	"fmt"
	"github.com/proglottis/gpgme"
)

type GPG struct {
	ctx     *gpgme.Context
	homeDir string
}

type Key struct {
	KeyID       string
	Fingerprint string
	UIDs        []string
}

// New creates a new GPG instance
func New(homeDir string) (*GPG, error) {
	ctx, err := gpgme.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GPG: %w", err)
	}

	if homeDir != "" {
		if err := ctx.SetHomeDir(homeDir); err != nil {
			return nil, fmt.Errorf("failed to set GPG home directory: %w", err)
		}
	}

	return &GPG{
		ctx:     ctx,
		homeDir: homeDir,
	}, nil
}

// ListKeys returns all public keys in the keyring
func (g *GPG) ListKeys(secret bool) ([]Key, error) {
	keys, err := g.ctx.GetKeyList("", secret)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	var result []Key
	for _, k := range keys {
		key := Key{
			KeyID:       k.SubKeys[0].KeyID,
			Fingerprint: k.SubKeys[0].Fingerprint,
			UIDs:        make([]string, 0, len(k.UserIDs)),
		}
		for _, uid := range k.UserIDs {
			key.UIDs = append(key.UIDs, uid.UID)
		}
		result = append(result, key)
	}

	return result, nil
}

// ImportKey imports a public key
func (g *GPG) ImportKey(keyData string) error {
	data, err := gpgme.NewDataBytes([]byte(keyData))
	if err != nil {
		return fmt.Errorf("failed to create data buffer: %w", err)
	}
	defer data.Close()

	_, err = g.ctx.Import(data)
	if err != nil {
		return fmt.Errorf("failed to import key: %w", err)
	}

	return nil
}

// ExportKey exports a public key by ID
func (g *GPG) ExportKey(keyID string) (string, error) {
	data, err := gpgme.NewData()
	if err != nil {
		return "", fmt.Errorf("failed to create data buffer: %w", err)
	}
	defer data.Close()

	if err := g.ctx.Export(keyID, 0, data); err != nil {
		return "", fmt.Errorf("failed to export key: %w", err)
	}

	data.Seek(0, 0)
	exported, err := data.Read()
	if err != nil {
		return "", fmt.Errorf("failed to read exported key: %w", err)
	}

	return string(exported), nil
}

// DeleteKey removes a key from the keyring
func (g *GPG) DeleteKey(keyID string) error {
	key, err := g.ctx.GetKey(keyID)
	if err != nil {
		return fmt.Errorf("failed to find key: %w", err)
	}

	if err := g.ctx.DeleteKey(key, false); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
} 