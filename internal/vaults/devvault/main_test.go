// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package devvault

import (
	"testing"

	"github.com/withholm/polyenv/internal/model"
	atestVault "github.com/withholm/polyenv/internal/vaults/atestvault"
)

func TestDevVault(t *testing.T) {
	var mystore Store
	for _, s := range stores {
		if s.Name == "mystore" {
			mystore = s
			break
		}
	}
	if mystore.Name == "" {
		t.Fatal("store 'mystore' not found for test")
	}
	atestVault.TestVault(t, &Client{store: mystore, Name: "mystore"}, func() model.Vault {
		return &Client{}
	})
}

//trigger pipeline//
