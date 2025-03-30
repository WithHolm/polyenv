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

#### Usage
```
Polyenv init [flags]
```

#### Flags
```
  -h, --help   help for init