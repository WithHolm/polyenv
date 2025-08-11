# polyenv

![alt](/docs/logo.png)

polyenv is a CLI tool that allows you to manage secrets in your environment and show values from multiple env files

for now only Azure Keyvault is supported, but more will be available in the future.

no hidden solution, no monthlye fees, no subscriptions. just a simple CLI tool that you can use to pull secrets from your already defined

## Features

- Initialize a .polyenv.toml file for config. this can be synced used with your git repo.
- Pull secrets from a selection of remote vaults.
- Load values from multiple env files within the same environment.
- no subscriptions, no hidden solution, no monthly fees. just a simple cli tool.

## Installation

check release page [here](https://github.com/WithHolm/polyenv/releases) to download the application

## Usage

### Commands

#### Initialize a new environment

it will ask you to add vaults and secrets:

- `polyenv init`: Asks you to select vault and secret
- `polyenv init --type {vaultType}`: Initializes the environment with the given [vault type](#supported-vaults)
- `polyenv init --type {vaultType} --arg key=value`: Initializes the environment with the given [vault type](#supported-vaults) and sets the given arguments set dotenv style

in all cases it will create a `{env}.polyenv.toml` file in the current directory. this file can be moved anywhere within your repo.

![init](/docs/demos/init.gif)

#### Add vault or secret

- `polyenv !{env} add vault`: Adds a new vault to the environment
- `polyenv !{env} add secret [vault name]`: Adds a new secret to the environment

#### Pull secrets from vault

depending on your [config](#polyenv-config), this will either set secrets in `.env.secrets.{env}` file or existing uinqe keys in existing `{env}.env||.env.{env}` files

- `polyenv !{env} pull`

### Supported vaults

#### Azure Key Vault

The vault uses Azure SDK's [credential chains](https://learn.microsoft.com/en-us/dotnet/azure/sdk/authentication/credential-chains?tabs=dac) for authentication. Either uses your local az cli credentials or any service principal credentials defined in env variables

|argument|alias|description|
|---|---|---|
|`tenant`||tenant id or domain|
|`subscription`|`sub`|subscription id or name|

example:

``` text
polyenv init --type keyvault --arg tenant=mytenant.com --arg subscription=mysubscription
```

#### Dev Vault

``` text
polyenv init --type devvault --arg store=mystore
```

### Options

- `--debug`: Enable debug mode
- `--disable-truncate-debug`: Disables truncating debug logging
  - some debug logs from external providers may be overly verbose so vault implementaiton may tuncate log message. this flag will disable that. use if you want to see the full log message.

## Polyenv Config

- **hypens to underscores**
  - when selecting new secrets, it will replace hyphens with underscores when selecting new name. it makes it easier to define new secrets.
- **uppercase locally**
  - when selecting new secrets, it will convert the name to uppercase. this makes it easier to define new secrets.
- **use dot secret file for secrets**
  - will save any secrets to `.env.secret.{env}` file instead of `.env` file. makes it easier to git ignore secrets.

## Developer Information

### Project Structure

- `cmd/`: Contains the Cobra command implementations
- `internal/`: Internal packages
  - `tools/`: Utility functions
  - `vaults/`: Vault implementations (currently Azure Key Vault)
- `main.go`: Entry point of the application

### Adding New Vaults

To add support for a new vault type:

1. Create a new package under `internal/vaults/`
2. Implement the `Vault` interface defined in `internal/vaults/repository.go`
3. Update the `NewInitVault` function in `repository.go` to include the new vault type

### Key Files

- `cmd/init.go`: Handles the initialization wizard
- `cmd/push.go`: Implements the push command
- `cmd/pull.go`: Implements the pull command
- `internal/vaults/keyvault/wizard.go`: Contains the Azure Key Vault-specific wizard implementation
- `internal/vaults/keyvault.go`: Implements the `Vault` interface for Azure Key Vault

### Environment Variables

The tool uses Azure SDK's DefaultAzureCredential for authentication. Make sure the appropriate environment variables or configuration files are set up for Azure authentication.

## Contributing

fork the project, make your changes, and submit a pull request.

## License

(Add license information here)

This README provides an overview of the project, its commands, and essential information for developers. You may want to expand on certain sections, such as installation instructions, contribution guidelines, and licensing information, based on your specific project requirements.
