package model

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

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

type StoredEnv struct {
	Value string
	Key   string
	File  string
}

func (s Secret) ToString() string { return fmt.Sprintf("%s (%s)", s.RemoteKey, s.ContentType) }

// Gets the secret content from the vault
func (s Secret) GetContent(v Vault) (string, error) {
	err := v.PullElevate()
	if err != nil {
		return "", fmt.Errorf("failed to elevate permissions: %w", err)
	}

	ret, er := v.Pull(s)
	if er != nil {
		return "", fmt.Errorf("failed to pull secret: %w", er)
	}
	return ret.Value, nil
}

// Pushes the secret content to the vault
func (s Secret) SetContent(v Vault, content SecretContent) error {
	err := v.PushElevate()
	if err != nil {
		return fmt.Errorf("failed to elevate permissions: %w", err)
	}

	er := v.Push(content)
	if er != nil {
		return fmt.Errorf("failed to push secret: %w", er)
	}
	return nil
}

// saves stored env variable to dotenv file.
// will update if key already exists, or create new if not
func (st StoredEnv) Save() error {
	mp, e := godotenv.Read(st.File)
	if e != nil {
		return fmt.Errorf("failed to parse dotenv file: %w", e)
	}
	currentvalue, ok := mp[st.Key]
	if ok {
		if currentvalue == st.Value {
			return nil
		}
	}

	mp[st.Key] = st.Value
	slog.Debug("saving env", "key", st.Key, "file", st.File)
	e = godotenv.Write(mp, st.File)
	if e != nil {
		return fmt.Errorf("failed to write to dotenv file: %w", e)
	}
	return nil
}

// removes stored env variable from dotenv file
func (st StoredEnv) Remove() error {
	mp, e := godotenv.Read(st.File)
	if e != nil {
		if os.IsNotExist(e) {
			return nil
		}
		return fmt.Errorf("failed to parse dotenv file: %w", e)
	}

	_, ok := mp[st.Key]
	if ok {
		delete(mp, st.Key)
	}

	e = godotenv.Write(mp, st.File)
	if e != nil {
		return fmt.Errorf("failed to write to dotenv file: %w", e)
	}
	return nil
}
