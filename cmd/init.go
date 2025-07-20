package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/polyenvfile"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

var vaultType vaults.VaultType
var initargs []string

var initCmd = &cobra.Command{
	Use:   "init [optional:file]",
	Short: "init the current .env file to keyvault",
	Long: `
		init will set up the .env file for syncing with your enterprise-vault.
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		Path = tools.SetPathOrArg(Path, args)

		return nil
	},
	Run: initialize,
}

func init() {
	description := "quick init will lead you directly to the setup for the given vault"
	initCmd.Flags().Var(&vaultType, "type", description)
	initCmd.Flags().StringArrayVarP(&initargs, "arg", "a", []string{}, "arguments to pass to the vault, defined dotenv syle: --arg key=value. can be used multiple times")
	slog.Debug("init called", "args", initCmd)
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
	slog.Debug("init", "Path", Path, "Vault", vaultType)
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get current working directory: " + err.Error())
		os.Exit(1)
	}

	var newFileType string
	var newFileName string
	var existingFile string
	polyenvFile := polyenvfile.File{
		VaultMap: make(map[string]map[string]any),
		Secrets:  make(map[string]model.Secret),
		Vaults:   make(map[string]model.Vault),
		Options: polyenvfile.VaultOptions{
			HyphenToUnderscore:         true,
			UppercaseLocally:           true,
			UseDotSecretFileForSecrets: true,
		},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("No path defined").
				Description("No dotenv file defined, do you want to use an existing file or create a new one?").
				Options(
					huh.NewOption("Existing file", "existing"),
					huh.NewOption("New file", "new"),
				).
				Value(&newFileType),
		),
		huh.NewGroup(
			huh.NewInput().
				Description("Enter the name of the file to use. you can omit .env. it will be added automatically, unless you use the non-standard '{name}.env' format").
				Placeholder("dev").
				Suggestions([]string{"dev", "prod", "staging"}).
				Value(&newFileName),
			huh.NewNote().DescriptionFunc(func() string {
				if newFileName == "" {
					return ".env"
				}
				if strings.HasPrefix(newFileName, ".env") {
					return newFileName
				} else if !strings.HasSuffix(newFileName, ".env") {
					return strings.Join([]string{".env", newFileName}, ".")
				}
				return newFileName
			}, &newFileName),
		).WithHide(newFileType != "new"),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Existing file").
				Description("Existing file").
				OptionsFunc(func() []huh.Option[string] {
					out := []huh.Option[string]{}

					list, err := tools.GetAllFiles(cwd, []string{".env"})
					if err != nil {
						slog.Error("failed to get files: " + err.Error())
						os.Exit(1)
					}
					// filepath.Rel()
					for _, f := range list {
						if strings.HasSuffix(f, ".polyenv") {
							continue
						}
						relativePath, err := filepath.Rel(cwd, f)
						if err != nil {
							slog.Error("failed to get relative path: " + err.Error())
							os.Exit(1)
						}
						out = append(out, huh.NewOption(relativePath, f))
					}
					return out
				}, nil).
				Value(&existingFile),
		).WithHide(newFileType != "existing"),
	)
	tui.RunHuh(form)
	switch newFileType {
	case "new":
		Path = filepath.Join(cwd, newFileName)
	case "existing":
		Path = existingFile
		NameTag := filepath.Base(Path)
		slog.Info("str", "path", Path, "name", NameTag)
	}
	return

	envPath, envName := filepath.Split(Path)

	slog.Debug("init called", "path", envPath, "envfile", envName)

	AddVault(&polyenvFile, "", map[string]any{})

	//set global options
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Global options").
				Description("The next questions are about global options. all vaults will adhere to these").
				Next(true).
				NextLabel("Enter to continue"),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Global: Do you want to use dot vault for secrets?").
				Description("Any secrets you pull will be stored in a .env.secret.* file instead of .env file. \nThis makes it easier to add them to gitignore. while being accessible with any .env import system").
				Affirmative("Yes").
				Negative("No").
				Value(&polyenvFile.Options.UseDotSecretFileForSecrets),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Global: Convert hyphens to underscores in env name when setting setting new secrets?").
				Description("convert remote secret name 'my-secret' to 'my_secret' locally").
				Affirmative("Yes").
				Negative("No").
				Value(&polyenvFile.Options.HyphenToUnderscore),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Global: Automatically uppercase env name when setting new secrets?").
				Description("convert remote vault name 'my-secret' to 'MY-SECRET' locally").
				Affirmative("Yes").
				Negative("No").
				Value(&polyenvFile.Options.UppercaseLocally),
		),
	)
	tui.RunHuh(f)

	fmt.Print(polyenvFile)
	polyenvFile.Save()
	// itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).MarginRight(1).Bold(true)
	// commandStyle := lipgloss.NewStyle().MarginRight(1).Italic(true)
	// // list := list.New(
	// // 	"pull, output to terminal",
	// // 	list.New(fmt.Sprintf("polyenv pull --path %s --out term", Path)).ItemStyle(commandStyle),
	// // 	"pull, output to terminal as json",
	// // 	list.New(fmt.Sprintf("polyenv pull --path %s --out termjson", Path)).ItemStyle(commandStyle),
	// // 	fmt.Sprintf("pull, output to %s", Path),
	// // 	list.New(fmt.Sprintf("polyenv pull --path %s --out file", Path)).ItemStyle(commandStyle),
	// // 	"if output is not specified, it will default to terminal",
	// // ).Enumerator(list.Dash).ItemStyle(itemStyle)
	// slog.Info(fmt.Sprint(list))
}
