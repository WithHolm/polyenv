package polyenvfile

import "strings"

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

func (opt VaultOptions) GetVaultOptionHelper() []VaultOptionHelper {
	return []VaultOptionHelper{
		{
			Name:    "underscore locally",
			Value:   opt.HyphenToUnderscore,
			Summary: "When adding new secrets: Convert hyphens to underscores\nWhen creating new secrets: Convert underscores to hyphens remotely\nEx: my-secret -> my_secret",
		},
		{
			Name:    "uppercase locally",
			Value:   opt.UppercaseLocally,
			Summary: "When adding new secrets: Uppercase names\nWhen creating new secrets: Lowercase names remotely\nEx: my-secret -> MY-SECRET",
		},
		{
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
