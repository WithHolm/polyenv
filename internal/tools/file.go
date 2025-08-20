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

type Matchtype int

const (
	//Name of file must contain the given string
	MatchNameContains Matchtype = iota
	//Name of file must indifferent match the given string (ignore case)
	MatchNameIExact
)

// type FileFilter struct {
// 	Value []string
// 	Type  Matchtype
// }

// func (f FileFilter) String() string {
// 	slices.Sort(f.Value)
// 	return fmt.Sprintf("%d:%s", f.Type, strings.Join(f.Value, "-"))
// }

// get all files recurcivley in the given root directory
func GetAllFiles(root string, filter []string, typ Matchtype) (out []string, err error) {
	slices.Sort(filter)
	//cache cause i know i have some files im searching for multiple times.. saves hot path calls
	key := fmt.Sprintf("%s|%d:%s", root, typ, strings.Join(filter, "-"))
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
					slog.Debug("search: appending contains-match", "path", path, "name", d.Name(), "filter", f)
					out = append(out, path)
				}
			case MatchNameIExact:
				if strings.EqualFold(d.Name(), f) {
					slog.Debug("search: appending iexact-match", "path", path, "name", d.Name(), "filter", f)
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

	slog.Debug("search: file cache add", "key", key, "out", out)
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
