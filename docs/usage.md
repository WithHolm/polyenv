# Usage

Polyenv is pretty stright forward to use.

* to init the .env file for syncing with a remote vault, run `Polyenv init`
* to push secrets to a already defined remote vault, run `Polyenv push`
* to pull secrets from a already defined remote vault, run `Polyenv pull`


## Vaults

Polyenv supports the following vaults:

* Azure Key Vault

## commands

### Init
The init command is used to initialize the .env file for syncing with a remote vault.
the resulting config is stored in the .polyenv folder in your home directory.
this file can safetly be synced to a git repo as it contains the reference to the secrets storage, and not the secrets themselves.

if you just use `init` without any arguments, the full wizard will be shown, however can skip all of it by setting `--type` and `--arg`
#### Usage
```
Polyenv init [flags]
```

#### Flags
```
  --p, --path string   path to the .env file to use (default "local.env")
  --type string        the type of vault to use (currently only "keyvault")
  --arg key=value      set an argument for the init command. uses dotenv style values. for example: --arg t=mytenant.
```
