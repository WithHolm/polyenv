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

``` text
#mysecret.polyenv
[vault.myvault]
type = "keyvault"
tenant = "my-tenant"
uri = "myvault.vault.azure.net"

[vault.myvault2]
type = "keyvault"
tenant = "my-tenant"
uri = "myvault2.vault.azure.net"

[vault.myvault3]
type = "keyvault"
tenant = "my-tenant"
uri = "myvault3.vault.azure.net"

[options]
replaceHyphen = true
autoUppercase = true
ignoreContentType = ["application/x-pkcs12"]

[secret.localname]
vault = "myvault"
key = "myVaultSecret"

[secret.localname2]
vault = "myvault2"
key = "myVaultSecret2"
```

``` text
#mysecret.env
LOCALNAME=value1
LOCALNAME2=value2
```

what happens if you have a secret in `.env` file and want to push it to a vault or you register a new secret?

* connect secret to vault (list of available vaults)
  * select one or register a new one
* set `remote-name`