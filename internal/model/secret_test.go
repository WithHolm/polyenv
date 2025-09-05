package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/huh"
)

func TestDetectSecret(t *testing.T) {
	testCases := []struct {
		name         string
		env          StoredEnv
		expectSecret bool
		expectReason string
	}{
		{
			name:         "Not a secret",
			env:          StoredEnv{Key: "USERNAME", Value: "testuser", File: ""},
			expectSecret: false,
			expectReason: "",
		},
		{
			name:         "Empty Value",
			env:          StoredEnv{Key: "SOME_KEY", Value: "", File: ""},
			expectSecret: false,
			expectReason: "",
		},
		{
			name:         "Keyword match in key",
			env:          StoredEnv{Key: "DB_PASSWORD", Value: "12345", File: ""},
			expectSecret: true,
			expectReason: "Keyword 'PASSWORD' in key",
		},
		{
			name:         "GitHub Token by Regex",
			env:          StoredEnv{Key: "GHA_TOKEN", Value: "ghp_abcdefghijklmnopqrstuvwxyz1234567890", File: ""},
			expectSecret: true,
			expectReason: "GitHub Token",
		},
		{
			name:         "Slack Token by Regex",
			env:          StoredEnv{Key: "BotToken", Value: "xoxp-12345-67890-12345-67890-12345-abcdef", File: ""},
			expectSecret: true,
			expectReason: "Slack Token",
		},
		{
			name:         "High Entropy Value",
			env:          StoredEnv{Key: "SESSION_ID", Value: "z9s1vFpQLmN8DsB5fVbTjR2cW3aY4xZ6", File: ""},
			expectSecret: true,
			expectReason: "High entropy value",
		},
		{
			name:         "PEM Key by Regex",
			env:          StoredEnv{Key: "CERT", Value: "-----BEGIN RSA PRIVATE KEY-----", File: ""},
			expectSecret: true,
			expectReason: "PEM private key",
		},
		{
			name:         "Generic Base64",
			env:          StoredEnv{Key: "API", Value: "12345678901234567890123456789012", File: ""},
			expectSecret: true,
			expectReason: "Generic Base64 (40+)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isSecret, reason := tc.env.DetectSecret()

			if isSecret != tc.expectSecret {
				t.Errorf("Expected isSecret to be %v, but got %v", tc.expectSecret, isSecret)
			}

			if reason != tc.expectReason {
				t.Errorf("Expected reason to be '%s', but got '%s'", tc.expectReason, reason)
			}
		})
	}
}

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
