# polyenv

polyenv is a CLI tool that allows you to to grab secrets directly from your enterprise vaults and either save them to local .env file or output them to terminal.
for now only Azure Keyvault is supported, but more will be available in the future.

## Features

- Initialize a .polyenv file for syncing. this can be used with your git repo
- Push secrets from .env file to your remote vault
- Pull secrets from remote vault to .env file

## Installation

check release page [here](https://github.com/WithHolm/polyenv/releases) to download the application

## Usage

### Commands

1. Initialize a .env file for syncing:

```
polyenv init [--path <path-to-env-file>]
```

2. Push secrets to vault:

```
polyenv push [--path <path-to-env-file>]
```

3. Pull secrets from vault:

```
polyenv pull [--path <path-to-env-file>]
```

### Options

- `--path, -p`: Specify the path to the .env file (default: "local.env")
- `--debug`: Enable debug mode

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

(Add contribution guidelines here)

## License

(Add license information here)

This README provides an overview of the project, its commands, and essential information for developers. You may want to expand on certain sections, such as installation instructions, contribution guidelines, and licensing information, based on your specific project requirements.
