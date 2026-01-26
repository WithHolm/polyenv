package polyenvfile

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tui"
)

// select what kind of secret to add
func tuiSelectAddType(conf model.VaultAddConfig) (string, error) {
	opts := make([]huh.Option[string], 0)
	if conf.CanAddExisting {
		opts = append(opts, huh.NewOption("existing", "existing"))
	}
	if conf.CanAddNew {
		opts = append(opts, huh.NewOption("new", "new"))
	}
	var addSecret string
	if len(opts) > 1 {
		f := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Add secret").
					Description("How do you want to add the secret?").
					Options(opts...).
					Value(&addSecret),
			),
		)
		tui.RunHuh(f)
	} else if len(opts) == 0 {
		return "", fmt.Errorf("vault does not support adding secrets")
	} else {
		slog.Debug("only one option, using that", "option", opts[0].Value)
		addSecret = opts[0].Value
	}
	return addSecret, nil
}

// handler for adding existing secrets
func tuiAddExistingSecret(v model.Vault, file *File, vaultName string) ([]model.Secret, error) {
	var selectedSecrets []model.Secret
	err := v.ListElevate()
	if err != nil {
		return selectedSecrets, fmt.Errorf("failed to elevate permissions: %w", err)
	}

	f := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[model.Secret]().
				Title("Select secret(s)").
				Description("Multiple secrets can be selected. secrets with '!' are not enabled.").
				OptionsFunc(func() (opt []huh.Option[model.Secret]) {
					// list items
					list, e := v.List()
					if e != nil {
						slog.Error("failed to list secrets: " + e.Error())
						os.Exit(1)
					}

					pre := make([]huh.Option[model.Secret], 0)
					for _, secret := range list {
						//check if it has a local secret
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
	return selectedSecrets, nil
}

// GenerateRandomString securely generates a random string of a specific byte length
// using crypto/rand and encodes it to Base64.
func GenerateRandomString(length int) (string, error) {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!?1234567890")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b), nil
}

// add a single new secret
func tuiAddNewSecret(v model.Vault, file *File, vaultName string) ([]model.Secret, error) {
	var selectedSecrets []model.Secret
	var addAnother bool
	for {
		key := ""
		val := ""

		suggestions := make([]string, 5)
		var err error
		for i := 0; i < 5; i++ {
			suggestions[i], err = GenerateRandomString(10)
			if err != nil {
				return nil, err
			}
		}

		f := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Secret name").
					Description("Enter the key of the secret to use").
					CharLimit(512).
					Validate(v.ValidateSecretName).Value(&key),
				huh.NewInput().
					Title("Secret Value").
					Description("Enter the value of the secret").
					EchoMode(huh.EchoModePassword).
					Suggestions(suggestions).
					Value(&val),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add another?").
					Description("Do you want to add another secret?").
					Affirmative("Yes").
					Negative("No").
					Value(&addAnother),
			),
		)

		tui.RunHuh(f)

		//convert to byte
		sec := model.NewSecretValue(val)
		// destroy original string as quickly as possible
		val = ""

		v.Push(model.SecretContent{
			ContentType: "Secret",
			Value:       sec,
			RemoteKey:   key,
			LocalKey:    "",
		})

		selectedSecrets = append(selectedSecrets, model.Secret{
			Vault:       vaultName,
			ContentType: "Secret",
			Enabled:     true,
			RemoteKey:   key,
			LocalKey:    "",
		})
		if !addAnother {
			break
		}
	}

	return selectedSecrets, nil
}

// handler for adding new secrets
func (file *File) TuiAddSecret(vaultName string) error {
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
	conf := v.GetAddConfig()
	addSecret, err := tuiSelectAddType(conf)
	if err != nil {
		return err
	}

	sec := make([]model.Secret, 0)
	switch addSecret {
	case "existing":
		sec, err = tuiAddExistingSecret(v, file, vaultName)
	case "new":
		sec, err = tuiAddNewSecret(v, file, vaultName)
	default:
		return fmt.Errorf("unknown add secret type: %s", addSecret)
	}

	if err != nil {
		return err
	}

	//init secrets if not already done
	if file.Secrets == nil {
		file.Secrets = make(map[string]model.Secret)
	}

	for _, sec := range sec {
		localSecret, hasLocalSecret := file.GetSecretInfo(sec.RemoteKey, vaultName)
		var displayname string

		description := "select name to use when referencing in env? Enter will use the remote name"
		if hasLocalSecret {
			description = fmt.Sprintf("do you want to change the local name? Enter will use the current name: %s", localSecret.LocalKey)
		}

		placeholder := sec.RemoteKey
		if hasLocalSecret {
			placeholder = localSecret.LocalKey
		}

		f := huh.NewForm(
			//set local name for the remote secret
			huh.NewGroup(
				huh.NewInput().
					Title(sec.RemoteKey).
					Description(description).
					Placeholder(placeholder).
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
						return file.Options.ConvertString(sec.RemoteKey)
					}
					return file.Options.ConvertString(displayname)
				}, &displayname),
			),
		)
		tui.RunHuh(f)

		if displayname == "" && hasLocalSecret {
			displayname = localSecret.LocalKey
		} else if displayname == "" {
			displayname = sec.RemoteKey
		}
		sec.Vault = vaultName
		displayname = file.Options.ConvertString(displayname)

		if hasLocalSecret {
			delete(file.Secrets, localSecret.LocalKey)
		}
		sec.LocalKey = displayname
		file.Secrets[displayname] = sec
	}

	err = file.Save()
	if err != nil {
		return err
	}

	return nil
}
