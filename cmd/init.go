package cmd

import (
	"dotenv-myvault/internal/charmselect"
	"dotenv-myvault/internal/vaults"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init the current .env file to keyvault",
	Long: `
		init will set up the .env file for syncing with your enterprise-vault.
	`,
	Run: initialize,
}

func init() {
	rootCmd.AddCommand(initCmd)
	// initCmd.Flags().StringVarP(&Path, "path", "p", ".env", "path to the .env file to push. uses /.env by default")
}

func initialize(cmd *cobra.Command, args []string) {
	slog.Info("init")
	secretSelect := charmselect.New()
	secretSelect.Title = "What secret provider do you want to use?"
	secretSelect.AddItem("keyvault", "Azure Key Vault")
	secretSelection := secretSelect.Run()

	vault, err := vaults.NewInitVault(secretSelection[0].Key())
	if err != nil {
		slog.Debug("failed to create vault: " + err.Error())
		os.Exit(1)
	}
	slog.Debug("vault wizard initiated:", "provider", secretSelection[0].Key())
	vault.WizardWarmup()
	for {
		q := vault.WizardNext()

		if q.Title == "" {
			break
		}

		chrmSelect := charmselect.New()
		chrmSelect.Title = q.Title
		for _, q := range q.Questions {
			chrmSelect.AddItem(q.Key, q.Description)
		}
		selectresult := chrmSelect.Run()

		if len(selectresult) == 0 {
			slog.Debug("no selection made.. exiting")
			os.Exit(0)
		}

		slog.Debug("selected:", "q", q.Title, "key", selectresult[0].Key(), "description", selectresult[0].Description())
		// call callback function for question
		q.Callback(selectresult[0].Key())
	}
	opts := vault.WizardComplete()
	err = vault.SetOptions(opts)
	if err != nil {
		slog.Error("failed to set options for vault: " + err.Error())
		os.Exit(1)
	}

	err = vault.Warmup()
	if err != nil {
		slog.Error("failed to warm vault: " + err.Error())
		os.Exit(1)
	}

	vaults.SaveVault(vault, Path)

	slog.Debug("done")
}
