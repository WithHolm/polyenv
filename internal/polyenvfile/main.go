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

type VaultOptions struct {
	HyphenToUnderscore         bool `toml:"hyphens_to_underscores"`
	UppercaseLocally           bool `toml:"uppercase_locally"`
	UseDotSecretFileForSecrets bool `toml:"use_dot_secret_file_for_secrets"`
}

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

//region file

// open vault from env tag
func OpenFile(env string) (File, error) {
	root, e := tools.GetGitRootOrCwd()
	if e != nil {
		return File{}, e
	}

	allfiles, e := tools.GetAllFiles(root, []string{".polyenv.toml"})
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
	for k, v := range vaultFile.VaultMap {
		slog.Debug("processing configured vault", "key", k)

		slog.Debug("vault", "options", v)
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

	for k, v := range vaultFile.Secrets {
		v.LocalKey = k
		vaultFile.Secrets[k] = v
	}

	return vaultFile, nil
}

func (file *File) Save() {
	file.VaultMap = make(map[string]map[string]any)
	for k, v := range file.Vaults {
		slog.Debug("marshalling vault", "displayname", k, "vault", v.ToString())

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

// Get the vault by name
func (file *File) GetVault(name string) (model.Vault, error) {
	vault, ok := file.Vaults[name]
	if !ok {
		return nil, fmt.Errorf("vault not found")
	}
	return vault, nil
}

func (file *File) GetVaultNames() []string {
	out := make([]string, 0)
	for k := range file.Vaults {
		out = append(out, k)
	}
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

// converts string using rules set by options
func (opt VaultOptions) ConvertString(s string) string {
	if opt.UppercaseLocally && strings.ToUpper(s) != s {
		// slog.Debug("converting to uppercase", "string", s)
		s = strings.ToUpper(s)
	}
	if opt.HyphenToUnderscore && strings.Contains(s, "-") {
		// slog.Debug("converting to underscore", "string", s)
		s = strings.ReplaceAll(s, "-", "_")
	}
	return s
}

// return all dotenv keys in the project in files that include current environment
// {env}.env || .env.{env} || .env.secret.{env} || {env}.env.secret
func (f *File) AllDotenvKeys() (out []model.StoredEnv, err error) {
	// get all files
	cwd, err := tools.GetGitRootOrCwd()
	if err != nil {
		return nil, err
	}
	allfiles, err := tools.GetAllFiles(cwd, []string{".env"})
	if err != nil {
		return nil, err
	}

	// sort by amount of /
	sort.Slice(allfiles, func(i, j int) bool {
		return strings.Count(allfiles[i], "/") > strings.Count(allfiles[j], "/")
	})

	for _, fl := range allfiles {
		// {env}.env || .env.{env} || .env.secret.{env} || {env}.env.secret
		if !strings.Contains(fl, f.Name) {
			continue
		}

		slog.Debug("found file", "file", fl)

		m, e := godotenv.Read(fl)
		if e != nil {
			slog.Error("failed to read file", "error", e)
			os.Exit(1)
		}

		for k, v := range m {
			out = append(out, model.StoredEnv{
				Key:   k,
				Value: v,
				File:  fl,
			})
		}
	}

	return out, nil
}

// generate a new file name based on the current file name
func (f *File) GenerateFileName(extension string) string {
	// extension = strings.TrimPrefix(extension, ".")
	extension = strings.TrimSuffix(extension, ".")

	return extension + "." + f.Name
}
