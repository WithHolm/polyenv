package polyenvfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/devvault"
)

func TestOpenFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	content := `
[options]
hyphens_to_underscores = true
uppercase_locally = true

[vault.test-vault]
type = "devvault"
store = "mystore"

[secret.MYKEY]
remote_key = "mykey"
vault = "test-vault"
`

	filePath := filepath.Join(tmpDir, "dev.polyenv.toml")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// To isolate the test, we can temporarily change the current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	f, err := OpenFile("dev")
	if err != nil {
		t.Fatalf("OpenFile() returned an error: %v", err)
	}

	if f.Name != "dev" {
		t.Errorf("expected name to be 'dev', but got '%s'", f.Name)
	}

	if !f.Options.HyphenToUnderscore {
		t.Error("expected HyphenToUnderscore to be true, but it was false")
	}

	if !f.Options.UppercaseLocally {
		t.Error("expected UppercaseLocally to be true, but it was false")
	}

	if len(f.Vaults) != 1 {
		t.Fatalf("expected 1 vault, but got %d", len(f.Vaults))
	}

	if _, ok := f.Vaults["test-vault"]; !ok {
		t.Error("expected to find vault 'test-vault', but it was not there")
	}

	if len(f.Secrets) != 1 {
		t.Fatalf("expected 1 secret, but got %d", len(f.Secrets))
	}

	if _, ok := f.Secrets["MYKEY"]; !ok {
		t.Error("expected to find secret 'MYKEY', but it was not there")
	}
}

func TestFile_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	file := File{
		Path: tmpDir,
		Name: "dev",
		Options: VaultOptions{
			HyphenToUnderscore: true,
			UppercaseLocally:   true,
		},
		Vaults: map[string]model.Vault{
			"test-vault": &devvault.Client{},
		},
		Secrets: map[string]model.Secret{
			"MYKEY": {
				RemoteKey: "mykey",
				Vault:     "test-vault",
			},
		},
	}

	file.Save()

	filePath := filepath.Join(tmpDir, "dev.polyenv.toml")
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatalf("Save() did not create the file at %s", filePath)
	}
}

func TestFile_GetVault(t *testing.T) {
	file := File{
		Vaults: map[string]model.Vault{
			"test-vault": &devvault.Client{},
		},
	}

	vault, err := file.GetVault("test-vault")
	if err != nil {
		t.Fatalf("GetVault() returned an error: %v", err)
	}

	if vault.DisplayName() != "Development Vault" {
		t.Errorf("expected vault display name to be 'Development Vault', but got '%s'", vault.DisplayName())
	}

	_, err = file.GetVault("non-existent-vault")
	if err == nil {
		t.Error("GetVault() should have returned an error for a non-existent vault, but it didn't")
	}
}

func TestFile_GetVaultNames(t *testing.T) {
	file := File{
		Vaults: map[string]model.Vault{
			"vault1": &devvault.Client{},
			"vault2": &devvault.Client{},
		},
	}

	names := file.GetVaultNames()
	expectedNames := []string{"vault1", "vault2"}

	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("expected vault names %v, but got %v", expectedNames, names)
	}
}

func TestFile_ValidateSecretName(t *testing.T) {
	file := File{
		Options: VaultOptions{
			HyphenToUnderscore: true,
			UppercaseLocally:   true,
		},
		Secrets: map[string]model.Secret{
			"EXISTING_SECRET": {},
		},
	}

	err := file.ValidateSecretName("new-secret")
	if err != nil {
		t.Errorf("ValidateSecretName() returned an error for a valid name: %v", err)
	}

	err = file.ValidateSecretName("existing-secret")
	if err == nil {
		t.Error("ValidateSecretName() should have returned an error for an existing name, but it didn't")
	}
}

func TestFile_GetSecretInfo(t *testing.T) {
	file := File{
		Secrets: map[string]model.Secret{
			"MYKEY": {
				RemoteKey: "mykey",
				Vault:     "test-vault",
			},
		},
	}

	secret, found := file.GetSecretInfo("mykey", "test-vault")
	if !found {
		t.Fatal("GetSecretInfo() did not find the secret")
	}

	if secret.RemoteKey != "mykey" {
		t.Errorf("expected remote key to be 'mykey', but got '%s'", secret.RemoteKey)
	}

	_, found = file.GetSecretInfo("non-existent-secret", "test-vault")
	if found {
		t.Error("GetSecretInfo() found a non-existent secret")
	}
}
