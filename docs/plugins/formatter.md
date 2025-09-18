# Using Formatters

A Formatter in Polyenv transforms your secrets into a specific text format. This is most commonly used with the `polyenv !env export` command to control how the secrets are printed to the screen.

You can specify a formatter using the `--as` flag.

```shell
polyenv !{env} export --as <formatter_name>
```

Below is a list of the available formatters and what they do.

---

## `dotenv`

It produces the standard `KEY=VALUE` format used by `.env` files. This is useful for redirecting output to a file.

**Example Output:**
```dotenv
API_KEY="secret-value"
DATABASE_URL="postgres://user:pass@host/db"
```

**Usage:**
```shell
# Export to a .env file
polyenv !{env} export --as dotenv 
```

---

## `posix` (or `bash`)

This formatter produces output that can be directly evaluated by POSIX-compliant shells like `bash` and `zsh` to load secrets into the current session's environment.

**Example Output:**
```shell
export API_KEY="secret-value"
export DATABASE_URL="postgres://user:pass@host/db"
```

**Usage:**
```shell
# Load secrets directly into your bash or zsh shell
eval "$(polyenv !{env} export --as posix)"
```

---

## `pwsh`

This formatter produces output that can be directly evaluated by **PowerShell** to load secrets into the current session's environment.

**Example Output:**
```powershell
Set-Item "env:API_KEY" -value "secret-value"
Set-Item "env:DATABASE_URL" -value "postgres://user:pass@host/db"
```

**Usage:**
```powershell
# Load secrets directly into your PowerShell session
polyenv !{env} export --as pwsh | iex
```

---

## `json`

This formatter outputs the secrets as a single JSON object, where the secret names are the keys.

**Example Output:**
```json
{
  "API_KEY": "secret-value",
  "DATABASE_URL": "postgres://user:pass@host/db"
}
```

**Usage:**
```shell
polyenv !{env} export --as json
```

---

## `jsonArr`

This formatter outputs the secrets as a JSON array of objects, with each object containing a `key` and `value`.

**Example Output:**
```json
[
  {
    "key": "API_KEY",
    "value": "secret-value"
  },
  {
    "key": "DATABASE_URL",
    "value": "postgres://user:pass@host/db"
  }
]
```

**Usage:**
```shell
polyenv !{env} export --as jsonArr
```

---

## `azdevops`

This formatter is specifically for use in **Azure DevOps Pipelines**. It creates logging commands that set pipeline variables.

please note that polyenv in most cases will detect if the value you are trying to use is a secret or not. this will either be by checking if the source is a vault, the entropy of the string, regex checks og by checking the key names. if it detects a secret it will set the variable as a secret.
this is mainly to protect you from accidentally logging secrets to the pipeline even if you are not intending to (like not using a vault... tsk tsk). i may add a flag to disable this in the future.

**Example Output:**
```
##vso[task.setvariable variable=API_KEY;issecret=true]secret-value
##vso[task.setvariable variable=DATABASE_URL;issecret=true]postgres://user:pass@host/db
```

**Usage (in an Azure DevOps Pipeline script):**
```yaml
- bash: |
    polyenv !{env} export --as azdevops
  displayName: 'Load Secrets into Pipeline'
```

---

## `pick`

This is an interactive formatter. Instead of printing all secrets, it will present you with a list of the secrets, allowing you to choose one to print to the screen. This is useful if you only need to quickly grab a single value.

**Usage:**
```shell
polyenv !{env} export --as pick
```

---

## `stats`

This formatter doesn't output the secrets themselves. Instead, it provides metadata and statistics about the env values being processed, such as if the env should be concidered a secret or if they come from a vault.

**Usage:**
```shell
polyenv !{env} export --as stats
```