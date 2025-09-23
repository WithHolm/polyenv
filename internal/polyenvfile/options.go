package polyenvfile

import (
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	List "github.com/charmbracelet/lipgloss/list"
)

type VaultOptions struct {
	HyphenToUnderscore         bool `toml:"hyphens_to_underscores"`
	UppercaseLocally           bool `toml:"uppercase_locally"`
	UseDotSecretFileForSecrets bool `toml:"use_dot_secret_file_for_secrets"`
}

type VaultOptionHelper struct {
	Name    string
	Value   any
	Summary string
}

// converts string using rules set by options
func (opt VaultOptions) ConvertString(s string) string {
	if opt.UppercaseLocally && strings.ToUpper(s) != s {
		// slog.Debug("converting to uppercase", "string", s)
		s = strings.ToUpper(s)
	}
	if opt.HyphenToUnderscore && strings.Contains(s, "-") {
		// slog.Debug("converting to underscore", "string", s)
		s = strings.ReplaceAll(s, "-", "_")
	}
	return s
}

func (opt VaultOptions) GetVaultOptionHelper() map[string]VaultOptionHelper {
	return map[string]VaultOptionHelper{
		"underscoreLocally": {
			Name:    "underscore locally",
			Value:   opt.HyphenToUnderscore,
			Summary: "When adding new secrets: Convert hyphens to underscores\nWhen creating new secrets: Convert underscores to hyphens remotely\nEx: my-secret -> my_secret",
		},
		"uppercaseLocally": {
			Name:    "uppercase locally",
			Value:   opt.UppercaseLocally,
			Summary: "When adding new secrets: Uppercase names\nWhen creating new secrets: Lowercase names remotely\nEx: my-secret -> MY-SECRET",
		},
		"dotEnvSecrets": {
			Name:  "use dot secret file for secrets",
			Value: opt.UseDotSecretFileForSecrets,
			Summary: strings.Join(
				[]string{
					"Any secrets you pull will be stored in a .env.secret.<env> file instead of .env file.",
					"This makes it easier to add them to gitignore. while being accessible with any .env import system",
					"",
					"I HIGLY SUGGEST YOU USE THIS",
				},
				"\n",
			),
		},
	}
}

func toPrettyBool(b bool) string {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39"))
	if b {
		return green.Render("Yes")
	}
	return red.Render("No")
}

func (opt VaultOptions) ListCurrentOptions() string {
	items := opt.GetVaultOptionHelper()
	currentOptList := List.New()
	bold := lipgloss.NewStyle().Bold(true)

	currentOptList.Items(
		items["underscoreLocally"].Name, List.New(
			bold.Render(toPrettyBool(opt.HyphenToUnderscore)),
		),
		items["uppercaseLocally"].Name, List.New(
			bold.Render(toPrettyBool(opt.UppercaseLocally)),
		),
		items["dotEnvSecrets"].Name, List.New(
			bold.Render(toPrettyBool(opt.UseDotSecretFileForSecrets)),
		),
	)
	return currentOptList.String()
}

func (opt VaultOptions) TuiOpts() *huh.Form {
	optMap := opt.GetVaultOptionHelper()
	// truefalse := []huh.Option[bool]{
	return huh.NewForm(

		huh.NewGroup(
			huh.NewConfirm().
				Description(optMap["underscoreLocally"].Summary).
				Title(optMap["underscoreLocally"].Name).
				Affirmative("Yes").
				Negative("No").
				Value(&opt.HyphenToUnderscore).WithButtonAlignment(lipgloss.Left),
			huh.NewConfirm().
				Description(optMap["uppercaseLocally"].Summary).
				Title(optMap["uppercaseLocally"].Name).
				Affirmative("Yes").
				Negative("No").
				Value(&opt.UppercaseLocally).WithButtonAlignment(lipgloss.Left),
			huh.NewConfirm().
				Description(optMap["dotEnvSecrets"].Summary).
				Title(optMap["dotEnvSecrets"].Name).
				Affirmative("Yes").
				Negative("No").
				Value(&opt.UseDotSecretFileForSecrets).WithButtonAlignment(lipgloss.Left),
		),
	)
}
