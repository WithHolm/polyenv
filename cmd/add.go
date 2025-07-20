package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/polyenvfile"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

var addCmd = &cobra.Command{
	Use:   "add [secret|vault] [optional:vault name]",
	Short: "init the current .env file to keyvault",
	Long: `
		init will set up the .env file for syncing with your enterprise-vault.
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		Path = tools.SetPathOrArg(Path, args)

		return nil
	},
	Run: initialize,
}

func init() {
	description := "quick init will lead you directly to the setup for the given vault"
	// addCmd.
	addCmd.Flags().Var(&vaultType, "type", description)
	addCmd.Flags().StringArrayVarP(&initargs, "arg", "a", []string{}, "arguments to pass to the vault, defined dotenv syle: --arg key=value. can be used multiple times")
	rootCmd.AddCommand(addCmd)
}

func add(cmd *cobra.Command, args []string) {
	slog.Error("Not implemented")
	os.Exit(1)
}

// Add Vault
// goes through the wizard and adds the vault to the polyenv file
func AddVault(polyenvFile *polyenvfile.File, vaultTypeStr string, vaultInitArgs map[string]any) {
	var vaultType vaults.VaultType
	if vaultTypeStr == "" {
		form := huh.NewForm(
			huh.NewGroup(
				vaults.VaultTypeSelector(&vaultType),
			),
		)
		tui.RunHuh(form)
	}

	vault, err := vaults.GetVaultInstance(string(vaultType))
	if err != nil {
		slog.Error("failed to start vault: " + err.Error())
		os.Exit(1)
	}

	//warm up the vault wizard
	e := vault.WizWarmup(vaultInitArgs)
	if e != nil {
		slog.Error("failed to start vault wizard", "error", e)
	}

	for {
		f := vault.WizNext()
		if f == nil {
			break
		}
		tui.RunHuh(f)
	}

	var vaultDisplayName string
	var addSecret bool
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Vault name").
				Description(fmt.Sprintf("Do you want to use '%s' as displayname for vault or enter a new one?", vault.DisplayName())).
				Placeholder(vault.DisplayName()).
				CharLimit(512).
				Validate(func(s string) error {
					for k, v := range polyenvFile.Vaults {
						if k == s {
							return fmt.Errorf("vault name already exists: %s", v.ToString())
						}
					}
					return nil
				}).
				Value(&vaultDisplayName),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add secret").
				Description("Do you want to select secret(s) to pull from the vault?").
				Affirmative("Yes").
				Negative("No").
				Value(&addSecret),
		),
	)
	tui.RunHuh(f)

	// get valt displayname and append to polyenv file
	if vaultDisplayName == "" {
		vaultDisplayName = vault.DisplayName()
	}
	polyenvFile.Vaults[vaultDisplayName] = vault

	if addSecret {
		AddSecret(polyenvFile, vaultDisplayName)
	}
}

// Add Secret
func AddSecret(polyenvFile *polyenvfile.File, vaultName string) {
	var vault model.Vault

	if vaultName == "" {
		vault = *polyenvFile.TuiSelectVault()
	} else {
		var ok bool
		vault, ok = polyenvFile.Vaults[vaultName]
		if !ok {
			slog.Error("vault not found", "vault", vaultName)
			os.Exit(1)
		}
	}

	// select secrets
	var selectedSecrets []model.Secret
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[model.Secret]().
				Title("Select secret(s)").
				Description("Multiple secrets can be selected. secrets with '!' are not enabled.").
				OptionsFunc(func() (opt []huh.Option[model.Secret]) {
					list, err := vault.List()
					if err != nil {
						slog.Error("failed to list secrets: " + err.Error())
						os.Exit(1)
					}
					for _, secret := range list {
						name := secret.RemoteKey
						if !secret.Enabled {
							name = "!" + name
						}
						s := fmt.Sprintf("%s (%s)", name, secret.ContentType)
						opt = append(opt, huh.NewOption(s, secret))
					}
					return opt
				}, nil).Value(&selectedSecrets),
		),
	)
	tui.RunHuh(f)

	//process secrets
	// groups:=make([]*huh.Group, 0)
	vaultmap := make(map[string]model.Secret)
	for _, secret := range selectedSecrets {
		var displayname string
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(secret.RemoteKey).
					Description("do you want to give the secret another name locally?").
					Prompt(secret.RemoteKey).
					Validate(func(s string) error {
						return polyenvFile.ValidateSecretName(s)
					}).Value(&displayname),
			),
		)
		tui.RunHuh(f)
		if displayname == "" {
			displayname = secret.RemoteKey
		}
		displayname = polyenvFile.Options.ConvertString(displayname)
		vaultmap[displayname] = secret
	}
}
