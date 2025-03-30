package vaults

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/vaults/keyvault"

	_ "embed"
)

type Vault interface {
	DisplayName() string                  // returns the display name of the vault
	Push(name string, value string) error // push a single secret
	Pull() (map[string]string, error)     // pull all secrets
	List() ([]string, error)              // list all secrets
	Flush(key []string) error             // remove secret from remote vault
	Opsie() error                         // tries to un-delte
	Warmup() error                        // Warm the vault connection. used for pull,push,grab(soon)
	SetOptions(map[string]string) error   // set options for the vault
	GetOptions() map[string]string        // get options for the vault
	WizardWarmup() error                  // get questions for init
	// WizardForm() *huh.Group               // get questions from the vault
	WizardNext() *huh.Form             //get next question. return null when done
	WizardComplete() map[string]string //everything is done, do your cleanup and return the options for the vault
}

// DEV: ADD NEW VAULTS HERE.
// check if the vault implements the Vault interface
var _ Vault = &keyvault.KeyvaultClient{}

// registry of all vaults
var Registry = map[string]Vault{
	"keyvault": &keyvault.KeyvaultClient{},
}

// returns Registry as options for huh.Select
func GetVaultsAsOptions() []huh.Option[string] {
	ret := make([]huh.Option[string], 0)
	for k, v := range Registry {
		ret = append(ret, huh.NewOption(v.DisplayName(), k))
	}
	return ret
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

func SaveVault(vlt Vault, dotenvFile string) error {
	path := tools.GetVaultOptsPath(dotenvFile)
	slog.Debug("making vault options file", "path", dotenvFile)

	opts := vlt.GetOptions()
	if opts["VAULT_TYPE"] == "" {
		slog.Debug("please, developer, add 'VAULT_TYPE' as output to GetOptions()")
		return fmt.Errorf("vault type cannot be empty")
	}
	out := make([]string, 0)
	out = append(out, template)

	for k, v := range vlt.GetOptions() {
		out = append(out, fmt.Sprintf("%s = %s", k, v))
	}

	//str to byte
	out = append(out, "\n")
	//0644:rw-r--r--
	err := os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
	if err != nil {
		panic("failed to write file: " + err.Error())
	}
	return nil
}
