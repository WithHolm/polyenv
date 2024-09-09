# dotenv-keyvault
implementaion of dotenv vault using keyvault

A version of dotenv that uses keyvault as 'vault' instead of the dotenv projects default one. 
Requires the user to have active access to the specified keyvault when this command is run".

uses godotenv for doenv file handling.


`app init` will initialize the .env file for syncing with your enterprise-vault.

* if no arument is provided it will take you through a wizard to create a new vault

`app pull` will pull all secrets from keyvault and add them to the .env file. 

* `-o, --out string` here to post the results of the pull. 'env' for directly to env variables, 'file' for .env file (default "env")
* `-p, --path string` path to the '.env' file to pull. appends '.vaultopts' when searching. Uses /.env by default (default ".env")

`app push` will push the .env file to keyvault.

* `-p, --path string` path to the .env file to push. uses /.env by default (default ".env")
* `-t, --tenant string` tenant for the keyvault. only needed first time.
* `-v, --vaultName string` name of the keyvault to push to. only needed first time.
* `--vaultType string` type of vault. only keyvault is supported at the moment (default "keyvault")