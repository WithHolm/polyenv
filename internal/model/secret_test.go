package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/huh"
)

func TestSecret_ToString(t *testing.T) {
	s := Secret{
		RemoteKey:   "my-secret",
		ContentType: "text/plain",
	}

	expected := "my-secret (text/plain)"
	if s.ToString() != expected {
		t.Errorf("expected '%s', but got '%s'", expected, s.ToString())
	}
}

func TestStoredEnv_Save(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, ".env")

	// Create the file before writing to it
	_, err = os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	// if err := f.Close(); err != nil {
	// 	t.Fatalf("failed to close file: %v", err)
	// }

	se := StoredEnv{
		Key:   "MY_KEY",
		Value: "my_value",
		File:  filePath,
	}

	err = se.Save()
	if err != nil {
		t.Fatalf("Save() returned an error: %v", err)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expectedContent := "MY_KEY=\"my_value\"\n"
	if string(fileContent) != expectedContent {
		t.Errorf("expected file content to be '%s', but got '%s'", expectedContent, string(fileContent))
	}
}

func TestStoredEnv_Remove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, ".env")

	initialContent := "MY_KEY=my_value\nANOTHER_KEY=another_value\n"
	err = os.WriteFile(filePath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	se := StoredEnv{
		Key:  "MY_KEY",
		File: filePath,
	}

	err = se.Remove()
	if err != nil {
		t.Fatalf("Remove() returned an error: %v", err)
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expectedContent := "ANOTHER_KEY=\"another_value\"\n"
	if string(fileContent) != expectedContent {
		t.Errorf("expected file content to be '%s', but got '%s'", expectedContent, string(fileContent))
	}
}

func TestSecret_GetContent(t *testing.T) {
	v := &TestVault{
		PullValue: "my-secret-value",
	}
	s := Secret{
		RemoteKey: "my-secret",
	}

	content, err := s.GetContent(v)
	if err != nil {
		t.Fatalf("GetContent() returned an error: %v", err)
	}

	if content != "my-secret-value" {
		t.Errorf("expected content to be 'my-secret-value', but got '%s'", content)
	}
}

func TestSecret_SetContent(t *testing.T) {
	v := &TestVault{}
	s := Secret{
		RemoteKey: "my-secret",
	}
	content := SecretContent{
		Value: "my-secret-value",
	}

	err := s.SetContent(v, content)
	if err != nil {
		t.Fatalf("SetContent() returned an error: %v", err)
	}

	if !v.PushCalled {
		t.Error("Push() was not called on the vault")
	}
}

// Mock vault for testing
type TestVault struct {
	PullValue  string
	PushCalled bool
}

func (v *TestVault) String() string                                 { return "" }
func (v *TestVault) DisplayName() string                            { return "" }
func (v *TestVault) Warmup() error                                  { return nil }
func (v *TestVault) Marshal() map[string]any                        { return nil }
func (v *TestVault) Unmarshal(m map[string]any) error               { return nil }
func (v *TestVault) ValidateSecretName(name string) (string, error) { return name, nil }
func (v *TestVault) ListElevate() error                             { return nil }
func (v *TestVault) List() ([]Secret, error)                        { return nil, nil }
func (v *TestVault) PushElevate() error                             { return nil }
func (v *TestVault) Push(s SecretContent) error {
	v.PushCalled = true
	return nil
}
func (v *TestVault) SecretSelectionHandler(s *[]Secret) bool { return false }
func (v *TestVault) SupportsVaults() bool                    { return false }
func (v *TestVault) PullElevate() error                      { return nil }
func (v *TestVault) Pull(s Secret) (SecretContent, error) {
	return SecretContent{Value: v.PullValue}, nil
}
func (v *TestVault) WizWarmup(map[string]any) error { return nil }
func (v *TestVault) WizNext() (*huh.Form, error)    { return nil, nil }
func (v *TestVault) WizComplete() error             { return nil }
