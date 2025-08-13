//go:build !omitdevpackage

package vaults

import (
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/devvault"
)

func init() {
	regMu.RLock()
	defer regMu.RUnlock()
	reg["devvault"] = func() model.Vault { return &devvault.Client{} }
}

// func main() {
// 	RegisterVault(func() model.Vault { return &devvault.Client{} }, "devvault")
// }
