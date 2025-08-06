package vaults

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/keyvault"
)

// registry
var reg = map[string]func() model.Vault{
	"keyvault": func() model.Vault { return &keyvault.Client{} },
}
var regMu sync.RWMutex
var logged = false

// returns a new instance of the vault
func NewVaultInstance(vaultType string) (model.Vault, error) {
	regMu.RLock()
	defer regMu.RUnlock()
	v, ok := reg[vaultType]
	if !ok {
		return nil, fmt.Errorf("unknown vault type: %s", vaultType)
	}
	return v(), nil
}

// returns a list of all vault types taken from the registry
func List() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	if !logged {
		slog.Debug("registered vaults", "count", len(reg))
		logged = true
	}

	keys := make([]string, 0)
	for k := range reg {
		keys = append(keys, k)
	}
	return keys
}
