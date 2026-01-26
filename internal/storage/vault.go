package storage

import (
	"fmt"
	"sync"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/storage/vault/keyvault"
	"github.com/withholm/polyenv/internal/storage/vault/localvault"
)

// var reg = map[string]func() model.Vault{
// 	"keyvault": func() model.Vault { return &keyvault.Client{} },
// 	"local":    func() model.Vault { return &localvault.Client{} },
// }

var factory *VaultFactory

type VaultFactory struct {
	regMu    sync.RWMutex
	Registry map[string]func() model.Vault
}

func init() {
	factory = NewVaultFactory()
	factory.RegisterVault("keyvault", func() model.Vault { return &keyvault.Client{} })
	factory.RegisterVault("local", func() model.Vault { return &localvault.Client{} })
}

// creates a new vault factory (singleton)
func NewVaultFactory() *VaultFactory {
	if factory != nil {
		return factory
	}
	return &VaultFactory{
		regMu:    sync.RWMutex{},
		Registry: make(map[string]func() model.Vault),
	}
}

// registers a new vault
func (r *VaultFactory) RegisterVault(key string, init func() model.Vault) error {
	r.regMu.Lock()
	defer r.regMu.Unlock()

	if _, ok := r.Registry[key]; ok {
		return fmt.Errorf("vault already registered: %s", key)
	}

	r.Registry[key] = init
	return nil
}

// opens a new instance of vault by name
func (r *VaultFactory) GetVault(key string) (model.Vault, error) {
	r.regMu.RLock()
	defer r.regMu.RUnlock()

	init, ok := r.Registry[key]
	if !ok {
		return nil, fmt.Errorf("vault not found: %s", key)
	}

	return init(), nil
}

// returns a list of all vault names
func (r *VaultFactory) GetVaultNames() []string {
	r.regMu.RLock()
	defer r.regMu.RUnlock()

	out := make([]string, 0)
	for k := range r.Registry {
		out = append(out, k)
	}
	return out
}
