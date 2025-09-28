// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
//trigger
