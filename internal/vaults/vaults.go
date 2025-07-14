package vaults

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
	"github.com/withholm/polyenv/internal/tools"

	_ "embed"
)

type Vault interface {
	// returns the display name of the vault
	DisplayName() string
	// push a single secret
	Push(name string, value string) error
	// pull all secrets
	Pull() (map[string]string, error)
	// list all secrets
	List() ([]string, error)
	// remove secret from remote vault
	Flush(key []string) error
	// tries to un-delte or un-fuck the last push/flush
	Opsie() error
	// Warm the vault connection.
	Warmup() error
	// validate the incoming config
	ValidateConfig(options map[string]string) error
	// set options for the vault
	SetOptions(map[string]string) error
	// get options for the vault
	// GetOptions() map[string]string

	WizardWarmup(map[string]string) error
	WizardNext() *huh.Form
	WizardComplete() map[string]string
}

type VaultOptions struct {
	ReplaceHyphen     bool     `toml:"replaceHyphen"`
	AutoUppercase     bool     `toml:"autoUppercase"`
	IgnoreContentType []string `toml:"ignoreContentType"`
}

func VaildateVaultOpts(opts map[string]string) error {
	if opts["VAULT_TYPE"] == "" {
		return fmt.Errorf("vault type cannot be empty")
	}

	vlt, ok := Registry[opts["VAULT_TYPE"]]
	if !ok {
		return fmt.Errorf("unknown vault type: %s", opts["VAULT_TYPE"])
	}

	vlt.ValidateConfig(opts)

	return nil
}

// open vault from path
func OpenVault(path string) (Vault, error) {
	opts, err := tools.GetVaultFile(path)
	if err != nil {
		return nil, err
	}

	v, err := NewInitVault(opts["VAULT_TYPE"])
	if err != nil {
		return nil, err
	}

	err = v.SetOptions(opts)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func NewInitVault(vaultType string) (Vault, error) {
	if vaultType == "" {
		return nil, fmt.Errorf("vault type cannot be empty")
	}

	vault, ok := Registry[vaultType]
	if !ok {
		return nil, fmt.Errorf("unknown vault type: %s", vaultType)
	}
	return vault, nil
}

func NewVault(vaultType string, options map[string]string) (Vault, error) {
	if vaultType == "" {
		return nil, fmt.Errorf("vault type cannot be empty")
	}

	v, err := NewInitVault(vaultType)
	if err != nil {
		return nil, err
	}
	v.SetOptions(options)
	err = v.Warmup()
	if err != nil {
		return nil, err
	}
	return v, nil
}

//go:embed template
var template string

// write vault file
func WriteFile(path string, options map[string]string) error {
	path = GetVaultPath(path)

	if options["VAULT_TYPE"] == "" {
		slog.Debug("please, developer, add 'VAULT_TYPE' as output to GetOptions()")
		return fmt.Errorf("vault type cannot be empty")
	}

	out := make([]string, 0)
	out = append(out, template)
	s, err := godotenv.Marshal(options)
	if err != nil {
		return err
	}
	out = append(out, s)

	//str to byte
	out = append(out, "\n")
	//0644:rw-r--r--
	err = os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
	if err != nil {
		panic("failed to write file: " + err.Error())
	}

	return nil
}
