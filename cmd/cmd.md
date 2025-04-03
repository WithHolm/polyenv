# cmd

## init

- choose vault provider
- defer to vault provider to select vault
- select what secrets to pull
- select if you want to save just ref to secret or the secret itself (means you gave to run the `pull` before)
- create local settings file

### azure keyvault

- choose vault
  - tenant
  - subscription
  - vault

ref is `@azure:vaulturl/secret`

### aws secrets manager

i dont use aws, so im not sure how to do this...

- choose vault
  - region
  - account
  - vault

ref is `@aws:region/account/secret`

## pull

output to:

- .env file
- local settings file (az func)
- json string
- pwsh scriptblock? (ie (polyenv pull -o pwsh).invoke() to set env vars)
