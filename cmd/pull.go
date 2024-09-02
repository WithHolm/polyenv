package cmd

import (
	"dotenv-keyvault/internal/tools"
	"dotenv-keyvault/internal/vaults"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var pullPath string
var pullOutput string

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull all secrets from keyvault",
	Long: `
		pull all secrets from keyvault. 
		Any pull will also override existing .env file. 
	`,
	Run: pull,
}

func init() {
	rootCmd.AddCommand(PullCmd)
	PullCmd.Flags().StringVarP(&pullPath, "path", "p", ".env", "path to the '.env' file to pull. appends '.vaultopts' when searching. Uses /.env by default")
	PullCmd.Flags().StringVarP(&pullOutput, "out", "o", "env", "where to post the results of the pull. 'env' for directly to env variables, 'file' for .env file")
}

func pull(cmd *cobra.Command, args []string) {
	fmt.Println("pull called")

	if pullPath != "" {
		err := tools.CheckDoubleDashS(pullPath, "path")
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}

		// in case they set '--path env.vaultopts'
		if pullPath == tools.GetVaultOptsPath(pullPath) {
			log.Fatal("--path cannot be set to the vault options file")
			os.Exit(1)
		}

	}

	if pullOutput != "" {
		err := tools.CheckDoubleDashS(pullOutput, "out")
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
	}

	// get absolute path
	if !filepath.IsAbs(pullPath) {
		// path is absolute
		_path, err := filepath.Abs(pullPath)
		if err != nil {
			log.Fatal("failed to get absolute path: " + err.Error())
			os.Exit(1)
		}
		pullPath = _path
	}
	// check if opts file exists
	if !tools.VaultOptsExist(pullPath) {
		optfile := tools.GetVaultOptsPath(pullPath)
		log.Fatal("no vault options file found: ", optfile)
		os.Exit(1)
	}

	// read opts file
	vaultOpts, err := tools.GetVaultFile(pullPath)
	if err != nil {
		log.Fatal("failed to get vault options: " + err.Error())
		os.Exit(1)
	}

	// connect and pull
	cli, err := vaults.NewVault(vaultOpts["VAULT_TYPE"], vaultOpts["VAULT_NAME"], vaultOpts)
	secrets, err := cli.Pull()
	if err != nil {
		log.Fatal("failed to pull secrets: " + err.Error())
		os.Exit(1)
	}

	if pullOutput == "env" {
		for k, v := range secrets {
			log.Println("setting env variable: " + k)
			os.Setenv(k, v)
		}
	}

	if pullOutput == "file" {
		err := godotenv.Write(secrets, pullPath)
		if err != nil {
			log.Fatal("failed to write .env file: " + err.Error())
			os.Exit(1)
		}
	}

}
