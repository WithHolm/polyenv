soo.. teh file format thing:

my first thought was to have a polyenv file that just said "this is where you get your secrets from". and than every secret in the vault would be copied down to a separate file.

``` text
# mysecret.polyenv
Tenant: my-tenant
Uri:myvault.vault.azure.net
``

``` text
# mysecret.env
KEY1=value1
KEY2=value2
KEY3=value3
```

this might work, but its also a bit pain in the ass to maintain:

* what if you want a different name locally than whats in the vault?
* what if you need items from different vaults?
* what if you only need to touch a single secret in env file?

so what is needed?

* a way to specify multiple vaults
* a way to reference a "local secret" to a vault (referenced in previous point)

so.. toml could be best option here. it allows for comments and is "structured" while seeming pretty open.

``` toml
#mysecret.polyenv
[vault.myvault]
type = "keyvault"
tenant = "my-tenant"
uri = "myvault.vault.azure.net"

[vault.myvault2]
type = "keyvault"
tenant = "my-tenant"
uri = "myvault2.vault.azure.net"

# Example for AWS Secrets Manager.
# Note the vault-specific 'region' field, which is required for AWS.
#please not that there isnt a aws vault type yet..
[vault.aws-myvault]
type = "aws-secrets-manager"
region = "us-east-1"
# The optional 'path_prefix' will be prepended to the remote_key of any
# secret that uses this vault definition.
path_prefix = "production/my-app/"

[options]
hyphens_to_underscores = true
uppercase_locally = true

[secret.keyvault-secret]
vault = "myvault"
remote_key = "myVaultSecret"

[secret.keyvault-secret2]
vault = "myvault2"
remote_key = "myVaultSecret2"

[secret.my-aws-secret]
vault = "aws-myvault"
remote_key = "credentials/my-aws-secret"

[secret.my-aws-secret2]
vault = "aws-myvault"
```

``` dotenv
#mysecret.env
LOCALNAME=value1
LOCALNAME2=value2
```

what happens if you have a secret in `.env` file and want to push it to a vault or you register a new secret?

* select secret from `.env` file (`huh` list of "un synced" secrets if none is defined when starting app)
* connect secret to vault (`huh` list of available vaults)
  * select one or register a new one
* set `remote-name` 

## Options:

* hyphens_to_underscores
  * convert hyphens to underscores in the local environment
  * this will convert secret keys in polyenv to doenv -> `secret.my-secret` -> `my_secret`
  * default: `true`
* uppercase_locally
  * convert all keys in the local environment to uppercase
  * this will convert secret keys in polyenv to doenv -> `secret.my-secret` -> `MY-SECRET`
  * default: `true`