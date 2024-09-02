package cmd

import (
	"dotenv-keyvault/internal/tools"
	"dotenv-keyvault/internal/vaults"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var pushPath string
var pushVaultName string
var pushVaultType string
var pushVaultTenant string
var pushOmitcomments bool
var pushFlush bool

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push the current .env file to keyvault",
	Long: `
		push the current .env file to keyvault. 
		Any push will also override existing keyvault secrets, however it will not delete any secrets that are not in the .env file.
		Will add a .env.vaultopts file to the current directory, so later you can push and pull from the same vault without having to specify the options again.
	`,
	Run: push,
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushPath, "path", "p", ".env", "path to the .env file to push. uses /.env by default")
	pushCmd.Flags().StringVarP(&pushVaultName, "vaultName", "v", "", "name of the keyvault to push to. only needed first time")
	pushCmd.Flags().StringVar(&pushVaultType, "vaultType", "keyvault", "type of vault. only keyvault is supported at the moment")
	pushCmd.Flags().StringVarP(&pushVaultTenant, "tenant", "t", "", "tenant for the keyvault")
	// pushCmd.Flags().BoolP("omitcomments", "c", false, "omit comments from the .env file when pushing")
	// pushCmd.Flags().BoolP("flush", "f", false, "will remove any existing keyvault secrets before pushing.")
}

func push(cmd *cobra.Command, args []string) {
	fmt.Println("push called")

	// mabye no need? good to have tbh..
	if pushPath == "" {
		panic("path cannot be empty")
	}

	// get absolute path
	if !filepath.IsAbs(pushPath) {
		// path is absolute
		_path, err := filepath.Abs(pushPath)
		if err != nil {
			panic("failed to get absolute path: " + err.Error())
		}
		pushPath = _path
	}

	// check if path exists
	_, err := os.Stat(pushPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		panic("cannot find with path " + pushPath)
	} else if err != nil {
		panic("failed to stat file: " + err.Error())
	}

	if pushVaultName != "" {
		err := tools.CheckDoubleDashS(pushVaultName, "vaultName")
		if err != nil {
			panic(err.Error())
		}
	}

	if pushVaultTenant != "" {
		err := tools.CheckDoubleDashS(pushVaultTenant, "tenant")
		if err != nil {
			panic(err.Error())
		}
	}

	if !tools.VaultOptsExist(pushPath) {
		if pushVaultName == "" {
			log.Fatal("--vaultName cannot be empty when pushing for the first time")
		}

		if pushVaultTenant == "" {
			log.Fatal("--tenant cannot be empty when pushing for the first time")
		}

		fmt.Println("no vault options file found, creating one")
		err := tools.InitVaultFile(pushPath, map[string]string{
			"VAULT_NAME":   pushVaultName,
			"VAULT_TYPE":   pushVaultType,
			"VAULT_TENANT": pushVaultTenant,
		})

		if err != nil {
			log.Fatal("failed to create vault options file: " + err.Error())
			os.Exit(1)
		}
	}

	// read opts file
	vaultOpts, err := tools.GetVaultFile(pushPath)
	if err != nil {
		log.Fatal("failed to get vault options: " + err.Error())
		os.Exit(1)
	}
	cli, err := vaults.NewVault(vaultOpts["VAULT_TYPE"], vaultOpts["VAULT_NAME"], vaultOpts)

	// push dotenv to vault
	dotenvm, err := godotenv.Read(pushPath)
	if err != nil {
		panic("failed to read .env file: " + err.Error())
	}
	for k, v := range dotenvm {
		fmt.Println("pushing '" + k + "' to vault")
		err := cli.Push(k, v)
		if err != nil {
			panic("failed to push to vault: " + err.Error())
		}
	}
}
