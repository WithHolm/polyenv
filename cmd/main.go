// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/withholm/polyenv/internal/tools"
)

// var author string

// var Path string //path to the .env file.. used by all commands

// is debug mode enabled
var Debug bool

// some other thing
var DisableTruncateDebug bool

var (
	// These variables are populated by the Go linker during the build process.
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "polyenv [action] [arguments]",
	Short: "a version of dotenv vault that can use other possible providers instead of the 'standard' dotenv-vault.",
	Long: `
		manage your .env files, enables you to use other possible providers instead of the 'standard' dotenv-vault.
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//log
		var cmdName string
		if cmd.Name() == "version" {
			cmdName = "version"
		} else {
			cmdName = cmd.Name()
		}

		if DisableTruncateDebug {
			tools.AppConfig().SetTruncateDebug(false)
		}

		slog.Debug("command", "name", cmdName)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			slog.Debug("flag", "name", f.Name, "value", f.Value)
		})
		slog.Debug("args", "values", args)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of polyenv",
	Long: `
		Print the version number of polyenv.
		`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("polyenv %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
	},
}

func init() {
	//enable PersistentPreRun on all levels
	cobra.EnableTraverseRunHooks = true
	// add persistend flags (flags that are set for all commands)
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&DisableTruncateDebug, "disable-truncate-debug", false, "dont truncate debug logging")
	// add version command

	rootCmd.AddCommand(versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
