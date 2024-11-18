# wizard

the wizard is used in the init stage, to ask the user for the information needed to connect to the vault storage.

## wizard flow

Before the wizard of your specific vault is started it calls `WizardWarmup` on the vault.  
This is used to start any goroutines needed for your vault to connect to and get as much info as possible before the wizard begins. you can of course grab more info after the wizard has started, but it often leads to longer wait times for the user.

right after we will ask for `WizardNext` to be called.  
this will return a single `VaultWizardCard` object, this is where the used is asked a question, with title, selections, and a callback function (what happens after the user selects an answer).

State for this (what question you are on), is handled in the `WizardNext` by the vault handler. it will return questions until its done (by returning an empty `VaultWizardCard` object).

``` golang
//example of a wizard card
VaultWizardCard{
    Title: "My Question?",
    Questions: []VaultWizardSelection{
        {
            Key:         "answerOne",
            Description: "First answer",
        },
        {
            Key:         "answerTwo",
            Description: "Second answer",
        },
    },
    Callback: func(s string) error {
        vault.wiz.question = s
        return nil
    },
}
```

when the wizard is done, the vault handler will call `WizardComplete` to get the options for the vault. these options will be written to the .env.vaultopts file.
