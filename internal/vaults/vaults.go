package vaults

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/vaults/keyvault"
	"github.com/withholm/polyenv/internal/vaults/temp"

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
	GetOptions() map[string]string

	NewWizardWarmup() error
	NewWizardNext() *huh.Form
	NewWizardComplete() map[string]string

	UpdateWizardWarmup(map[string]string) error
	UpdateWizardNext() *huh.Form
	UpdateWizardComplete() map[string]string
}

// registry of all vaults
var Registry = map[string]Vault{
	"keyvault": &keyvault.Client{},
	"temp":     &temp.Client{},
}

// returns Registry as options for huh.Select
func GetVaultsAsHuhOptions() []huh.Option[string] {
	ret := make([]huh.Option[string], 0)
	for k, v := range Registry {
		ret = append(ret, huh.NewOption(v.DisplayName(), k))
	}
	return ret
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

func SaveVault(vlt Vault, dotenvFile string) error {
	return WriteFile(dotenvFile, vlt.GetOptions())
}
