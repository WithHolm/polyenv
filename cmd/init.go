package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/vaults"
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
}

func initialize(cmd *cobra.Command, args []string) {
	slog.Debug("init called", "envfile", Path)
	var vaultKey string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a secret provider").
				Options(
					vaults.GetVaultsAsOptions()...,
				).Value(&vaultKey),
		),
	)
	e := form.Run()
	if e != nil {
		fmt.Fprintf(os.Stderr, "failed to run wizard: %s\n", e.Error())
		os.Exit(1)
	}
	slog.Info("selected vault", "vault", vaultKey)

	Vault, err := vaults.NewInitVault(vaultKey)
	if err != nil {
		slog.Error("failed to create vault: " + err.Error())
		os.Exit(1)
	}

	err = Vault.WizardWarmup()
	if err != nil {
		slog.Error("failed to warm vault for init: " + err.Error())
		os.Exit(1)
	}

	wizForm := Vault.WizardNext()

	for wizForm != nil {
		err = wizForm.Run()
		if err != nil {
			slog.Error("failed to run wizard: " + err.Error())
			os.Exit(1)
		}
		wizForm = Vault.WizardNext()
	}

	slog.Info("done setting up vault")
	_, e = vaults.NewVault(vaultKey, Vault.WizardComplete())

	if e != nil {
		slog.Error("failed to create vault: " + e.Error())
		os.Exit(1)
	}

	//save the vault options
	vaults.SaveVault(Vault, Path)

	//ask if they want to pull imidiately
	// var pull bool
	// form = huh.NewForm(
	// 	huh.NewGroup(
	// 		huh.NewConfirm().
	// 			Title("Do you want to pull all secrets from the vault?").
	// 			DescriptionFunc(func() string {
	// 				//if there is a file on path, ask to confirm..
	// 				if f, err := os.Stat(Path); err == nil && f.Size() > 0 {
	// 					return "This will overwrite any existing .env files"
	// 				}
	// 				return ""
	// 			}, nil).
	// 			Affirmative("Yes, yank it daddy").
	// 			Negative("No, I'll do it later").
	// 			Value(&pull),
	// 	),
	// )
	// form.Run()
	// if pull {
	// 	Vault.Pull()
	// }

	slog.Debug("done")
}
