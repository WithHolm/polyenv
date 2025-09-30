// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tools

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestGetVaultFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "with extension",
			path:     "my/path/file.polyenv",
			expected: "my/path/file.polyenv",
		},
		{
			name:     "without extension",
			path:     "my/path/file",
			expected: "my/path/file.polyenv",
		},
		{
			name:     "empty path",
			path:     "",
			expected: ".polyenv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetVaultFilePath(tc.path)
			if actual != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, actual)
			}
		})
	}
}

func TestTestVaultFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0666); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "file exists",
			path:      existingFile,
			expectErr: false,
		},
		{
			name:      "file does not exist",
			path:      filepath.Join(tmpDir, "not-exists.txt"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := TestVaultFileExists(tc.path)
			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tc.expectErr, err)
			}
		})
	}
}

func TestGetAllFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure for testing
	files := []string{
		"README.md",
		"main.go",
		"internal/api/api.go",
		"internal/api/README.md",
		"vendor/some/package.go",
		".git/config",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(""), 0666); err != nil {
			t.Fatal(err)
		}
	}

	testCases := []struct {
		name     string
		filter   []string
		typ      Matchtype
		expected []string
	}{
		{
			name:   "contains README",
			filter: []string{"README"},
			typ:    MatchNameContains,
			expected: []string{
				filepath.Join(tmpDir, "README.md"),
				filepath.Join(tmpDir, "internal/api/README.md"),
			},
		},
		{
			name:     "iexact main.go",
			filter:   []string{"MAIN.go"},
			typ:      MatchNameIExact,
			expected: []string{filepath.Join(tmpDir, "main.go")},
		},
		{
			name:     "no matches",
			filter:   []string{"nonexistent"},
			typ:      MatchNameContains,
			expected: nil,
		},
		{
			name:   "skip vendor and .git",
			filter: []string{".go"},
			typ:    MatchNameContains,
			expected: []string{
				filepath.Join(tmpDir, "internal/api/api.go"),
				filepath.Join(tmpDir, "main.go"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear cache for each run
			globalFileCache = NewFileCache()

			actual, err := GetAllFiles(tmpDir, tc.filter, tc.typ)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			sort.Strings(actual)
			sort.Strings(tc.expected)

			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestExtractNameFromDotenv(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		expected  string
		expectErr bool
	}{
		{name: "simple env", filename: ".env", expected: "", expectErr: false},
		{name: "simple env secret", filename: ".env.secret", expected: "", expectErr: false},
		{name: "named env", filename: ".env.dev", expected: "dev", expectErr: false},
		{name: "named env secret", filename: ".env.secret.prod", expected: "prod", expectErr: false},
		{name: "prefixed env", filename: "backend.env", expected: "backend", expectErr: false},
		{name: "prefixed env secret", filename: "frontend.env.secret", expected: "frontend", expectErr: false},
		{name: "not an env file", filename: "config.yaml", expected: "", expectErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ExtractNameFromDotenv(tc.filename)
			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, but got: %v", tc.expectErr, err)
			}
			if actual != tc.expected {
				t.Errorf("Expected %s, but got %s", tc.expected, actual)
			}
		})
	}
}
