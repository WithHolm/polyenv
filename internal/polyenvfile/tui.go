package polyenvfile

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
	"github.com/withholm/polyenv/internal/tui"
	"github.com/withholm/polyenv/internal/vaults"
)

//region Vault

// TuiAddVault adds a new vault to the polyenv file via the tui
func (file *File) TuiAddVault(vaultTypeStr string, vaultInitArgs map[string]any) {
	// var vaultType vaults.VaultType
	if vaultTypeStr == "" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					OptionsFunc(func() (ret []huh.Option[string]) {
						for _, k := range vaults.List() {
							v, e := vaults.NewVaultInstance(k)
							if e != nil {
								slog.Error("failed to get vault", "error", e)
								os.Exit(1)
							}
							opt := huh.NewOption(v.DisplayName(), k)
							ret = append(ret, opt)
						}
						return ret
					}, vaultTypeStr).Value(&vaultTypeStr),
			),
		)
		tui.RunHuh(form)
	}

	vault, err := vaults.NewVaultInstance(vaultTypeStr)
	if err != nil {
		slog.Error("failed to start vault: " + err.Error())
		os.Exit(1)
	}

	//warm up the vault wizard
	e := vault.WizWarmup(vaultInitArgs)
	if e != nil {
		slog.Error("failed to start vault wizard", "error", e)
	}

	for {
		f, e := vault.WizNext()
		if e != nil {
			slog.Error("failed to get next form", "error", e)
			os.Exit(1)
		}
		if f == nil {
			break
		}
		tui.RunHuh(f)
	}

	err = vault.WizComplete()
	if err != nil {
		slog.Error("failed to complete vault wizard", "error", err)
		os.Exit(1)
	}

	var vaultDisplayName string
	var addSecret bool
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Vault name").
				Description(fmt.Sprintf("Do you want to use '%s' as displayname for vault or enter a new one?", vault.DisplayName())).
				Placeholder(vault.DisplayName()).
				CharLimit(512).
				Validate(func(s string) error {
					for k, v := range file.Vaults {
						if k == s {
							return fmt.Errorf("vault name already exists: %s", v.String())
						}
					}
					return nil
				}).
				Value(&vaultDisplayName),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add secret").
				Description("Do you want to select secret(s) to pull from the vault?").
				Affirmative("Yes").
				Negative("No").
				Value(&addSecret),
		),
	)
	tui.RunHuh(f)

	// get valt displayname and append to polyenv file
	if vaultDisplayName == "" {
		vaultDisplayName = vault.DisplayName()
	}

	if file.Vaults == nil {
		file.Vaults = make(map[string]model.Vault)
	}
	file.Vaults[vaultDisplayName] = vault
	file.Save()

	err = vault.Warmup()
	if err != nil {
		slog.Error("failed to warmup vault", "error", err)
		os.Exit(1)
	}

	if addSecret {
		file.TuiAddSecret(vaultDisplayName)
	}
}

// Tui Select vault from list
func (file *File) TuiSelectVault() string {
	var displayName string
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select vault").
				OptionsFunc(func() (ret []huh.Option[string]) {
					for k, v := range file.Vaults {
						ret = append(ret, huh.NewOption(fmt.Sprintf("%s (%s)", k, v.String()), k))
					}
					return ret
				}, nil).
				Value(&displayName),
		),
	)
	tui.RunHuh(f)
	vault, ok := file.Vaults[displayName]
	if !ok {
		slog.Error("vault not found", "vault", displayName)
		os.Exit(1)
	}
	slog.Debug("selected vault", "vault", vault.String())
	return displayName
}

