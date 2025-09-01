package model

import "github.com/charmbracelet/huh"

type Vault interface {
	// returns current instance as string for wizards
	String() string
	// returns the display name of the vault
	DisplayName() string

	// Warm the vault connection.
	Warmup() error

	//convert vault to a marshalable map
	Marshal() map[string]any

	//convert vault from marshalable map
	Unmarshal(m map[string]any) error

	//Validate the secret name.
	//used when new secret is made, optionally with a suggestion
	// ValidateSecretName(name string) (string, error)

	// return true if you handled the form.
	// false if you want to use the default form
	SecretSelectionHandler(sec *[]Secret) bool

	// list all secrets
	List
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
	WizNext() (*huh.Form, error)
	//completes the wizard. returns the data to be saved. error if something went wrong
	WizComplete() error
}
