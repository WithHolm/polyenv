//go:build !omitdevpackage

// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package devvault contains a development vault that can be used for testing
package devvault

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
)

var vaultName = "devvault"

var stores = []Store{
	{
		Name: "mystore",
		Keys: map[string]Key{
			"mykey": {
				ContentType: "text/plain",
				Value:       "myvalue",
				Enabled:     true,
			},
		},
	},
	{
		Name: "myOtherstore",
		Keys: map[string]Key{
			"mykey": {
				ContentType: "text/plain",
				Value:       "myothervalue",
				Enabled:     true,
			},
			"myOtherkey": {
				ContentType: "text/plain",
				Value:       "myothervalue",
				Enabled:     false,
			},
			"myThirdValue": {
				ContentType: "text/plain",
				Value:       "mythirdvalue",
				Enabled:     false,
			},
		},
	},
}

// const devstore []Store = make([]Store, 0)

type Store struct {
	Name string         `json:"name"`
	Keys map[string]Key `json:"keys"`
}

type Key struct {
	ContentType string `json:"content_type"`
	Value       string `json:"value"`
	Enabled     bool   `json:"enabled"`
}

type Client struct {
	wizState   int
	Name       string `toml:"name"`
	store      Store  `toml:"-"`
	PushCalled bool   `toml:"-"`
}

func getVaults() (out []Store, err error) {
	slices.SortFunc(stores, func(a, b Store) int {
		return strings.Compare(a.Name, b.Name)
	})
	// slices.Sort(stores)
	return stores, nil
}

func (c *Client) String() string {
	return "devvault/" + c.store.Name
}

func (c *Client) DisplayName() string {
	return "Development Vault"
}

func (c *Client) SecretSelectionHandler(sec *[]model.Secret) bool {
	return false
}

func (c *Client) Warmup() error {
	if c.store.Name == "" {
		for _, v := range stores {
			if v.Name == c.Name {
				c.store = v
				return nil
			}
		}
		return fmt.Errorf("cannot find vault '%s'", c.Name)
	}

	return nil
}

func (c *Client) Marshal() map[string]any {
	return map[string]any{
		"type":  vaultName,
		"store": c.store.Name,
	}
}

// create a type from the given map
func (c *Client) Unmarshal(m map[string]any) error {
	s, ok := m["store"]
	if !ok {
		return fmt.Errorf("invalid or missing 'store' key")
	}
	st := s.(string)
	slog.Debug("unmarshal", "store", st)
	var store Store
	val, e := getVaults()
	if e != nil {
		return fmt.Errorf("failed to get vaults: %w", e)
	}
	slog.Debug("devvault: found stores", "length", len(val))

	for _, v := range val {
		if v.Name == st {
			store = v
			break
		}
	}
	slog.Debug("devvault: found the correct store", "name", store.Name)
	if store.Name == "" {
		return fmt.Errorf("cannot find store for vault '%s'", st)
	}
	c.store = store
	c.Name = st
	return nil
}

func (c *Client) ValidateSecretName(name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("should not be empty")
	}
	return name, nil
}

//region Wiz

func (c *Client) WizWarmup(m map[string]any) error {
	c.wizState = 0
	return nil
}

func (c *Client) WizNext() (*huh.Form, error) {
	defer func() { c.wizState++ }()
	switch c.wizState {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a vault").
					OptionsFunc(func() (ret []huh.Option[string]) {
						vaults, err := getVaults()
						if err != nil {
							panic(fmt.Errorf("failed to get vaults: %w", err))
						}
						for _, v := range vaults {
							opt := huh.NewOption(v.Name, v.Name)
							ret = append(ret, opt)
						}

						return ret
					}, nil).Value(&c.Name),
			),
		), nil
	}
	return nil, nil
}

func (c *Client) WizComplete() error {
	return nil
}

//region Push

func (c *Client) Push(s model.SecretContent) error {

	return nil
}

func (c *Client) PushElevate() error {
	return nil
}

//region Pull

func (c *Client) Pull(s model.Secret) (model.SecretContent, error) {
	var sec model.SecretContent
	sec.RemoteKey = s.RemoteKey
	sec.LocalKey = s.LocalKey
	sec.ContentType = s.ContentType
	sec.Value = c.store.Keys[s.RemoteKey].Value
	return sec, nil
}

func (c *Client) PullElevate() error {
	return nil
}

//region List

func (c *Client) ListElevate() error {
	return nil
}

func (c *Client) List() (out []model.Secret, err error) {
	out = make([]model.Secret, 0)
	for k, v := range c.store.Keys {
		slog.Debug("adding secret", "key", k)
		out = append(out, model.Secret{
			RemoteKey:   k,
			LocalKey:    k,
			ContentType: v.ContentType,
			Enabled:     v.Enabled,
		})
	}
	return out, nil
}
