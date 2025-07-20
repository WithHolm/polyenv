package tools

import (
	"fmt"
	"log/slog"
	"os"
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
// func AppendDotEnvExtension(path string) string {
// 	if strings.Contains(path, ".env") {
// 		return path
// 	}
// 	return ".env"path + ""
// }

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

// gets path from either Path flag or positional argument
func SetPathOrArg(Path string, args []string) string {
	if len(args) >= 1 && Path != "" {
		slog.Error("Both --path and positional arguments are set. Please use only one of the two.")
		os.Exit(1)
	} else if len(args) == 1 && Path == "" {
		slog.Debug("using positional argument as path", "path", args[0])
		Path = args[0]
	}

	if Path == "" {
		slog.Error("no path set. please set --path or positional argument")
		os.Exit(1)
	}
	return Path
}
