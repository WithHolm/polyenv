package tools

import (
	"fmt"
	"strings"
)

// extract filename from cobra args
func ExtractFilenameArg(args []string) (s string) {
	for _, arg := range args {
		if len(strings.Split(arg, "!")[0]) > 0 {
			return strings.Split(arg, "!")[0]
		} else if !strings.Contains(arg, "!") {
			return arg
		}
	}
	return ""
}

// append .env extension to path if not already there
func AppendDotEnvExtension(path string) string {
	if strings.Contains(path, ".env") {
		return path
	}
	return path + ".env"
}

// extract vault name from cobra args
func ExtractVaultNameArg(args []string, vaults []string) (string, error) {
	for _, arg := range args {
		if strings.Contains(arg, "!") {
			arg := strings.Split(arg, "!")[1]
			arg = strings.TrimSpace(arg)
			arg = strings.ToLower(arg)
			for _, v := range vaults {
				if v == arg {
					return v, nil
				}
			}
			return "", fmt.Errorf("'%s' defined as vault (using '!'), but its not one of the available: %s", arg, vaults)
		}
	}
	return "", nil
}
