// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package devvault

import (
	"testing"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/vaulttest"
)

func TestDevVault(t *testing.T) {
	vaulttest.TestVault(t, &Client{store: stores[0]}, func() model.Vault {
		return &Client{}
	})
}

//trigger pipeline//
