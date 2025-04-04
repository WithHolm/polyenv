package tools

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func GetVaultOptsPath(envfile string) string {
	if strings.HasSuffix(envfile, ".polyenv") {
		return envfile
	}

	return envfile + ".polyenv"
}

func TestVaultFileExists(envfile string) error {
	// vaultFile := GetVaultOptsPath(envfile)
	file, err := os.Stat(envfile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("env file does not exist")
	}
	if err != nil {
		return fmt.Errorf("failed to find env file: %s", err)
	}
	slog.Debug("found env file", "file", envfile, "size", file.Size(), "file", file)
	return nil
	// return true
}

func VaultOptsExist(envfile string) bool {
	vaultFile := GetVaultOptsPath(envfile)
	_, err := os.Stat(vaultFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// open read vault options from .env.vaultopts
func GetVaultFile(envfile string) (map[string]string, error) {
	vaultFile := GetVaultOptsPath(envfile)

	_, err := os.Stat(vaultFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no vault options file found")
	}

	ret, err := godotenv.Read(vaultFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault options file: %s", err)
	}
	return ret, nil
}

// get vault options from .env.vaultopts
func InitVaultFile(envfile string, opts map[string]string) error {
	// vaultFile := GetVaultOptsPath(envfile)
	// _, err := os.Stat(vaultFile)
	// if err != nil && errors.Is(err, os.ErrNotExist) {

	// 	_, err := vaults.NewVault(opts["VAULT_TYPE"], opts["VAULT_NAME"], opts)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to initialize vault: %s", err)
	// 	}

	// 	makeVaultFile(vaultFile, opts)
	// }
	return nil
}

// func makeVaultFile(path string, options map[string]string) {
// 	path = GetVaultOptsPath(path)
// 	log.Println("making vault options file")
// 	template := make([]string, 0)
// 	template = append(template, "# this file is automatically generated by polyenv")
// 	template = append(template, "# do not edit this file")
// 	for k, v := range options {
// 		template = append(template, fmt.Sprintf("%s = %s", k, v))
// 	}

// 	//str to byte
// 	template = append(template, "\n")
// 	out := []byte(strings.Join(template, "\n"))
// 	err := os.WriteFile(path, out, 0644)
// 	if err != nil {
// 		panic("failed to write file: " + err.Error())
// 	}
// }
