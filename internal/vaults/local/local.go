// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package local contains a local vault that connects to a local cred-store
package local

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tui"
	keyring "github.com/zalando/go-keyring"
)

var vaultName = "local"

type Client struct {
	Service  string `toml:"service"`
	wizKey   string `toml:"-"`
	wizState int    `toml:"-"`
}

func (c *Client) String() string {
	return vaultName
}

func (c *Client) DisplayName() string {
	return "Local Cred Store"
}

func (c *Client) SecretSelectionHandler(sec *[]model.Secret) bool {
	var key string
	var val string
	tui.RunHuh(
		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("key").
					Description("enter key of secret in local cred-store").
					Validate(func(s string) error {
						str, err := keyring.Get(c.Service, s)
						if err == nil && str != "" {
							return fmt.Errorf("key already exists")
						}
						return nil
					}).
					Value(&key),
				huh.NewInput().
					Title("value").
					Description("enter value of secret").
					EchoMode(huh.EchoModePassword).
					Value(&val),
				huh.NewNote().DescriptionFunc(func() string {
					return fmt.Sprintf("saved as: %s:%s", c.Service, key)
				}, &key),
			),
		),
	)

	err := keyring.Set(c.Service, key, val)

	if err != nil {
		slog.Error("failed to set key to local cred-store", "error", err)
		os.Exit(1)
	}
	*sec = append(*sec, model.Secret{
		RemoteKey:   key,
		LocalKey:    key,
		ContentType: "text/plain",
		Enabled:     true,
	})
	return true
}

func (c *Client) Warmup() error {
	slog.Debug("'local' warmup: adding test-key to local cred-store")
	err := keyring.Set("polyenv", "warmupKey", "warmupVal")
	if err != nil {
		return fmt.Errorf("failed to set data to local cred-store during warmup: %w", err)
	}
	slog.Debug("'local' warmup: reading test-key from local cred-store")
	val, err := keyring.Get("polyenv", "warmupKey")
	if err != nil {
		return fmt.Errorf("failed to read data from local cred-store during warmup: %w", err)
	}
	if val != "warmupVal" {
		return fmt.Errorf("failed to get same data as we put in")
	}
	slog.Debug("'local' warmup: removing test-key from local cred-store")
	err = keyring.Delete("polyenv", "warmupKey")
	if err != nil {
		return fmt.Errorf("failed to remove data from local cred-store during warmup: %w", err)
	}
	return nil
}

func (c *Client) Marshal() map[string]any {
	return map[string]any{
		"type":    vaultName,
		"service": c.Service,
	}
}

func (c *Client) Unmarshal(m map[string]any) error {
	if service, ok := m["service"].(string); ok {
		c.Service = service
	}
	return nil
}

// region Wiz
func (c *Client) WizWarmup(m map[string]any) error {
	err := c.Warmup()
	if err != nil {
		return err
	}

	if m["service"] != nil {
		c.Service = m["service"].(string)
	}
	if m["s"] != nil {
		c.Service = m["s"].(string)
	}

	if m["key"] != nil {
		c.wizKey = m["key"].(string)
	}
	if m["k"] != nil {
		c.wizKey = m["k"].(string)
	}

	c.wizState = 0
	return nil
}

func (c *Client) WizNext() (*huh.Form, error) {
	defer func() { c.wizState++ }()
	if c.Service != "" {
		return nil, nil
	}

	switch c.wizState {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("service").
					Description("give the service a name. \n its acts as a sort of 'namespace' so your several services can use same 'key' without collision").
					Value(&c.Service),
			),
		), nil
	}
	return nil, nil
}

func (c *Client) WizComplete() error {
	return nil
}

//endregion

// region Push
func (c *Client) Push(s model.SecretContent) error {

	val, err := keyring.Get(c.Service, s.RemoteKey)
	if err == nil && val == s.Value {
		//if value is the same, no need to push
		return nil
	} else if err != keyring.ErrNotFound {
		//if error from get is anything but not found, return err
		slog.Error("failed to get data from local cred-store during push", "error", err)
		return err
	}

	slog.Debug("adding/updating secret in local cred-store. will be added", "service", c.Service, "key", s.RemoteKey)

	err = keyring.Set(c.Service, s.RemoteKey, s.Value)
	if err != nil {
		return fmt.Errorf("failed to set data to local cred-store during push: %w", err)
	}
	return fmt.Errorf("push is not supported for manual vaults")
}

func (c *Client) PushElevate() error {
	return nil
}

//endregion

// region Pull
func (c *Client) Pull(s model.Secret) (model.SecretContent, error) {
	var sec model.SecretContent
	sec.RemoteKey = s.RemoteKey
	sec.LocalKey = s.LocalKey
	sec.ContentType = s.ContentType

	val, err := keyring.Get(c.Service, s.RemoteKey)
	if err == keyring.ErrNotFound {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Secret not in local cred-store. Enter value for %s", s.LocalKey)).
					EchoMode(huh.EchoModePassword).
					Value(&val),
			),
		)
		tui.RunHuh(form)
		err = keyring.Set(c.Service, s.RemoteKey, val)
		if err != nil {
			return sec, fmt.Errorf("failed to set data to local cred-store during pull: %w", err)
		}

	} else if err != nil {
		return sec, err
	}
	sec.Value = val
	return sec, nil
}

func (c *Client) PullElevate() error {
	return nil
}

//endregion

// region List
func (c *Client) ListElevate() error {
	return nil
}

func (c *Client) List() ([]model.Secret, error) {
	return []model.Secret{}, nil
}

//endregion
