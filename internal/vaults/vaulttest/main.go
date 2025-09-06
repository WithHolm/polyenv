package vaulttest

import (
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

// TestVault is a generic test suite for the model.Vault interface.
// It should be called by each implementation of the Vault interface.
func TestVault(t *testing.T, v model.Vault, newVault func() model.Vault) {
	t.Run("DisplayName", func(t *testing.T) {
		if v.DisplayName() == "" {
			t.Error("DisplayName() should not be empty")
		}
	})

	t.Run("List", func(t *testing.T) {
		secrets, err := v.List()
		if err != nil {
			t.Fatalf("List() returned an error: %v", err)
		}
		if len(secrets) == 0 {
			t.Log("warning: List() returned no secrets, some tests will be skipped")
		}
	})

	t.Run("Pull", func(t *testing.T) {
		secrets, err := v.List()
		if err != nil {
			t.Fatalf("List() returned an error: %v", err)
		}
		if len(secrets) == 0 {
			t.Skip("skipping Pull test because List() returned no secrets")
		}

		secret := secrets[0]
		content, err := v.Pull(secret)
		if err != nil {
			t.Fatalf("Pull() returned an error: %v", err)
		}
		if content.Value == "" {
			t.Error("Pull() returned an empty value")
		}
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		m := v.Marshal()
		if m == nil {
			t.Fatal("Marshal() returned nil")
		}

		t.Log("Marshal() returned the following map:")
		t.Log(m)

		newV := newVault()
		err := newV.Unmarshal(m)
		if err != nil {
			t.Fatalf("Unmarshal() returned an error: %v", err)
		}
	})
}
