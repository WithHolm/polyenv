# env

usage is part of environment handling.

Outputs all known variables in env files for current environment for easy consumption by other tools.

> NOTE:  
> This is under active development and may change in the future as more "targets" are added. if you have feature requests or some way you would like the output to be formatted, please open an issue with an example, and how to test this. and we will work on this with you.

## flags

- `-o, --output`: the output format.

## Outputs

all actual outputs are written to stdout. while logging is written to stderr.
most terminalt will automatically not process messages set to stderr. so you can use this to see what is happening.
If you have any issues, please open an issue and we will try to help.

### json

outputs all env variables as json.

```json
{
  "MY_KEY": "myvalue",
  "MY_OTHER_KEY": "myothervalue"
}
```

### azdevops

outputs all env variables as tasks for azure devops.

```yaml
##vso[task.setvariable variable=MY_KEY;issecret=false]myvalue
##vso[task.setvariable variable=MY_OTHER_KEY;issecret=false]myothervalue
```

### github

outputs all env variables directly to file defined in GITHUB_ENV.

### azas

outputs all env variables as azure azure app service key-value pairs.
> NOTE:  
> only keyvault secret references are supported for this output.

```json
{
  "MY_KEY": "myvalue",
  "MY_OTHER_KEY": "myothervalue",
  "MY_KEY2": "@Microsoft.KeyVault(SecretUri=https://myvault.vault.azure.net/secrets/mykey2)"
}
```

you can then use this as a source for bicep or terraform.

### pwshSb

outputs all env variables as powershell set-item commands.

the command will look like this:

```powershell
set-item 'env:MY_KEY' -Value 'myvalue'
```

you can use it like this:

```powershell
polyenv !dev env -o pwshSb | iex
```


