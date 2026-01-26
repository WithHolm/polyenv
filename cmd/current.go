// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/polyenvfile"
)

var Environment string
var PolyenvFile *polyenvfile.File

func init() {
	//genereate command and all sub commands for each environment
	env, e := polyenvfile.ListEnvironments()
	if e != nil {
		slog.Warn("failed to discover environments. environment scoped commands will not work.", "error", e)
		return
	}

	// for each detected environment, create a command and all sub commands
	// ie !{env} {commands}
	for _, v := range env {
		V := v
		cmd := &cobra.Command{
			Use:   fmt.Sprintf("!%s [command] [arguments]", V),
			Short: fmt.Sprintf("manage %s environment", V),
			Long:  fmt.Sprintf("manage %s environment", V),
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				Environment = V
				p, e := polyenvfile.OpenFile(V)
				if e != nil {
					return fmt.Errorf("failed to open polyenv file: %w", e)
				}
				PolyenvFile = p
				return nil
			},
		}

		cmd.AddCommand(generateAddCommand())
		cmd.AddCommand(generatePullCommand())
		cmd.AddCommand(generateEnvCommand())

		rootCmd.AddCommand(cmd)
	}

}