// region secret
// add a new secret to the polyenv file via the tui. requires displayname of already existing vault
func (file *File) TuiAddSecret(vaultName string) {
	if vaultName == "" {
		slog.Error("secret name cannot be empty")
		os.Exit(1)
	}
	v, ok := file.Vaults[vaultName]
	if !ok {
		slog.Error("vault not found", "vault", vaultName)
		os.Exit(1)
	}
	err := v.Warmup() //making sure its ready to use
	if err != nil {
		slog.Error("failed to warmup vault", "error", err)
		os.Exit(1)
	}

	// select secrets
	var selectedSecrets []model.Secret

	// if vault had its own secret selection form, use that
	handledByVault := v.SecretSelectionHandler(&selectedSecrets)

	// otherwise use the default form
	if !handledByVault {
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[model.Secret]().
					Title("Select secret(s)").
					Description("Multiple secrets can be selected. secrets with '!' are not enabled.").
					OptionsFunc(func() (opt []huh.Option[model.Secret]) {
						e := v.ListElevate()
						if e != nil {
							slog.Error("failed to elevate permissions", "error", e)
							os.Exit(1)
						}

						list, err := v.List()
						if err != nil {
							slog.Error("failed to list secrets: " + err.Error())
							os.Exit(1)
						}

						pre := make([]huh.Option[model.Secret], 0)
						for _, secret := range list {
							localSecret, hasLocalSecret := file.GetSecretInfo(secret.RemoteKey, vaultName)

							slog.Debug("secret", "name", secret.RemoteKey, "enabled", secret.Enabled, "local", hasLocalSecret)

							secretName := secret.RemoteKey
							if !secret.Enabled {
								secretName = "!" + secretName
							}
							s := fmt.Sprintf("%s (%s)", secretName, secret.ContentType)
							o := huh.NewOption(s, secret)
							if hasLocalSecret {
								o.Key += fmt.Sprintf(" (%s)", localSecret.LocalKey)
								o = o.Selected(true)
								pre = append(pre, o)
								continue
							}
							opt = append(opt, o)
						}
						return slices.Concat(pre, opt)
					}, nil).Value(&selectedSecrets),
			),
		)
		tui.RunHuh(f)
	}

	//process secrets
	if file.Secrets == nil {
		file.Secrets = make(map[string]model.Secret)
	}
	for _, secret := range selectedSecrets {
		localSecret, hasLocalSecret := file.GetSecretInfo(secret.RemoteKey, vaultName)
		var displayname string
		f := huh.NewForm(
			//set local name for the remote secret
			huh.NewGroup(
				huh.NewInput().
					Title(secret.RemoteKey).
					DescriptionFunc(func() string {
						if hasLocalSecret {
							return fmt.Sprintf("do you want to change the local name? Enter will use the current name: %s", localSecret.LocalKey)
						}
						return "select name to use when referencing in env? Enter will use the remote name."
					}, nil).
					PlaceholderFunc(func() string {
						if hasLocalSecret {
							return localSecret.LocalKey
						}
						return secret.RemoteKey
					}, nil).
					Validate(func(s string) error {
						v, ok := file.Secrets[s]
						if ok {
							return fmt.Errorf("secret name already exists: %s", v.ToString())
						}
						return file.ValidateSecretName(s)
					}).Value(&displayname),
				huh.NewNote().TitleFunc(func() string {
					if displayname == "" && hasLocalSecret {
						return file.Options.ConvertString(localSecret.LocalKey)
					} else if displayname == "" {
						return file.Options.ConvertString(secret.RemoteKey)
					}
					return file.Options.ConvertString(displayname)
				}, &displayname),
			),
		)
		tui.RunHuh(f)
		if displayname == "" && hasLocalSecret {
			displayname = localSecret.LocalKey
		} else if displayname == "" {
			displayname = secret.RemoteKey
		}
		secret.Vault = vaultName
		displayname = file.Options.ConvertString(displayname)

		if hasLocalSecret {
			delete(file.Secrets, localSecret.LocalKey)
		}
		secret.LocalKey = displayname
		file.Secrets[displayname] = secret
	}

	//remove secrets from local that where de-selected during selection
	for k, secret := range file.Secrets {
		if secret.Vault != vaultName {
			continue
		}
		//check if current secret is in selectedSecrets. if its not, remove it from local
		remove := true
		for _, v := range selectedSecrets {
			if v.RemoteKey == secret.RemoteKey {
				remove = false
				break
			}
		}
		if remove {
			slog.Debug("removing secret from local", "name", k)
			delete(file.Secrets, k)
		}
	}
	file.Save()

}

//region opts

