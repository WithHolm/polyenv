package keyvault

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/withholm/polyenv/internal/tools"
)

type (
	WizardType string

	NewWizard struct {
		tenant             string
		subscriptions      []string
		uri                string
		vaultKeys          []string
		ignoreContentTypes []string
		pickVaultKeys      bool

		defineDetails               bool
		detail_envNameTag           string
		detail_pushAppendExpiration string
		detail_replaceHyphen        bool
		detail_autoUppercase        bool
	}

	GraphQueryItem struct {
		Name           string
		ResourceGroup  string
		SubscriptionId string
		TenantId       string
		VaultUri       string
		Location       string
	}
)

const (
	New    WizardType = "new"
	Update WizardType = "update"
)

var formGroup int

func newWizard() NewWizard {
	return NewWizard{
		tenant:        "",
		subscriptions: make([]string, 0),
		uri:           "",
		vaultKeys:     make([]string, 0),
		pickVaultKeys: false,

		defineDetails:               false,
		detail_pushAppendExpiration: "P1Y",
		detail_envNameTag:           "envName",
		detail_replaceHyphen:        true,
		detail_autoUppercase:        true,
	}
}

func (wiz *NewWizard) Warmup() error {
	err := checkAzCliInstalled()
	if err != nil {
		return err
	}
	return nil
}

func (wiz *NewWizard) Complete() map[string]string {
	return map[string]string{
		"VAULT_TYPE":           "keyvault",
		"TENANT":               wiz.tenant,
		"URI":                  wiz.uri,
		"KEYS":                 strings.Join(wiz.vaultKeys, ","),
		"IGNORE_CONTENT_TYPES": strings.Join(wiz.ignoreContentTypes, ","),
		"TAG":                  wiz.detail_envNameTag,
		"APPEND_EXPIRATION":    strings.ToUpper(wiz.detail_pushAppendExpiration),
		"REPLACE_HYPHEN":       fmt.Sprintf("%t", wiz.detail_replaceHyphen),
		"AUTO_UPPERCASE":       fmt.Sprintf("%t", wiz.detail_autoUppercase),
	}
}

