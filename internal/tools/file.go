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
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("env file does not exist")
	}
	if err != nil {
		return fmt.Errorf("failed to find env file: %s", err)
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

func GetAllFiles(root string, filter []string) (out []string, err error) {
	//cache cause i know i have some files im searching for multiple times.. saves hot path calls

	slices.Sort(filter)
	key := fmt.Sprintf("%s-%s", root, strings.Join(filter, "-"))

	globalFileCache.mu.RLock()
	cache, ok := globalFileCache.cache[key]
	globalFileCache.mu.RUnlock()
	if ok {
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

		if d.IsDir() {
			// Skip directories that are unlikely to contain relevant files.
			if d.Name() == ".git" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		for _, f := range filter {
			if strings.Contains(d.Name(), f) {
				out = append(out, path)
				return nil // Found a match, move to the next file.
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

	slog.Debug("added to cache", "key", key, "out", out)
	globalFileCache.cache[key] = fcache{
		out: out,
		err: nil,
	}
	return out, nil
}

func ExtractNameFromDotenv(filename string) string {
	//.env|file.env|.env.file

	//.env.file
	diff := strings.TrimPrefix(filename, ".env.")
	if diff != filename {
		return diff
	}

	// file.env
	diff = strings.TrimSuffix(filename, ".env")
	if diff != filename {
		if strings.HasSuffix(diff, ".") {
			return strings.TrimSuffix(diff, ".")
		}
		return diff
	}
	return diff
}
