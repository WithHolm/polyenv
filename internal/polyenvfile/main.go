// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package polyenvfile contains all functions related to the polyenv file
package polyenvfile

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/vaults"
)

// struct to hold opened ployenv file
type File struct {
	//full path to the file
	Path string `toml:"-"`
	//name of the environment
	Name     string                    `toml:"-"`
	Options  VaultOptions              `toml:"options"`
	VaultMap map[string]map[string]any `toml:"vault"`
	Vaults   map[string]model.Vault    `toml:"-"`
	Secrets  map[string]model.Secret   `toml:"secret"`
}

func ValidateEnvName(name string) error {
	NotAllowed := []string{".", "..", ".git", "polyenv", "env", ".env.secret"}
	for _, n := range NotAllowed {
		if strings.Contains(name, n) {
			return fmt.Errorf("name cannot be %s", n)
		}
	}
	return nil
}

// //region file

// region open file
// open vault from env tag
func OpenFile(env string) (File, error) {
	e := ValidateEnvName(env)
	if e != nil {
		return File{}, e
	}
	slog.Debug("opening polyenv file", "env", env)
	root, e := tools.GetGitRootOrCwd()
	if e != nil {
		return File{}, e
	}

	allfiles, e := tools.GetAllFiles(root, []string{env + ".polyenv.toml"}, tools.MatchNameIExact)
	if e != nil {
		return File{}, e
	}

	var path string
	for _, f := range allfiles {
		if strings.HasPrefix(filepath.Base(f), env) {
			path = f
			break
		}
	}
	if path == "" {
		return File{}, fmt.Errorf("no env file found with name '%s'", env)
	}
	slog.Debug("found polyenv file", "path", path)

	var vaultFile File

	// decode the file to struct
	meta, err := toml.DecodeFile(path, &vaultFile)
	if err != nil {
		return vaultFile, fmt.Errorf("failed to read vault options file: %s", err)
	}

	if len(meta.Undecoded()) > 0 {
		slog.Warn("got undecoded items in vault file", "undecoded", meta.Undecoded())
	}

	vaultFile.Path = filepath.Dir(path)
	vaultFile.Name, _, _ = strings.Cut(filepath.Base(path), ".")

	//convert map of toml vaults to map of vaultmodel.Vault
	vaultFile.Vaults = make(map[string]model.Vault)
	if len(vaultFile.VaultMap) > 0 {
		slog.Debug("file has vaults", "length", len(vaultFile.VaultMap))
	}
	for k, v := range vaultFile.VaultMap {
		slog.Debug("processing configured vault", "key", k)

		// slog.Debug("vault", "options", v)
		t, ok := v["type"]
		if !ok {
			return vaultFile, fmt.Errorf("vault '%s': key 'type' is missing in polyenv file", k)
		}

		vaultType := fmt.Sprintf("%s", t)
		if vaultType == "" {
			return vaultFile, fmt.Errorf("vault '%s': vault 'type' is missing in .polyenv file", k)
		}

		vault, err := vaults.NewVaultInstance(string(vaultType))
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error getting instance of vault '%s': %w", k, vaultType, err)
		}
		err = vault.Unmarshal(v)
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error unmarshalling config: %w", k, err)
		}
		vaultFile.Vaults[k] = vault
	}

	if len(vaultFile.Secrets) > 0 {
		slog.Debug("file has secrets", "length", len(vaultFile.Secrets))
	}
	for k, v := range vaultFile.Secrets {
		slog.Debug("secret reference", "key", k, "vault", v.Vault)
		v.LocalKey = k
		vaultFile.Secrets[k] = v
	}

	return vaultFile, nil
}

// region save file

