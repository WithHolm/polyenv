package model

import "fmt"

type Secret struct {
	Vault       string `toml:"vault"`
	ContentType string `toml:"content_type"`
	Enabled     bool   `toml:"-"`
	RemoteKey   string `toml:"remote_key"`
	LocalKey    string `toml:"-"`
}

// Used when pushing or pulling secrets
type SecretContent struct {
	ContentType string
	Value       string
	RemoteKey   string
	LocalKey    string
}

func (s Secret) ToString() string { return fmt.Sprintf("%s (%s)", s.RemoteKey, s.ContentType) }

// Gets the secret content from the vault
func (s Secret) GetContent(v Vault) (string, error) {
	err := v.PullElevate()
	if err != nil {
		return "", fmt.Errorf("failed to elevate permissions: %s", err)
	}

	ret, er := v.Pull(s)
	if er != nil {
		return "", fmt.Errorf("failed to pull secret: %s", er)
	}
	return ret.Value, nil
}

// Pushes the secret content to the vault
func (s Secret) SetContent(v Vault, content SecretContent) error {
	err := v.PushElevate()
	if err != nil {
		return fmt.Errorf("failed to elevate permissions: %s", err)
	}

	er := v.Push(content)
	if er != nil {
		return fmt.Errorf("failed to push secret: %s", er)
	}
	return nil
}
