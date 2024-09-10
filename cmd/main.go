package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var author string
var Path string //path to the .env file.. used by all commands

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dotenv-myvault [ACTION] [ARGUMENTS]",
	Short: "a version of dotenv vault that can use other possible providers instead of the 'standard' dotenv-vault.",
	Long: `
		A version of dotenv that uses keyvault as 'vault' instead of the dotenv projects default one. 
		Requires the user to have active access to the specified keyvault when this command is run".
	`,
	// run:

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// TODO: SLOG INIT THING HERE
func init() {
	// add push and pull commands
	rootCmd.PersistentFlags().StringVar(&author, "author", "Philip Meholm (withholm)", "Author name for copyright attribution")

	opts := &slog.HandlerOptions{
		// AddSource: true,
		// Level:     slog.LevelInfo,
	}

	_, b := os.LookupEnv("SLOG_ADD_SOURCE")
	if b == true {
		opts.AddSource = true
	}
	s, b := os.LookupEnv("SLOG_LEVEL")
	if b == true {
		switch s {
		case "debug":
			opts.Level = slog.LevelDebug
		case "info":
			opts.Level = slog.LevelInfo
		case "warn":
			opts.Level = slog.LevelWarn
		case "error":
			opts.Level = slog.LevelError
		}
	} else {
		opts.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
