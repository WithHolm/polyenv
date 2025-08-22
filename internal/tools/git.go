package tools

import (
	"os"
	"path/filepath"
)

var Cwd string
var CwdErr error
var CurrWd string

// GetGitRoot finds the root directory of the git repository containing the current working directory.
// It does this by walking up the directory tree from the CWD and looking for a ".git" directory or file.
// returns cwd if not in a git repository
func _getGitRootOrCwd() (string, error) {
	if Cwd != "" || CwdErr != nil {
		return Cwd, CwdErr
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return cwd, nil
		}
		dir = parentDir
	}
}

// returns cached git root or current working directory
func GetGitRootOrCwd() (string, error) {
	return _getGitRootOrCwd()
	// if Cwd != "" || CwdErr != nil {
	// 	return Cwd, CwdErr
	// }
	// Cwd, CwdErr = _getGitRootOrCwd()
	// return Cwd, CwdErr
}
