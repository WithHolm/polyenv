package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var author string
var Path string //path to the .env file.. used by all commands
var Debug bool

//var logger *slog.Logger

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "polyenv [action] [arguments]",
	Short: "a version of dotenv vault that can use other possible providers instead of the 'standard' dotenv-vault.",
	Long: `
		manage your .env files, enables you to use other possible providers instead of the 'standard' dotenv-vault.
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logger
		// opts := &slog.HandlerOptions{
		// 	Level: slog.LevelInfo,
		// }

		opts := log.Options{
			Level:        log.InfoLevel,
			ReportCaller: Debug,
		}
		if Debug {
			opts.Level = log.DebugLevel
			// opts.ReportCaller = true
		}
		handler := log.NewWithOptions(os.Stderr, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	},
}

func init() {
	// add persistend flags (flags that are set for all commands)
	// rootCmd.PersistentFlags().StringVar(&author, "author", "Philip Meholm (withholm)", "Author name for copyright attribution")
	rootCmd.PersistentFlags().StringVarP(&Path, "path", "p", ".env", "Path to the .env file")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Enable debug logging")
	// rootCmd.PersistentFlags().BoolVar(&Debug, "whatif", false, "Enable whatif. will not push or pull, but will show what would be done")

	// add version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of polyenv",
		Long: `
		Print the version number of polyenv.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version 0.0.1")
		},
	}
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

// // set path
// func setPathOrArg(args []string) {
// 	if len(args) >= 1 && Path != "" {
// 		slog.Error("Both --path and positional arguments are set. Please use only one of the two.")
// 		os.Exit(1)
// 	} else if len(args) == 1 && Path == "" {
// 		slog.Debug("using positional argument as path", "path", args[0])
// 		Path = args[0]
// 	}

// 	if Path == "" {
// 		slog.Error("no path set. please set --path or positional argument")
// 		os.Exit(1)
// 	}
// }
