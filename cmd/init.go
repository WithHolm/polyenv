package cmd

import (
	"fmt"
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
var acceptPolyenvDefaults bool
var vaultTypes []string

// var checkgitignore bool

var initCmd = &cobra.Command{
	Use:   "init [environment] [--type vaulttype] [--arg key=value]...",
	Short: "initiales a new environment",
	Long: `
		init will set up environment for managment.
	`,
	Args: cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			Environment = args[0]
		}

		if vaultType != "" && !slices.Contains(vaultTypes, vaultType) {
			slog.Error("invalid vault type", "type", vaultType)
			cmd.Usage()
			return
		}

	},
	Run: initialize,
}

func init() {
	//append 'none' to skip vault creation for demos..
	vaultTypes = vaults.List()
	vaultTypes = append(vaultTypes, "none")
	description := fmt.Sprintf("quick init will lead you directly to the setup for the given vault: %s", vaultTypes)

	initCmd.Flags().StringVar(&vaultType, "vault", "", description)
	initCmd.Flags().StringArrayVarP(&initargs, "arg", "a", []string{}, "arguments to pass to the vault, defined dotenv syle: --arg key=value. can be used multiple times")
	err := initCmd.RegisterFlagCompletionFunc("vault", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return vaultTypes, cobra.ShellCompDirectiveKeepOrder
	})
	initCmd.Flags().BoolVar(&acceptPolyenvDefaults, "accept-default-settings", false, "Accept default settings for polyenv file")
	cobra.CheckErr(err)
	rootCmd.AddCommand(initCmd)
}

func initialize(cmd *cobra.Command, args []string) {
	// if checkgitignore {
	// 	polyenvfile.GitignoreMatchesEnvSecret()
	// 	os.Exit(0)
	// }

	file := polyenvfile.TuiNewFile(Environment)

	file.TuiAddOpts(nil, acceptPolyenvDefaults)

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

	if shouldAddVault && vaultType != "none" {
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

	slog.Info("polyenv file created", "path", file.Path, "name", file.Name)
	slog.Info("You can have this anywhere in your project. It will store info needed to pull secrets from vaults.")
	slog.Info(fmt.Sprintf("run 'polyenv !%s' to see what you can do with this env", Environment))
}
