package vaults

import "fmt"

type Vault interface {
	Push(name string, value string) error // push a single secret
	Pull() (map[string]string, error)     // pull all secrets
	Flush(key string) error               // flush a single secret
	FlushAll() error                      // flush all secrets
	Opsie() error                         // tries to un-delte
	Init() error                          // initialize the vault
	// Extension() string                    // returns the extension of the vault
}

func NewVault(vaultType string, vaultName string, options map[string]string) (Vault, error) {

	if vaultType == "" {
		return nil, fmt.Errorf("vault type cannot be empty")
	}

	if vaultType == "keyvault" {
		if options["VAULT_TENANT"] == "" {
			return nil, fmt.Errorf("--tenant cannot be empty. you can either use the domian for your tenant og GUID")
		}

		cli := KeyvaultClient{
			name:   vaultName,
			tenant: options["VAULT_TENANT"],
		}
		err := cli.Init()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize keyvault client: %s", err)
		}
		return &cli, nil
	}

	return nil, fmt.Errorf("unknown vault type: %s", vaultType)
}