// TODO: add indivitial checks for each option
// add options to the polyenv file
func (file *File) TuiAddOpts(opts *VaultOptions, acceptDefaults bool) {
	if opts != nil {
		file.Options = *opts
	}

	if !acceptDefaults {
		keep := true
		tui.RunHuh(
			huh.NewForm(
				huh.NewGroup(
					huh.NewNote().Description(file.Options.ListCurrentOptions()),
					huh.NewConfirm().
						Title("Do you want to keep these or edit them?").
						Affirmative("Keep").
						Negative("Edit").
						Value(&keep).WithButtonAlignment(lipgloss.Left),
				),
			),
		)
		acceptDefaults = keep
	}

	if !acceptDefaults {
		tui.RunHuh(file.Options.TuiOpts())
	}

	file.Save()
}

//region other

// New file via tui. returns the polyenv file that is saved to disk
func TuiNewFile(env string) (file *File) {
	c, e := tools.GetGitRootOrCwd()
	if e != nil {
		slog.Error("failed to get project root: " + e.Error())
		os.Exit(1)
	}

	file = &File{
		VaultMap: make(map[string]map[string]any),
		Secrets:  make(map[string]model.Secret),
		Vaults:   make(map[string]model.Vault),
		Options: VaultOptions{
			HyphenToUnderscore:         true,
			UppercaseLocally:           true,
			UseDotSecretFileForSecrets: true,
		},
	}

	if env == "" {
		var newFileType string
		var newFileName string
		var existingFile string

		// tui: create from envfile name or define new env
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Init").
					Description("do you want to use an existing env file as baseline or define a new environment?\nPolyenv files follows 'environment', and will use all available files given a environment.").
					Options(
						huh.NewOption("Existing", "existing"),
						huh.NewOption("New", "new"),
					).
					Value(&newFileType),
			),
		)
		tui.RunHuh(form)

		form = huh.NewForm(
			// on new file
			huh.NewGroup(
				huh.NewInput().
					Description("Enter the name of the environment to use. leave empty to use just '.env'").
					SuggestionsFunc(func() []string {
						a, e := tools.GetAllFiles(c, []string{".env"}, tools.MatchNameContains)
						if e != nil {
							slog.Error("failed to get files: " + e.Error())
							os.Exit(1)
						}
						o := []string{}
						for _, f := range a {
							s, err := tools.ExtractNameFromDotenv(filepath.Base(f))
							if err != nil {
								slog.Error("failed to extract name from dotenv", "error", err)
								os.Exit(1)
							}
							o = append(o, s)
						}
						return o
					}, newFileName).
					Value(&newFileName).
					Validate(func(input string) error {
						if strings.Contains(input, " ") {
							return fmt.Errorf("environment name cannot contain spaces")
						}
						s, err := tools.ExtractNameFromDotenv(input)
						if err != nil && err != tools.ErrFileNotEnvFile {
							return err
						}
						return FileExists(s)
					}),
				huh.NewNote().DescriptionFunc(func() string {
					s, err := tools.ExtractNameFromDotenv(newFileName)
					if err != nil && err != tools.ErrFileNotEnvFile {
						return err.Error()
					}
					return tuiSelectEnvNote(s)
				}, &newFileName),
			).WithHide(newFileType != "new"),
			// on existing file
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Existing file").
					Description("Existing file").
					OptionsFunc(func() []huh.Option[string] {
						out := []huh.Option[string]{}

						list, err := tools.GetAllFiles(c, []string{".env"}, tools.MatchNameContains)
						if err != nil {
							slog.Error("failed to get files: " + err.Error())
							os.Exit(1)
						}
						// filepath.Rel()
						for _, f := range list {
							if strings.HasSuffix(f, ".polyenv") {
								continue
							}
							relativePath, err := filepath.Rel(c, f)
							if err != nil {
								slog.Error("failed to get relative path: " + err.Error())
								os.Exit(1)
							}
							out = append(out, huh.NewOption(relativePath, f))
						}
						return out
					}, nil).
					Validate(func(s string) error {
						filename := filepath.Base(s)
						env, err := tools.ExtractNameFromDotenv(filename)
						if err != nil {
							return err
						}
						if strings.Contains(env, " ") {
							return fmt.Errorf("environment name cannot contain spaces")
						}
						return FileExists(env)
					}).
					Value(&existingFile),
				huh.NewNote().DescriptionFunc(func() string {
					filter, err := tools.ExtractNameFromDotenv(existingFile)
					if err != nil {
						return err.Error()
					}
					return tuiSelectEnvNote(filter)
				}, &existingFile),
			).WithHide(newFileType != "existing"),
		)
		tui.RunHuh(form)

		var err error
		switch newFileType {
		case "new":
			env, err = tools.ExtractNameFromDotenv(newFileName)
		case "existing":
			env, err = tools.ExtractNameFromDotenv(filepath.Base(existingFile))
		}
		if err != nil {
			slog.Error("failed to extract name from dotenv", "error", err)
			os.Exit(1)
		}
	} else {
		if strings.Contains(env, " ") {
			slog.Error("environment name cannot contain spaces")
			os.Exit(1)
		}
	}

	e = FileExists(env)
	if e != nil {
		slog.Error("environment already exists: " + e.Error())
		os.Exit(1)
	}

	file.Name = env
	file.Path = c
	// p.Save()
	return
}

