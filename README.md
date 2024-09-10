# dotenv-myvault

dotenv-myvault is a CLI tool that allows you to sync your .env files with alternatives to dotenv vault.
for now only Azure Keyvault is supported, but more will be available in the future.

## Features

- Initialize .env file for syncing with Azure Key Vault
- Push secrets from .env file to Azure Key Vault
- Pull secrets from Azure Key Vault to .env file
- Support for multiple tenants and subscriptions
- Interactive wizard for easy setup

## Installation

check release page [here](https://github.com/WithHolm/dotenv-myvault/releases) and download the goddamn exe

## Usage

### Commands

1. Initialize a .env file for syncing:

```
dotenv-myvault init [--path <path-to-env-file>]
```

2. Push secrets to Azure Key Vault:

```
dotenv-myvault push [--path <path-to-env-file>]
```

3. Pull secrets from Azure Key Vault:

```
dotenv-myvault pull [--path <path-to-env-file>]
```

### Options

- `--path, -p`: Specify the path to the .env file (default: ".env")

## Developer Information

### Project Structure

- `cmd/`: Contains the Cobra command implementations
- `internal/`: Internal packages
  - `charmselect/`: Custom implementation for interactive selection
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

(Add contribution guidelines here)

## License

(Add license information here)

This README provides an overview of the project, its commands, and essential information for developers. You may want to expand on certain sections, such as installation instructions, contribution guidelines, and licensing information, based on your specific project requirements.