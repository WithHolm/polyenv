package polyenvfile

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

type VaultOptions struct {
	HyphenToUnderscore         bool `toml:"hyphens_to_underscores"`
	UppercaseLocally           bool `toml:"uppercase_locally"`
	UseDotSecretFileForSecrets bool `toml:"use_dot_secret_file_for_secrets"`
}

// struct to hold opened ployenv file
type File struct {
	Path     string                    `toml:"-"`
	Name     string                    `toml:"-"`
	Options  VaultOptions              `toml:"options"`
	VaultMap map[string]map[string]any `toml:"vault"`
	Vaults   map[string]model.Vault    `toml:"-"`
	Secrets  map[string]model.Secret   `toml:"secret"`
}

// open vault from path
func OpenVaultFile(path string) (File, error) {
	path = tools.GetVaultFilePath(path)

	var vaultFile File

	err := tools.TestVaultFileExists(path)
	if err != nil {
		return vaultFile, err
	}

	// decode the file to struct
	meta, err := toml.DecodeFile(path, &vaultFile)
	if err != nil {
		return vaultFile, fmt.Errorf("failed to read vault options file: %s", err)
	}

	if len(meta.Undecoded()) > 0 {
		slog.Warn("got undecoded items in vault file", "undecoded", meta.Undecoded())
	}

	vaultFile.Path = path

	//convert map of toml vaults to map of vaultmodel.Vault
	for k, v := range vaultFile.VaultMap {
		slog.Debug("got vault", "key", k)

		slog.Debug("vault", "options", v)
		vaultType := fmt.Sprintf("%s", v["type"])
		if vaultType == "" {
			return vaultFile, fmt.Errorf("vault '%s': vault 'type' is missing in .polyenv file", k)
		}

		vault, err := vaults.GetVaultInstance(string(vaultType))
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error getting instance of vault '%s'", k, vaultType)
		}

		err = vault.ValidateConfig(v)
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error validating config: %s", k, err)
		}

		vbyte, err := toml.Marshal(v)
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error marshalling vault options: %s", k, err)
		}
		err = toml.Unmarshal(vbyte, &vault)
		if err != nil {
			return vaultFile, fmt.Errorf("vault '%s': error unmarshalling vault options: %s", k, err)
		}

		vaultFile.Vaults[k] = vault
	}

	return vaultFile, nil
}

func (file *File) Save() {
	file.VaultMap = make(map[string]map[string]any)
	for k, v := range file.Vaults {
		slog.Debug("processing vault", "displayname", k, "vault", v)
		// stupid conversion to map[string]any
		// convert to toml then back to map[string]any
		bytes, err := toml.Marshal(v)
		if err != nil {
			slog.Error("failed to marshal vault", "vault", v, "error", err)
			os.Exit(1)
		}

		maps := make(map[string]any)
		err = toml.Unmarshal(bytes, &maps)
		if err != nil {
			slog.Error("failed to decode vault", "vault", v, "error", err)
			os.Exit(1)
		}

		file.VaultMap[k] = maps
	}

	bytes, e := toml.Marshal(file)
	if e != nil {
		slog.Error("failed to marshal polyenv file", "error", e)
		os.Exit(1)
	}
	slog.Debug("saving polyenvfile", "path", file.Path)
	err := os.WriteFile(file.Path, bytes, 0644)
	if err != nil {
		slog.Error("failed to write polyenv file", "error", err)
		os.Exit(1)
	}
}

func (file *File) GetVault(name string) (model.Vault, error) {
	vault, ok := file.Vaults[name]
	if !ok {
		return nil, fmt.Errorf("vault not found")
	}
	return vault, nil
}

func (file *File) TuiSelectVault() *model.Vault {
	var displayName string
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select vault").
				OptionsFunc(func() (ret []huh.Option[string]) {
					for k, v := range file.Vaults {
						ret = append(ret, huh.NewOption(fmt.Sprintf("%s (%s)", k, v.ToString()), k))
					}
					return ret
				}, nil).
				Value(&displayName),
		),
	)
	tui.RunHuh(f)
	vault, ok := file.Vaults[displayName]
	if !ok {
		slog.Error("vault not found", "vault", displayName)
		os.Exit(1)
	}
	return &vault
}

func (file *File) AddVault(displayName string, vault model.Vault) {
	file.Vaults[displayName] = vault
}

func (file *File) RemoveVault(displayName string) {
	delete(file.Vaults, displayName)
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

// converts string using rules set by options
func (opt VaultOptions) ConvertString(s string) string {
	if opt.UppercaseLocally {
		s = strings.ToUpper(s)
	}
	if opt.HyphenToUnderscore {
		s = strings.ReplaceAll(s, "-", "_")
	}
	return s
}
