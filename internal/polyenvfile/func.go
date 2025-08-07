package polyenvfile

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/denormal/go-gitignore"
	"github.com/withholm/polyenv/internal/tools"
)

// PolyenvFileExists checks if a polyenv file exists in the current working directory
func FileExists(env string) error {
	workingDir, e := tools.GetGitRootOrCwd()
	if e != nil {
		return e
	}
	slog.Debug("working", "dir", workingDir)
	Files, err := tools.GetAllFiles(workingDir, []string{"polyenv.toml"})
	if err != nil {
		return err
	}
	slog.Debug("checking", "env", env, "files", Files)
	for _, f := range Files {
		filename := filepath.Base(f)
		slog.Debug("checking", "file", filename)
		if strings.HasPrefix(filename, env) {
			return fmt.Errorf("file '%s' already exists: %s", env, f)
		}
	}
	return nil
}

// Lists all environments. returns string slice of names of environments
func ListEnvironments() ([]string, error) {
	workingDir, e := tools.GetGitRootOrCwd()
	if e != nil {
		return nil, e
	}

	Files, err := tools.GetAllFiles(workingDir, []string{"polyenv.toml"})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, f := range Files {
		filename := filepath.Base(f)
		if strings.HasSuffix(filename, ".polyenv.toml") {
			out = append(out, strings.TrimSuffix(filename, ".polyenv.toml"))
		}
	}
	return out, nil
}

// checks if the .gitignore file matches the .env.secure file
func GitignoreMatchesEnvSecret(skipPath ...string) bool {
	skipPath = append(skipPath, []string{
		".git",
	}...)
	slog.Debug("check if gitignore matches .env.secret files")
	root, err := tools.GetGitRootOrCwd()
	if err != nil {
		return false
	}

	// gignore, err := gitignore.NewFromFile(filepath.Join(root, ".gitignore"))

	ig, err := gitignore.NewRepository(root)
	if err != nil {
		slog.Error("failed to parse .gitignore", "error", err)
		os.Exit(1)
	}

	// ignoreParent := []string{}
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		if path != root {
			// slog.Debug("checking", "dir", path)
			if ig.Absolute(path, d.IsDir()) != nil {
				return nil
			}
		}

		f := filepath.Join(path, ".env.secret.test")
		// slog.Debug("checking", "file", f)
		if ig.Absolute(f, false) != nil {
			return nil
		}

		return fmt.Errorf("gitignore does not ignore .env.secret files in %s", path)
	})

	if e != nil {
		slog.Debug("er when procssing gitignore", "err", e)
		return false
	}

	return true
}
