//go:build !omitdevpackage

package devvault

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
)

//go:embed dev.json
var devstoreFile []byte

var vaultName = "devvault"

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
	wizState int
	Name     string `toml:"name"`
	store    Store  `toml:"-"`
}

var devStore []Store

func getVaults() (out []Store, err error) {
	if len(devStore) == 0 {
		e := json.Unmarshal(devstoreFile, &devStore)
		if e != nil {
			return nil, e
		}
	}
	return devStore, nil
}

func (c *Client) ToString() string {
	return "devvault/" + c.store.Name
}

func (c *Client) DisplayName() string {
	return "Development Vault"
}

func (c *Client) Warmup() error {
	return nil
}

func (c *Client) Marshal() map[string]any {
	return map[string]any{
		"type":  vaultName,
		"store": c.store.Name,
	}
}

func (c *Client) Unmarshal(m map[string]any) error {
	s, ok := m["store"]
	if !ok {
		return fmt.Errorf("invalid or missing 'store' key")
	}
	st := s.(string)

	var store Store
	val, e := getVaults()
	if e != nil {
		return fmt.Errorf("failed to get vaults: %w", e)
	}

	for _, v := range val {
		if v.Name == st {
			store = v
			break
		}
	}

	if store.Name == "" {
		return fmt.Errorf("cannot find vault '%s'", st)
	}
	c.store = store
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
					}, nil).Value(&c.store.Name),
			),
		), nil
	}
	return nil, nil
}

func (c *Client) WizComplete() (map[string]any, error) {
	return nil, nil
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
		out = append(out, model.Secret{
			RemoteKey:   k,
			LocalKey:    k,
			ContentType: v.ContentType,
			Enabled:     true,
		})
	}
	return out, nil
}
