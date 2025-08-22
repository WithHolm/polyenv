# wizard

For now, the wizard is only used on init, and will ask the user a series of questions to get the options for the vault.

The wizard is a simple state machine, that will ask questions until it is done, meaning the vault implementation is controlling the narrative after the uses selects the vault (first question).

The questionaire package uses [huh](https://github.com/charmbracelet/huh) to render the questions, and the vault handler uses the `VaultWizardCard` struct to ask the user a question.


## Interface implementation

wizard is a part of "vault" interface and exposes 6 methods in 2 "flows":
- New
  - `NewWizardWarmup() error`
	- `NewWizardNext() *huh.Form`
	- `NewWizardComplete() map[string]string`
- Update
	- `UpdateWizardWarmup(map[string]string) error`
	- `UpdateWizardNext() *huh.Form`
	- `UpdateWizardComplete() map[string]string`

please note that the update path is not enabled as of now, but will be in the future, using the given functions.
For now, you can have it as a dummy function, and just return nil.



## wizard flow

the flow for new and update is the same, so im going to show the new flow here.

Before the wizard of your specific vault is started it calls `WizardWarmup` on the vault.
At this point you can initiate any functions that are needed (like checking if the correct executable is installed).

if warmup is successful, the wizard will call `WizardNext` to get the next question. this function will be called until you deliver back nil.


what i have done with keyvault so far is to use a switch
``` go
var formGroup int

func (wiz *NewWizard) Next() *huh.Form {
	// automatically increment the formGroup
	defer func() { formGroup++ }()

	switch formGroup {
	case 0:
	case 1:
	case 2:
	}

	return nil
}
```

this way you can easily add new questions to the wizard, and you can also use the formGroup to check if the user has answered all the questions.

if the user answers all the questions, you can call `WizardComplete` to get the final result.



```
