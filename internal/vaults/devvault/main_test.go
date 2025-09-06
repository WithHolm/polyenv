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
