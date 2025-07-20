package vaults

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/vaults/keyvault"
)

// TODO for dev!
type VaultType string

// !DEV:append your valuttype
const (
	vltTypeKeyvault VaultType = "keyvault"
)

// !DEV: append vault type to registry. this will be the items i use when instanciating the vault
// returns the registry
func Registry() map[string]model.Vault {
	return map[string]model.Vault{
		"keyvault": &keyvault.Client{},
	}
}

// returns a new instance of the vault
func GetVaultInstance(vaultType string) (model.Vault, error) {
	vaults := Registry()
	vault, ok := vaults[vaultType]
	if !ok {
		return nil, fmt.Errorf("unknown vault type: %s", vaultType)
	}
	return vault, nil
}

// used by cobra. returns the type of the vault
func (v *VaultType) String() string {
	return string(*v)
}

// used by cobra. returns the type of the vault
func (vtype *VaultType) Type() string {
	return fmt.Sprintf("vaultType (%s)", ListVaultTypes())
}

// returns a list of all vault types taken from the registry
func ListVaultTypes() []string {
	keys := make([]string, 0)
	for k := range Registry() {
		keys = append(keys, k)
	}
	return keys
}

// used by cobra. sets the vault type from string
func (vtype *VaultType) Set(value string) error {
	_, ok := Registry()[value]
	if !ok {
		return fmt.Errorf("unknown vault type: %s. must be one of %s", value, ListVaultTypes())
	}
	*vtype = VaultType(value)
	return nil
}

// used by init. lists all vault types as huh select.
func VaultTypeSelector(ref *VaultType) *huh.Select[VaultType] {
	return huh.NewSelect[VaultType]().
		Title("Select a vault type").
		OptionsFunc(func() []huh.Option[VaultType] {
			opt := make([]huh.Option[VaultType], 0)
			for k, v := range Registry() {
				opt = append(opt, huh.NewOption(v.DisplayName(), VaultType(k)))
			}
			return opt
		}, ref).
		Value(ref)
}
