package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/polyenvfile"
)

var Environment string
var PolyenvFile *polyenvfile.File

func init() {
	//genereate command and all sub commands for each environment
	env, e := polyenvfile.ListEnvironments()
	if e != nil {
		slog.Error("failed to list environments", "error", e)
		os.Exit(1)
	}
	for _, v := range env {
		V := v
		cmd := &cobra.Command{
			Use:   fmt.Sprintf("!%s [command] [arguments]", V),
			Short: fmt.Sprintf("manage %s environment", V),
			Long:  fmt.Sprintf("manage %s environment", V),
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				Environment = V
				p, e := polyenvfile.OpenFile(V)
				if e != nil {
					slog.Error("failed to open polyenv file", "error", e)
					os.Exit(1)
				}
				PolyenvFile = &p
			},
		}

		cmd.AddCommand(generateAddCommand())
		cmd.AddCommand(generatePullCommand())
		cmd.AddCommand(generateEnvCommand())

		rootCmd.AddCommand(cmd)
	}

}
