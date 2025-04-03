package vaults

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "embed"

	"github.com/joho/godotenv"
)

const (
	constSolution  = "polyenv"
	constExtension = ".polyenv"
)

// get path to .polyenv file. will return same path if it already has .polyenv or not
func GetVaultPath(path string) string {
	if strings.HasSuffix(path, constExtension) {
		return path
	}

	// add .polyenv to path
	path = path + constExtension
	return path
}

// check if vault options item exist on path
func VaultPathExists(path string) bool {
	vaultPath := GetVaultPath(path)

	_, err := os.Stat(vaultPath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

// read the vault file. return a map of all keys
func ReadFile(path string) (map[string]string, error) {
	vaultPath := GetVaultPath(path)

	if !VaultPathExists(vaultPath) {
		return nil, fmt.Errorf("no %s file found", constSolution)
	}

	ret, err := godotenv.Read(vaultPath)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

//go:embed template
var template string

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
