package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/vaults"
)

var validateCmd = &cobra.Command{
	Use:   "Validate",
	Short: "validate the .polyenv file",
	Long: `
		Validate will validate the .polyenv file.
	`,
	Run: validate,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func validate(cmd *cobra.Command, args []string) {
	slog.Debug("validate called", "args", args)
	setPathOrArg(args)

	opts, err := vaults.ReadFile(Path)

	if err != nil {
		slog.Error("failed to read vault options file", "path", Path)
		os.Exit(1)
	}

	err = vaults.VaildateVaultOpts(opts)
	if err != nil {
		slog.Error("failed to validate vault options", "path", Path)
		os.Exit(1)
	}

	// e := tools.TestVaultFileExists(Path)
	// if e != nil {
	// 	slog.Error("error checking env file", "path", Path)
	// 	os.Exit(1)
	// }

	// optsPath := tools.GetVaultOptsPath(Path)
	// if !tools.VaultOptsExist(optsPath) {
	// 	slog.Error("no vault options file found", "path", Path)
	// 	os.Exit(1)
	// }

	// opts := vaults.
	// 	slog.Debug("found vault options file", "path", optsPath)

}
