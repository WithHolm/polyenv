// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tools

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

var (
	ErrFileNotEnvFile = errors.New("file is not a env or env-secret file")
)

// returns the path to the vault file
func GetVaultFilePath(path string) string {
	if strings.HasSuffix(path, ".polyenv") {
		return path
	}

	return path + ".polyenv"
}

func TestVaultFileExists(envfile string) error {
	// vaultFile := GetVaultOptsPath(envfile)
	file, err := os.Stat(envfile)
	if err != nil {
		return err
	}

	slog.Debug("found env file", "file", envfile, "size", file.Size(), "file", file)
	return nil
}

type FileCache struct {
	mu    sync.RWMutex
	cache map[string]fcache
}

func NewFileCache() *FileCache {
	return &FileCache{
		cache: make(map[string]fcache),
	}
}

type fcache struct {
	out []string
	err error
}

var globalFileCache = NewFileCache()

type Matchtype string

const (
	//Name of file must contain the given string
	MatchNameContains Matchtype = "contains"
	//Name of file must indifferent match the given string (ignore case)
	MatchNameIExact = "iexact"
)

// get all files recurcivley in the given root directory
func GetAllFiles(root string, filter []string, typ Matchtype) (out []string, err error) {
	slog.Debug("search", "root", root, "filter", filter, "typ", typ)
	slices.Sort(filter)
	//cache cause i know i have some files im searching for multiple times.. saves hot path calls
	key := fmt.Sprintf("%s|%s[%s]", root, typ, strings.Join(filter, "|"))
	// slog.Debug("cache", "key", key)

	globalFileCache.mu.RLock()
	cache, ok := globalFileCache.cache[key]
	globalFileCache.mu.RUnlock()
	if ok {
		slog.Debug("search: cache hit", "key", key)
		return cache.out, cache.err
	}

	// if cache was not hit, acquire a write lock to populate it.
	globalFileCache.mu.Lock()
	defer globalFileCache.mu.Unlock()

	// Check again in case another goroutine populated it while we were waiting for the lock.
	cache, ok = globalFileCache.cache[key]
	if ok {
		return cache.out, cache.err
	}

	// Walk the directory
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// slog.Debug("checking", "path", path, "filter", filter)

		if d.IsDir() {
			// Skip directories that are unlikely to contain relevant files.
			if d.Name() == ".git" || d.Name() == "vendor" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		for _, f := range filter {
			switch typ {
			case MatchNameContains:
				if strings.Contains(d.Name(), f) {
					slog.Debug("search: contains", "path", path, "filter", f)
					out = append(out, path)
				}
			case MatchNameIExact:
				if strings.EqualFold(d.Name(), f) {
					slog.Debug("search: iexact", "path", path, "filter", f)
					out = append(out, path)
				}
			}
		}
		return nil
	})

	// Populate the cache
	if err != nil {
		globalFileCache.cache[key] = fcache{
			out: nil, // Don't cache partial results on error
			err: err,
		}
		return nil, err
	}

	// slog.Debug("search: file cache add", "key", key, "out", out)
	globalFileCache.cache[key] = fcache{
		out: out,
		err: nil,
	}
	return out, nil
}

// removes .env|.env.secret|.env.{name}|.env.secret.{name} from filename
func ExtractNameFromDotenv(filename string) (string, error) {
	if !strings.Contains(filename, ".") {
		return filename, nil
	}
	if !strings.Contains(filename, ".env") && !strings.Contains(filename, ".env.secret") {
		return "", ErrFileNotEnvFile
	}
	StartsWithEnv := strings.HasPrefix(filename, ".env")
	IsEnvSecret := strings.Contains(filename, ".env.secret")

	// if its a {name}.env or {name}.env.secret
	if !StartsWithEnv {
		if IsEnvSecret {
			return strings.TrimSuffix(filename, ".env.secret"), nil
		}
		return strings.TrimSuffix(filename, ".env"), nil
	}

	// else .env|.env.{name}|.env.secret|.env.secret.{name}
	// var ret string
	ret := strings.TrimPrefix(filename, ".env")

	ret = strings.TrimPrefix(ret, ".secret")

	ret = strings.TrimPrefix(ret, ".")

	return ret, nil
}
