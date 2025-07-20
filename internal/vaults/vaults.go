package vaults

import (
	_ "embed"
)

// func NewInitVault(vaultType string) (vaultmodel.Vault, error) {
// 	if vaultType == "" {
// 		return nil, fmt.Errorf("vault type cannot be empty")
// 	}

// 	vault, ok := Registry[vaultType]
// 	if !ok {
// 		return nil, fmt.Errorf("unknown vault type: %s", vaultType)
// 	}
// 	return vault, nil
// }

// func NewVault(vaultType string, options map[string]string) (vaultmodel.Vault, error) {
// 	if vaultType == "" {
// 		return nil, fmt.Errorf("vault type cannot be empty")
// 	}

// 	v, err := NewInitVault(vaultType)
// 	if err != nil {
// 		return nil, err
// 	}
// 	v.SetOptions(options)
// 	err = v.Warmup()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return v, nil
// }

// //go:embed template
// var template string

// // write vault file
// func WriteFile(path string, options map[string]string) error {
// 	path = GetVaultPath(path)

// 	if options["VAULT_TYPE"] == "" {
// 		slog.Debug("please, developer, add 'VAULT_TYPE' as output to GetOptions()")
// 		return fmt.Errorf("vault type cannot be empty")
// 	}

// 	out := make([]string, 0)
// 	out = append(out, template)
// 	s, err := godotenv.Marshal(options)
// 	if err != nil {
// 		return err
// 	}
// 	out = append(out, s)

// 	//str to byte
// 	out = append(out, "\n")
// 	//0644:rw-r--r--
// 	err = os.WriteFile(path, []byte(strings.Join(out, "\n")), 0644)
// 	if err != nil {
// 		panic("failed to write file: " + err.Error())
// 	}

// 	return nil
// }
