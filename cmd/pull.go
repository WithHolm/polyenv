package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/vaults"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type PullOutputType string

const (
	file         PullOutputType = "file"
	terminal     PullOutputType = "terminal"
	terminaljson PullOutputType = "terminaljson"
)

var pullPath string
var pullOutput string
var pullOutputType = file
var allPullOutputTypes = []PullOutputType{file, terminal, terminaljson}

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
	PullCmd.Flags().VarP(&pullOutputType, "out", "o", "where to post the results of the pull. 'terminal' for directly to terminal, 'file' for .env file")
	rootCmd.AddCommand(PullCmd)
}

// execute envault pull
func pull(cmd *cobra.Command, args []string) {
	// slog.Debug("pull called", "args", args)
	pullPath = Path
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
	slog.Debug("pull path", "path", pullPath)

	// check if opts file exists..
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

	// Init vault using config
	cli, err := vaults.NewVault(vaultOpts["VAULT_TYPE"], vaultOpts)
	if err != nil {
		log.Fatal("failed to create vault: " + err.Error())
		os.Exit(1)
	}
	// Pull secrets
	secrets, err := cli.Pull()
	if err != nil {
		log.Fatal("failed to pull secrets: " + err.Error())
		os.Exit(1)
	}

	err = outputSecrets(secrets, pullOutputType)
	if err != nil {
		log.Fatal("failed to output secrets: " + err.Error())
		os.Exit(1)
	}

	slog.Debug("done pulling secrets", "path", pullPath, "len", len(secrets))
}

func outputSecrets(secrets map[string]string, outputType PullOutputType) error {
	switch outputType {
	case terminal:
		for key, value := range secrets {
			fmt.Println(key + "=" + value)
		}
	case terminaljson:
		json, err := json.Marshal(secrets)
		if err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}
		fmt.Println(string(json))
	case file:
		//make file if it doesnt exist
		if _, err := os.Stat(pullPath); os.IsNotExist(err) {
			slog.Debug(fmt.Sprintf("creating .env file at %s", pullPath))
			err := os.WriteFile(pullPath, []byte{}, 0644)
			if err != nil {
				return fmt.Errorf("failed to create .env file: %s", err)
			}
		}

		err := godotenv.Write(secrets, pullPath)
		if err != nil {
			return fmt.Errorf("failed to write .env file: %s", err)
		}
	}
	return nil
}

// outputs type as string
func (out *PullOutputType) String() string {
	return string(*out)
}

// sets output type
func (out *PullOutputType) Set(value string) error {
	var errmgs = make([]string, 0)
	for _, typ := range allPullOutputTypes {
		if typ == PullOutputType(value) {
			*out = PullOutputType(value)
			return nil
		}
		// append error message, if none is found
		errmgs = append(errmgs, string(typ))
	}

	//return error if not found
	return fmt.Errorf("invalid output type: %s, must be %s", value, strings.Join(errmgs, ", "))
}

// returns output type
func (out *PullOutputType) Type() string {
	return "outputType"
}
