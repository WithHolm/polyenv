package model

import "github.com/charmbracelet/huh"

type Vault interface {
	// returns current instance as string for wizards
	ToString() string
	// returns the display name of the vault
	DisplayName() string

	// list all secrets
	List() ([]Secret, error)

	// Warm the vault connection.
	Warmup() error

	// validate the incoming config
	ValidateConfig(options map[string]any) error

	//Validate the secret name.used when new secret is made, optionally with a suggestion
	ValidateSecretName(name string) (string, error)

	Push
	Pull
	Wizard
}

type List interface {
	//elevate to list secrets (azure pim, aws ssm, etc)
	ListElevate() error
	// list all secrets
	List() ([]Secret, error)
}

type Push interface {
	// elevate permissions to push secrets (azure pim, aws ssm, etc)
	PushElevate() error
	// push content to a secret
	Push(s SecretContent) error
}

type Pull interface {
	//elevate permissions to pull secrets (azure pim, aws ssm, etc)
	PullElevate() error
	// pull content from a secret
	Pull(Secret) (SecretContent, error)
}

type Wizard interface {
	//starts the wizard. same as vault warmup, but you cannot assume to have the any of the correct data
	WizWarmup(map[string]any) error
	//next form in wizard. returns nil if there are no more forms
	WizNext() *huh.Form
	//completes the wizard. returns the data to be saved
	WizComplete() map[string]any
}
