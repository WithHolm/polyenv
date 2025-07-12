package vaults

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/vaults/keyvault"
)

// TODO for dev!
type VaultType string

// append your valuttype
const (
	vltTypeKeyvault VaultType = "keyvault"
)

// append vault type to registry. this will be
var Registry = map[string]Vault{
	"keyvault": &keyvault.Client{},
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
	for k := range Registry {
		keys = append(keys, k)
	}
	return keys
}

// used by cobra. sets the vault type from string
func (vtype *VaultType) Set(value string) error {
	_, ok := Registry[value]
	if !ok {
		return fmt.Errorf("unknown vault type: %s. must be one of %s", value, ListVaultTypes())
	}
	*vtype = VaultType(value)
	return nil
}

// used by init. lists all vault types as huh select.
func VaultTypeSelector(ref *VaultType) *huh.Select[VaultType] {
	opt := make([]huh.Option[VaultType], 0)
	for k, v := range Registry {
		opt = append(opt, huh.NewOption(v.DisplayName(), VaultType(k)))
	}

	return huh.NewSelect[VaultType]().
		Title("Select a vault type").
		Options(opt...).
		Value(ref)
}
