// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package polyenvfile

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/withholm/polyenv/internal/tools"
)

func TestListEnvironments(t *testing.T) {
	// tools.AppConfig().SetDebug(true)
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	//generate dummy files
	files := []string{
		"dev.polyenv.toml",
		"prod.polyenv.toml",
		"staging.polyenv.toml",
	}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), []byte(""), 0644)
		if err != nil {
			t.Fatalf("failed to create dummy file: %v", err)
		}
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

	//list all environments
	envs, err := ListEnvironments()
	if err != nil {
		t.Fatalf("ListEnvironments() returned an error: %v", err)
	}

	expectedEnvs := []string{"dev", "prod", "staging"}
	if len(envs) != len(expectedEnvs) {
		t.Errorf("expected %d environments, but got %d", len(expectedEnvs), len(envs))
	}

	for _, expectedEnv := range expectedEnvs {
		found := false
		for _, env := range envs {
			if env == expectedEnv {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected environment '%s' not found", expectedEnv)
		}
	}
}

func TestFileExists(t *testing.T) {
	tools.AppConfig().SetDebug(false)
	tmpDir, err := os.MkdirTemp("", "polyenv-test")
	//make sure the temp dir is clean before we start
	if err != nil {
		t.Fatalf("failed to create temp dir: %v\n", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	//generate 'dev' dummy file
	err = os.WriteFile(filepath.Join(tmpDir, "dev.polyenv.toml"), []byte(""), 0644)
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
	l, _ := os.Getwd()
	slog.Debug("cwd", "cwd", l)
	//check that dev file should exist
	err = FileExists("dev")
	if err == nil {
		t.Errorf("FileExists('dev') should have returned an error, but it didn't")
	}

	//check that prod file should not exist
	err = FileExists("prod")
	if err != nil {
		t.Errorf("FileExists('prod') should not have returned an error, but it did: %v", err)
	}
}
