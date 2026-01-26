// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/tui"
)

var addSecretCmds []*cobra.Command
var addVaultCmds []*cobra.Command
var addCmds []*cobra.Command

// var addVaultArgs []string

func generateAddCommand() *cobra.Command {
	var addVaultCmd = &cobra.Command{
		Use:   "vault [vault type]",
		Short: "add a new vault to the environment",
		Long: `
		add a new vault to the environment
	`,
		RunE: addVault,
	}
	//TODO: add args to addvault..

	addVaultCmds = append(addVaultCmds, addVaultCmd)

	var addSecretCmd = &cobra.Command{
		Use:   "secret [vault name]",
		Short: "add a new secret to your environment",
		Long: `
		add a new secret to the environment
	`,
		RunE: addSecret,
	}

	addSecretCmds = append(addSecretCmds, addSecretCmd)

	var addCmd = &cobra.Command{
		Use:   "add [secret|vault] [arguments]",
		Short: "add a new secret or vault to the environment",
		Long: `
		add will add a new secret or vault to the environment
	`,
		RunE: add,
	}
	addCmds = append(addCmds, addCmd)

	addCmd.AddCommand(addVaultCmd)
	addCmd.AddCommand(addSecretCmd)

	return addCmd
}

func add(cmd *cobra.Command, args []string) error {
	var selected *cobra.Command
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*cobra.Command]().OptionsFunc(func() (o []huh.Option[*cobra.Command]) {
				for _, v := range cmd.Commands() {
					o = append(o, huh.NewOption(v.Name(), v))
				}
				return o
			}, nil).Value(&selected).Title("Select what to add"),
		),
	)
	tui.RunHuh(f)
	if selected == nil {
		return fmt.Errorf("no vault selected")
	}
	selected.Run(cmd, args)
	return nil
}

func addSecret(cmd *cobra.Command, args []string) error {
	var Vault string
	if len(args) == 0 {
		Vault = PolyenvFile.TuiSelectVault()
	} else {
		Vault = args[0]
		if _, ok := PolyenvFile.Vaults[Vault]; !ok {
			slog.Error("vault not found", "vault", Vault)
			os.Exit(1)
		}
	}

	PolyenvFile.TuiAddSecret(Vault)
	return PolyenvFile.Save()
}

func addVault(cmd *cobra.Command, args []string) error {
	err := PolyenvFile.TuiAddVault("", map[string]any{})
	if err != nil {
		return err //fmt.Errorf("failed to add vault: %w", err)
	}
	return PolyenvFile.Save()
}
