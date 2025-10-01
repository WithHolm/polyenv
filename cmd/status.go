// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss/list"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/polyenvfile"
)

func init() {
	rootCmd.AddCommand(generateStatusCommand())
}

func generateStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "show the status of environments",
		Run:   status,
	}
}

func status(cmd *cobra.Command, args []string) {
	envs := []string{}
	if Environment == "" {
		var err error
		envs, err = polyenvfile.ListEnvironments()
		if err != nil {
			slog.Error("failed to list environments", "error", err)
			os.Exit(1)
		}

	} else {
		envs = append(envs, Environment)
	}
	li := list.New()
	for _, env := range envs {
		slog.Debug("checking", "env", env)
		p, e := polyenvfile.OpenFile(env)
		if e != nil {
			slog.Error("failed to open polyenv file", "error", e)
			os.Exit(1)
		}
		vaultList := list.New()
		for k, v := range p.Vaults {
			secList := list.New()
			for secKey, secVal := range p.Secrets {
				if secVal.Vault != k {
					continue
				}
				secList.Item(fmt.Sprintf("secret %s -> %s (%s)", secVal.RemoteKey, secKey, secVal.ContentType))
			}

			slog.Debug("vault", "name", k, "vault", v.String())
			vaultList.Items(fmt.Sprintf("vault '%s' -> (%s)", k, v.String()), secList)
		}
		// if env == "" {
		// 	env = "<none>"
		// }
		li.Items("!"+env, vaultList)
	}
	fmt.Println(li)
}
