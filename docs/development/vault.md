# Creating a New Vault

Polyenv is designed to be extensible. You can easily add support for new secret backends by creating a "Vault".

A Vault is a source or destination for secrets. To create a new one, you need to implement the `Vault` interface and register it.

### 1. The Vault Interface

Your new vault struct must implement the following interface, defined in `internal/model/vault.go`:

the interface may be broken down at a later date so you have more control if you only support "pull" og "push", but for "main control" right now, you need to implement all methods.

you have methods for converting to and from `map[str]any` (`Marshal` and `Unmarshal`), but i also recomend you to [implement toml marshalling for ease of use](https://github.com/BurntSushi/toml#using-the-marshaler-and-encodingtextunmarshaler-interfaces). this may be a requirement later on but for now you can use it to help yourself marshalling and unmarshalling.

#### 1.1 basics

The actual struct can use basic TOML for property tagging for esier marshalling, but the interface is designed to be extensible.

* `String` -> returns the name of the current instance (ie "TENANTID/KEYVAULTNAME")
* `DisplayName` -> returns the display name of the current instance (ie "Azure Key Vault")
* `Warmup` -> warms up the vault connection. this is always called after Unmarshal, and before any other method is called.
* `Marshal` -> converts the vault to a marshalable map
* `Unmarshal` -> converts the vault from a marshalable map
* `SecretSelectionHandler` -> when selecting "new secrets" normally i would use vault.list to get all secrets, and then ask the user to select one. but if you have your own way of selecting secrets (like local where user needs to provide a secret if it cannot be found), you can use this method to handle that. if you return false, the default selection form will be used (ie "i didn't handle this, you do it").

#### 1.2 List/Pull/Push

all of these are basically the same general idea but different commands and outputs

* `*Elevate` -> elevate persmission to list/pull/push secrets if needed. you can also just return `nil` if you dont need elevation. these will be called before the actual command is called.
* `Pull/Push/List` -> the actual command and will return results where needed.

#### 1.3 Wizard

the wizard is a state machine that will ask questions until it is done, meaning the vault implementation is controlling the narrative after the uses selects the vault (first question).

* `WizardWarmup` -> called before the wizard is started. this is where you can initiate any functions that are needed (like checking if the correct executable is installed). the input is also a direct conversion from input arguments defined by [Init](../usage/Init.md)
* `WizardNext` -> called to get the next question. this function will be called until you deliver back nil. check keyvault of local for an example.
* `WizardComplete` -> called when the wizard is done. this is where you can "set things in stone" and return any errors if you failed to complete the setup for some reason.

### 2. Create Your Vault File

Create a new directory and file for your vault, for example: `internal/vaults/myvault/myvault.go`.

for easier searchability, usually the main file will have the same name as the package.
depending on the size of your implementation you can add any other files you need (see keyvault for a "bad" example.. i know its not the best example, but it shows the structure)

### 3. Register Your Vault

Finally, register your new vault in `internal/vaults/registry.go` so that Polyenv knows it exists.

Add your vault to the `vaultRegistry` map:

```go
// internal/vaults/registry.go
import (
    // ... other imports
    "github.com/withholm/polyenv/internal/vaults/myvault"
)

var reg = map[string]func() model.Vault{
  "keyvault": func() model.Vault { return &keyvault.Client{} },
  "local":    func() model.Vault { return &local.Client{} },
  "myvault":  func() model.Vault { return &myvault.Client{} }, // <-- add this
}
```

### 4. Final Flow

1. your vault imports model/secret and model/vault, and implements the `Vault` interface (along with any tui and tools).
1. registry imports your vault
1. cli/userspace imports registry and calls `NewVaultInstance` to get a new instance of your vault with all the bells and whistles.

ive tried as well as i can to make available as much functionality as possible to you while keeping the import flow as strightforward as possible to avoid circular dependencies.
If you experience any issues, please open an issue.
