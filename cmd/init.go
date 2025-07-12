package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

var vaultType vaults.VaultType
var initargs []string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init the current .env file to keyvault",
	Long: `
		init will set up the .env file for syncing with your enterprise-vault.
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		argPath := tools.ExtractFilenameArg(args)
		if argPath != "" {
			Path = tools.AppendDotEnvExtension(argPath)
		}
		argVault, err := tools.ExtractVaultNameArg(args, vaults.ListVaultTypes())
		if err != nil {
			return err
		}
		if argVault != "" {
			vaultType = vaults.VaultType(argVault)
		}

		return nil
	},
	Run: initialize,
}

func init() {
	description := fmt.Sprintf("quick init will lead you directly to the setup for the given vault")
	initCmd.Flags().Var(&vaultType, "type", description)
	initCmd.Flags().StringArrayVarP(&initargs, "arg", "a", []string{}, "arguments to pass to the vault, defined dotenv syle: --arg key=value. can be used multiple times")

	rootCmd.AddCommand(initCmd)
}

func runHuh(f *huh.Form) {
	if f == nil {
		return
	}

	theme := huh.ThemeCatppuccin()
	theme.Focused.FocusedButton = theme.Blurred.FocusedButton.SetString("◉")
	theme.Focused.BlurredButton = theme.Blurred.BlurredButton.SetString("○")

	e := f.WithTheme(theme).WithProgramOptions(tea.WithAltScreen()).Run()

	if e != nil {
		fmt.Fprintf(os.Stderr, "failed to run wizard: %s\n", e.Error())
		os.Exit(1)
	}
}

/*
intended:
init --type vaulttype --arg key=value
init dev -> inits the file dev.env
init dev!keyvault -> inits the file dev.env and sets up the keyvault
init !keyvault -> inits the file local.env (default value) and sets up the keyvault
*/
func initialize(cmd *cobra.Command, args []string) {
	slog.Debug("init", "Path", Path, "Vault", vaultType)

	Path = tools.AppendDotEnvExtension(Path)

	// os.Exit(0)
	slog.Debug("init called", "envfile", Path)

	// theme := huh.ThemeCatppuccin()
	// theme.Focused.FocusedButton = theme.Blurred.FocusedButton.SetString("◉")
	// theme.Focused.BlurredButton = theme.Blurred.BlurredButton.SetString("○")

	if vaultType == "" {
		form := huh.NewForm(
			huh.NewGroup(
				vaults.VaultTypeSelector(&vaultType),
			),
		)
		runHuh(form)
	}

	Vault, err := vaults.NewInitVault(string(vaultType))
	if err != nil {
		slog.Error("failed to create vault: " + err.Error())
		os.Exit(1)
	}

	var InitArgs map[string]string
	if len(initargs) > 0 {
		var e error
		InitArgs, e = godotenv.Unmarshal(strings.Join(initargs, "\n"))
		if e != nil {
			slog.Error("failed to unmarshal vault arguments: " + e.Error())
			os.Exit(1)
		}
	}

	model := tui.NewInitModel(vaultType, InitArgs)
	// theme := huh.ThemeCatppuccin()
	// theme.Focused.FocusedButton = theme.Blurred.FocusedButton.SetString("◉")
	// theme.Focused.BlurredButton = theme.Blurred.BlurredButton.SetString("○")

	// p := tea.NewProgram(model, tea.WithAltScreen())
	p := tea.NewProgram(model)
	_, err = p.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Vault = model.Vault
	// if e != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to run wizard: %s\n", e.Error())
	// 	os.Exit(1)
	// }

	// err = Vault.WizardWarmup(InitArgs)
	// if err != nil {
	// 	slog.Error("failed to warm vault for init: " + err.Error())
	// 	os.Exit(1)
	// }

	// wizForm := Vault.WizardNext()
	// for wizForm != nil {
	// 	runHuh(wizForm)
	// 	// wizForm.Init()
	// 	// wizForm.NextField()
	// 	// wizForm.PrevField()
	// 	// err = wizForm.WithTheme(theme).WithProgramOptions(tea.WithAltScreen()).Run()
	// 	// if err != nil {
	// 	// 	slog.Error("failed to run wizard: " + err.Error())
	// 	// 	os.Exit(1)
	// 	// }
	// 	wizForm = Vault.WizardNext()
	// }

	slog.Debug("finished.. validating")
	VaultOpts := Vault.WizardComplete()
	// Vault.ValidateConfig(VaultOpts)

	// slog.Info("done setting up vault")
	err = Vault.ValidateConfig(VaultOpts)
	if err != nil {
		slog.Error("failed to validate vault '" + Vault.DisplayName() + "' options: " + err.Error())
		os.Exit(1)
	}

	err = vaults.WriteFile(Path, VaultOpts)
	if err != nil {
		slog.Error("failed to write vault options: " + err.Error())
		os.Exit(1)
	}

	//save the vault options
	// vaults.SaveVault(Vault, Path)
	// happy := true
	runHuh(
		huh.NewForm(huh.NewGroup(
			huh.NewNote().Title("WARNING").Description("add your dotenv file to .gitignore if you are going to pull to file!"),
		)),
	)

	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).MarginRight(1).Bold(true)
	commandStyle := lipgloss.NewStyle().MarginRight(1).Italic(true)
	list := list.New(
		"pull, output to terminal",
		list.New(fmt.Sprintf("polyenv pull --path %s --out term", Path)).ItemStyle(commandStyle),
		"pull, output to terminal as json",
		list.New(fmt.Sprintf("polyenv pull --path %s --out termjson", Path)).ItemStyle(commandStyle),
		fmt.Sprintf("pull, output to %s", Path),
		list.New(fmt.Sprintf("polyenv pull --path %s --out file", Path)).ItemStyle(commandStyle),
		"if output is not specified, it will default to terminal",
	).Enumerator(list.Dash).ItemStyle(itemStyle)
	slog.Info(fmt.Sprint(list))
}