func (wiz *NewWizard) Next() *huh.Form {
	// automatically increment the formGroup
	defer func() { formGroup++ }()

	switch formGroup {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select a tenant").
					OptionsFunc(func() []huh.Option[string] {
						// chn := make(chan armsubscriptions.TenantIDDescription)
						v, e := getTenants()
						if e != nil {
							slog.Error("failed to get tenants: " + e.Error())
							os.Exit(1)
						}
						ret := make([]huh.Option[string], 0)
						for _, tenant := range v {
							opt := huh.NewOption(*tenant.DisplayName, *tenant.TenantID)
							ret = append(ret, opt)
						}
						return ret
					}, nil).
					Value(&wiz.tenant),
			),
		)
	case 1:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select subscriptions").
					Description("Multiple subscriptions can be selected.").
					OptionsFunc(func() (ret []huh.Option[string]) {
						subs, err := getSubscriptions(wiz.tenant)
						if err != nil {
							slog.Error("failed to get subscriptions: " + err.Error())
							os.Exit(1)
						}
						for _, sub := range subs {
							opt := huh.NewOption(*sub.DisplayName, *sub.SubscriptionID)
							ret = append(ret, opt)
						}
						return ret
					}, &wiz.tenant).
					Value(&wiz.subscriptions),
				huh.NewSelect[string]().
					Title("Select a keyvault").
					OptionsFunc(func() []huh.Option[string] {
						// chn := make(chan armsubscriptions.TenantIDDescription)
						resources, e := getKeyvaults(wiz.subscriptions, wiz.tenant)
						if e != nil {
							slog.Error("failed to get keyvault resources: " + e.Error())
							os.Exit(1)
						}
						ret := make([]huh.Option[string], 0)
						for _, res := range resources {
							opt := huh.NewOption(res.Name, res.VaultUri)
							ret = append(ret, opt)
						}
						return ret
					}, &wiz.subscriptions).
					Value(&wiz.uri),
			),
			huh.NewGroup(
				huh.NewConfirm().
					Title("do you want to select what secrets to pull or do you want all available secrets?").
					Affirmative("Yes, select").
					Negative("No, all").
					Value(&wiz.pickVaultKeys),
			),
		)
	case 2:
		//either select keys to sync or content types to ignore
		return huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select secrets").
					Description("Multiple secrets can be selected. exclamation mark = not enabled").
					OptionsFunc(func() (opt []huh.Option[string]) {
						list, err := getKeyvaultKeys(wiz.uri, wiz.tenant)
						if err != nil {
							slog.Error("failed to get keyvault keys: " + err.Error())
							os.Exit(1)
						}
						for _, secret := range list {
							enabled := ""
							if !*secret.Attributes.Enabled {
								enabled = "!"
							}
							// "!mySecret (text/plain)" -> Not enabled, with name of 'mySecret' with text/plain as content type
							// mySecret (text/plain) -> enabled, with name of 'mySecret' with text/plain as content type
							s := fmt.Sprintf("%s%s (%s)", enabled, secret.ID.Name(), *secret.ContentType)
							opt = append(opt, huh.NewOption(s, secret.ID.Name()))
						}
						return opt
					}, nil).
					Value(&wiz.vaultKeys),
			).WithHide(!wiz.pickVaultKeys),
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Ignore content types").
					Description("what content types do you want to ignore? dont select anything and press enter to skip this step").
					OptionsFunc(func() (opt []huh.Option[string]) {
						list, err := getKeyvaultKeys(wiz.uri, wiz.tenant)
						if err != nil {
							slog.Error("failed to get keyvault keys: " + err.Error())
							os.Exit(1)
						}

						contentTypes := make([]string, 0)
						for _, secret := range list {
							if slices.Contains(contentTypes, *secret.ContentType) {
								continue
							}
							contentTypes = append(contentTypes, *secret.ContentType)
						}

						for _, contentType := range contentTypes {
							opt = append(opt, huh.NewOption(contentType, contentType))
						}
						return opt
					}, nil).
					Value(&wiz.ignoreContentTypes),
			).WithHide(wiz.pickVaultKeys),
		)
	case 3:
		// l := l
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).MarginRight(1)
		list := list.New(
			"Tag env name in keyvault", list.New(wiz.detail_envNameTag),
			"Append expiration to pushed secrets", list.New(wiz.detail_pushAppendExpiration),
			"switch between dash and underscore (underscore locally, dash on keyvault)", list.New(fmt.Sprintf("%t", wiz.detail_replaceHyphen)),
			"switch between lowercase and uppercase when pushing to keyvault (uppercase locally, lowercase on keyvault)", list.New(fmt.Sprintf("%t", wiz.detail_autoUppercase)),
		).Enumerator(list.Dash).ItemStyle(itemStyle)
		template := []string{
			"these are the current defaults:",
			fmt.Sprint(list),
			"the last 2 options are only applicable when pushing or pulling new secrets",
		}
		//select what tag to use in keyvault
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Defaults").Description(strings.Join(template, "\n")),
				huh.NewConfirm().
					Title("Do you want to change any of the defaults?").
					Affirmative("Yes").
					Negative("No").
					Value(&wiz.defineDetails),
			),
		)
	case 4:
		return huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Keyvault Tags").
					Description("keyvault have a limited character set, so i need to convert any env names. however im storing the original value in a tag on the secret."),
				huh.NewNote().
					Description("this will also have the conseqense that the secret name is delinked from the 'dotenv key name', with the link being this tag."),
				huh.NewInput().
					Description("what tag do you want to use? press enter to use default").
					Title("env name tag").
					Placeholder(wiz.detail_envNameTag).
					CharLimit(512).
					Value(&wiz.detail_envNameTag),
			).WithHide(!wiz.defineDetails),

			huh.NewGroup(
				huh.NewNote().
					Description("ISO 8601 specified duration, written as 'P{number}{unit}', where unit is one of the following: Y (for years), M (for months), W (for weeks). I wont allow for Time units."),
				huh.NewInput().
					Description("what expiration do you want to append to the secret? press enter to use default").
					Title("expiration").
					Placeholder(wiz.detail_pushAppendExpiration).
					CharLimit(10).
					Value(&wiz.detail_pushAppendExpiration).
					Validate(tools.ValidateIsoDate),
			).WithHide(!wiz.defineDetails),

			huh.NewGroup(
				huh.NewConfirm().
					Title("Do you want to use dash in keyvault and underscore in local env?\n helps when pulling new secrets, automatically converting them").
					Affirmative("Yes").
					Negative("No").
					Value(&wiz.detail_replaceHyphen),
				huh.NewConfirm().
					Title("Do you want to automatically uppercase the env name and lowercase the keyvault name?").
					Affirmative("Yes").
					Negative("No").
					Value(&wiz.detail_autoUppercase),
			).WithHide(!wiz.defineDetails),
		)
	}
	return nil
}
