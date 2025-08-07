package cmd

import (
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/polyenvfile"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

var vaultType string
var initargs []string
var checkgitignore bool

var initCmd = &cobra.Command{
	Use:   "init [environment] [--type vaulttype] [--arg key=value]...",
	Short: "initiales a new environment",
	Long: `
		init will set up environment for managment.
	`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			Environment = args[0]
		}

		if vaultType != "" && !slices.Contains(vaults.List(), vaultType) {
			slog.Error("invalid vault type", "type", vaultType)
			cmd.Usage()
			return
		}

	},
	Run: initialize,
}

func init() {
	description := "quick init will lead you directly to the setup for the given vault"
	initCmd.Flags().StringVar(&vaultType, "type", "", description)
	// initCmd.Flags().BoolVar(&useDefaultConfig, "use-default-config", false, "")
	initCmd.Flags().StringArrayVarP(&initargs, "arg", "a", []string{}, "arguments to pass to the vault, defined dotenv syle: --arg key=value. can be used multiple times")
	err := initCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return vaults.List(), cobra.ShellCompDirectiveKeepOrder
	})
	cobra.CheckErr(err)
	rootCmd.AddCommand(initCmd)
}

/*
intended:
init --type vaulttype --arg key=value
init dev -> inits the file dev.env
init dev!keyvault -> inits the file dev.env and sets up the keyvault
init !keyvault -> inits the file local.env (default value) and sets up the keyvault
*/
func initialize(cmd *cobra.Command, args []string) {
	if checkgitignore {
		polyenvfile.GitignoreMatchesEnvSecret()
		os.Exit(0)
	}

	file := polyenvfile.TuiNewFile(Environment)

	//set global options
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Global options").
				Description("Set options. all vaults will adhere to these..").
				Next(true).
				NextLabel("Enter to continue"),
		),
	)
	tui.RunHuh(f)
	file.TuiAddOpts(nil, false)

	e := file.TuiAddGitIgnore()
	if e != nil {
		slog.Error("Failed to add data to gitignore", "err", e)
		os.Exit(1)
	}

	// return
	// Determine if we should proceed with adding a vault.
	// We do this if --type or --arg flags are provided, otherwise we ask the user.
	shouldAddVault := vaultType != "" || len(initargs) > 0
	if !shouldAddVault {
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add vault to '" + file.Name + "'").
					Description("Do you want to add a new vault?").
					Affirmative("Yes").
					Negative("No").
					Value(&shouldAddVault),
			),
		)
		tui.RunHuh(f)
	}

	if shouldAddVault {
		vaultArgs := map[string]any{}
		if len(initargs) > 0 {
			if vaultType == "" {
				slog.Warn("new vault arguments defined, but no vault type specified. lets hope you select the correct vault :)")
			}

			for _, v := range initargs {
				key, val, ok := strings.Cut(v, "=")
				if !ok {
					slog.Warn("failed to parse argument", "argument", v)
					continue
				}
				slog.Debug("parsed argument", "key", key, "val", val)
				vaultArgs[key] = val
			}
		}
		file.TuiAddVault(string(vaultType), vaultArgs)
	}

}