// helper for TuiNewFile
func tuiSelectEnvNote(env string) string {
	slog.Debug("tui select env note", "env", env)
	cwd, err := tools.GetGitRootOrCwd()
	if err != nil {
		slog.Error("failed to get project root: " + err.Error())
		os.Exit(1)
	}

	allFiles, err := tools.GetAllFiles(cwd, []string{".env"}, tools.MatchNameContains)
	if err != nil {
		slog.Error("failed to get files: " + err.Error())
		os.Exit(1)
	}
	// slog.Debug("all files", "files", len(allFiles))

	matchedFiles := make([]string, 0)
	for _, f := range allFiles {
		fname, err := tools.ExtractNameFromDotenv(filepath.Base(f))
		if err != nil {
			slog.Error("failed to extract name from dotenv", "error", err)
			os.Exit(1)
		}
		cwdpath, err := filepath.Rel(cwd, f)
		if err != nil {
			slog.Error("failed to get relative path: " + err.Error())
			os.Exit(1)
		}
		cwdpath = filepath.ToSlash(cwdpath)
		slog.Debug(cwdpath, "filter", env, "fname", fname)
		if env == fname {
			matchedFiles = append(matchedFiles, cwdpath)
			// o += "\n" + cwdpath
		}
	}
	o := "Will use:\n"
	for _, f := range matchedFiles {
		o += fmt.Sprintf("- %s\n", f)
	}

	if len(matchedFiles) > 0 {
		o += "\n ..or "
	}

	if env == "" {
		o += "any file named '.env'"
	} else {
		files := []string{
			strings.Join([]string{".env", env}, "."),
			strings.Join([]string{env, "env"}, "."),
			strings.Join([]string{".env.secret", env}, "."),
		}

		o += fmt.Sprintf("any file named '%s'", strings.Join(files, ", "))
	}
	return o
}

func (file *File) TuiAddGitIgnore() error {
	if !file.Options.UseDotSecretFileForSecrets {
		return nil
	}

	if !RootIsGitRepo() {
		return nil
	}

	if GitignoreMatchesEnvSecret() {
		return nil
	}

	var addToGitignore bool
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("'.env.secret' not found in .gitignore").
				Description("Do you want me to add this to your .gitignore?").
				Affirmative("Yes, pretty please!").
				Negative("No, I will do it myself").
				Value(&addToGitignore),
		),
	)
	tui.RunHuh(f)

	if !addToGitignore {
		return nil
	}

	root, err := tools.GetGitRootOrCwd()
	if err != nil {
		return err
	}

	//normally os.O_CREATE for opening files, however i want it to error if gitignore is not found..
	openedFile, err := os.OpenFile(filepath.Join(root, ".gitignore"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gitignore: %w", err)
	}

	defer func() {
		e := openedFile.Close()
		if e != nil {
			slog.Error("failed to close file", "error", e)
		}
	}()

	if _, err = openedFile.WriteString("\n**/*env.secret*"); err != nil {
		return err
	}

	return nil
}