// Save polyenv file struct to disk
func (file *File) Save() {
	file.VaultMap = make(map[string]map[string]any)
	for k, v := range file.Vaults {
		slog.Debug("marshalling vault", "displayname", k, "vault", v.String())

		maps := v.Marshal()

		_, ok := maps["type"]
		if !ok {
			slog.Error("failed to marshal vault", "vault", v, "error", "missing type")
		}

		file.VaultMap[k] = maps
	}

	bytes, e := toml.Marshal(file)
	if e != nil {
		slog.Error("failed to marshal polyenv file", "error", e)
		os.Exit(1)
	}
	// filepath := filepath.Join(file.Path, file.Name+".polyenv.toml")
	slog.Debug("saving polyenvfile", "path", file.Path, "name", file.Name)
	err := os.WriteFile(filepath.Join(file.Path, file.Name+".polyenv.toml"), bytes, 0644)
	if err != nil {
		slog.Error("failed to write polyenv file", "error", err)
		os.Exit(1)
	}
}

// func (file *File) String() string {
// 	if file.Name == "" {
// 		return "<none>"
// 	}
// 	return file.Name
// }

// returns the full path to the polyenv file
func (file *File) Fullname() string {
	return filepath.Join(file.Path, file.Name+".polyenv.toml")
}

// Get the vault in file by name
func (file *File) GetVault(name string) (model.Vault, error) {
	vault, ok := file.Vaults[name]
	if !ok {
		return nil, fmt.Errorf("vault not found")
	}
	return vault, nil
}

// list all vaults in file
func (file *File) GetVaultNames() []string {
	out := make([]string, 0)
	for k := range file.Vaults {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// validates secret name.
func (file *File) ValidateSecretName(name string) error {
	name = file.Options.ConvertString(name)
	for k := range file.Secrets {
		if k == name {
			return fmt.Errorf("secret name already exists: %s", name)
		}
	}
	return nil
}

func (file *File) GetSecretInfo(remoteKey string, vault string) (model.Secret, bool) {
	for _, v := range file.Secrets {
		// slog.Info("secret", "remotekey", v.RemoteKey, "vault", v.Vault)
		if v.RemoteKey == remoteKey && v.Vault == vault {
			return v, true
		}
	}
	return model.Secret{}, false
}

//region vaultopts

// return all dotenv keys in the project in files that include current environment
// {env}.env || .env.{env} || .env.secret.{env} || {env}.env.secret
func (file *File) AllDotenvValues() (out []model.StoredEnv, err error) {
	configEnv := file.Name
	slog.Debug("getting all dotenv values within", "env", configEnv)
	// get all files
	cwd, err := tools.GetGitRootOrCwd()
	if err != nil {
		return nil, err
	}
	filter := []string{configEnv + ".env", ".env." + configEnv, ".env.secret." + configEnv}
	if configEnv == "" {
		filter = []string{".env", ".env.secret"}
	}

	allfiles, err := tools.GetAllFiles(cwd, filter, tools.MatchNameIExact)
	if err != nil {
		return nil, err
	}

	// sort by amount of /
	sort.Slice(allfiles, func(i, j int) bool {
		return strings.Count(allfiles[i], "/") > strings.Count(allfiles[j], "/")
	})

	slog.Debug("filtering dotenv for", "env", configEnv)

	for _, fl := range allfiles {
		fileEnv, err := tools.ExtractNameFromDotenv(filepath.Base(fl))
		if err != nil {
			slog.Error("failed to extract name from dotenv", "error", err)
			os.Exit(1)
		}
		if fileEnv != configEnv {
			slog.Debug("skipping dotenv", "file", fl, "detected file env", fileEnv, "config env", configEnv)
			continue
		}
		slog.Debug("found dotenv", "file", fl)

		m, e := godotenv.Read(fl)
		if e != nil {
			slog.Error("failed to read file", "error", e)
			os.Exit(1)
		}

		for k, v := range m {
			slog.Debug("dotenv value", "key", k)
			isSecret := false
			for _, s := range file.Secrets {
				if s.LocalKey == k {
					isSecret = true
					break
				}
			}
			out = append(out, model.StoredEnv{
				Key:      k,
				Value:    v,
				File:     fl,
				IsSecret: isSecret,
			})
		}
	}

	return out, nil
}

// generate a new file name based on the current file name
func (file *File) GenerateFileName(extension string) string {
	// extension = strings.TrimPrefix(extension, ".")
	extension = strings.TrimSuffix(extension, ".")

	return extension + "." + file.Name
}
