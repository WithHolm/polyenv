

## Init

supported arguments:

- 'tenant','t': the tenant id
- 'subscription','sub','s': the subscription id
- 'name': the keyvault name
- 'keys': the keyvault keys to pull
- 'ignore': the content types to ignore
- 'default' (bool): whether to use the default values for tag, expiration, hyphen, and uppercase
- 'tag': the tag to use for the env name
- 'exp': the expiration to append to the secret. uses iso 8601 duration
- 'hyphen' (bool): whether to replace hyphens with underscores
- 'uppercase' (bool): whether to automatically uppercase the env name and lowercase the keyvault name


all init arguments are set by defining `--arg key=value` on init. follows normal dotenv values for example:
``` bash
polyenv init --type keyvault --arg t=mytenant --arg s=mysubscription --arg name=mykeyvault --arg keys=mykey1,mykey2
```

if all basic
