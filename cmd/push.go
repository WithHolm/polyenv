package cmd

import (
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/withholm/dotenv-myvault/internal/tools"
	"github.com/withholm/dotenv-myvault/internal/vaults"

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
	// pushCmd.Flags().StringVarP(&Path, "path", "p", ".env", "path to the .env file to push. uses /.env by default")
}

func push(cmd *cobra.Command, args []string) {
	slog.Debug("push called")

	pushPath := Path
	// mabye no need? good to have tbh
	if pushPath == "" {
		panic("path cannot be empty")
	}

	// get absolute path yes
	if !filepath.IsAbs(pushPath) {
		// path is absolute
		_path, err := filepath.Abs(pushPath)
		if err != nil {
			panic("failed to get absolute path: " + err.Error())
		}
		pushPath = _path
	}
	// check if path exists

	// check if path exists
	_, err := os.Stat(pushPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		panic("cannot find with path " + pushPath)
	} else if err != nil {
		panic("failed to stat file: " + err.Error())
	}

	if !tools.VaultOptsExist(pushPath) {
		slog.Error("Vault options file not created for env file. please run 'init' before you can re-run 'push'", "path", Path)
		os.Exit(1)
	}

	// read opts file
	vaultOpts, err := tools.GetVaultFile(pushPath)
	if err != nil {
		log.Fatal("failed to get vault options: " + err.Error())
		os.Exit(1)
	}
	cli, err := vaults.NewVault(vaultOpts["VAULT_TYPE"], vaultOpts)
	cli.Warmup()
	// push dotenv to vault
	dotenvm, err := godotenv.Read(pushPath)
	if err != nil {
		panic("failed to read .env file: " + err.Error())
	}

	secretCount := len(dotenvm)
	wg := sync.WaitGroup{}
	wg.Add(secretCount)
	for k, v := range dotenvm {
		// async push
		go func(k string, v string) {
			defer wg.Done()
			slog.Debug("pushing '" + k + "' to vault")
			err := cli.Push(k, v)
			if err != nil {
				slog.Debug("failed to push" + k + " to vault: " + err.Error())
				os.Exit(1)
				// panic("failed to push" + k + " to vault: " + err.Error())
			}
		}(k, v)
	}
	wg.Wait()
	slog.Debug("done pushing secrets")
}
