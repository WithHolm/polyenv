// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !omitdevpackage

package storage

import (
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/storage/vault/devvault"
)

// register devvault
func init() {
	NewVaultFactory().RegisterVault(
		"devvault",
		func() model.Vault { return &devvault.Client{} },
	)
}
