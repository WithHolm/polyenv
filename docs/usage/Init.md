# `polyenv init`

The `init` command sets up a new environment for Polyenv to manage.

Running this command will either launch an interactive wizard to guide you through setup or perform a quick setup if enough arguments are provided.

## Usage

```shell
polyenv init [environment] [--vault vaulttype] [--arg key=value]... [flags]
```

## Arguments

-   `[environment]` (optional)
    The name of the environment you want to initialize. If not provided, the wizard will prompt you for a name.

## Flags

| Flag | Shorthand | Description |
| :--- | :--- | :--- |
| `--accept-default-settings` | | Skips the interactive wizard and accepts the default settings for the `polyenv.yaml` file. |
| `--arg` | `-a` | Provides arguments to the vault being initialized (e.g., a name for a Key Vault). This flag can be used multiple times. |
| `--help` | `-h` | Displays help information for the `init` command. |
| `--vault` | | Bypasses the vault selection screen and immediately starts the setup for the specified vault. Valid options are `[devvault, keyvault, local, none]`. |

### Global Flags

| Flag | Description |
| :--- | :--- |
| `--debug` | Enables debug logging to show verbose output. |
| `--disable-truncate-debug`| Prevents the truncation of long values in debug logs. |

## Examples

### Interactive Initialization

Run the command without any arguments to launch the full interactive wizard.

```shell
polyenv init
```

### Quick Initialization with a Local Vault

Quickly set up a new environment using a `local` vault, which is useful for getting started.

```shell
polyenv init my-dev-env --vault local
```

### Initializing with an Azure Key Vault

Initialize an environment using an Azure Key Vault, providing the vault name as an argument.

```shell
polyenv init production --vault keyvault --arg name=my-prod-keyvault
```
