# Environment Export

usage is part of environment handling.

Outputs all known variables in env files for current environment for easy consumption by other tools.

> NOTE:  
> This is under active development and may change in the future as more "targets" are added. if you have feature requests or some way you would like the output to be formatted, please open an issue with an example, and how to test this. and we will work on this with you.

## flags

- [--as](#--as) -> how to format the output.
- [--to](#--to) -> where to output to.

### --as

**usage:** `--as {format}` (think of it as "output as {format}")

if this is defined it will go through a formatter before being presented to the output.
what formatter is available depends on the output you are using.

the currently supported formatters are:

- `json`
  - json object: `{"MY_KEY": "myvalue"}`
- `jsonArr`
  - json array: `[{"key": "MY_KEY", "value": "myvalue"}]`
- `pwsh`
  - outputs values as powershell set-item commands -> `Set-Item 'env:MY_KEY' -Value 'myvalue'`
  - you can pipe this to `iex` to execute it: `polyenv !dev export --as pwsh | iex`
- `stats`
  - Outputs a table of all env variables in the current environment with stats about them.
  - this is a work in progress and will be expanded in the future.
  - good for showing all env variables without disclosing any values.
- `pick`
  - ask you to pick what env variables you want to include in the output.
  - will also ask you to pick what format you want the output to be in (if you have multiple formats defined)
- `dotenv`
  - outputs all env variables as dotenv format.
- `azdevops`
  - outputs all env variables as `##vso[task.setvariable variable=MY_KEY;issecret=false]myvalue` commands.
  - it will also automatically set secret to true if the value is a secret.
- `auto`
  - will select the best format for you based on the writer you are using.

> NOTE:  
> Not all writers support all formats. for example, the `github` writer will only support `dotenv` as this is the only supported format for github.  
> if you try to use a format that is not supported by the writer, you will get an error.

### --to

usage: `--to {writer}` (think of it as "output to {writer}")

where to output to.

the currently supported writers are:

- stdout
  - outputs to stdout.
  - supports all formats.
- github
  - outputs to the github env file.
  - supports `dotenv` format.
- ots
  - outputs to OneTimeSecret.com
  - supports all formats, but defaults to `pick` with `dotenv` format.
  - good for sharing env vars and secrets with external collaborators.
    - generate a link that can be opened only once

> NOTE:  
> Polyenv will log some stuff as "info", like "exporting x", but all logging will happen on stderr and should not inferer with any pipeline data flow. it will also NEVER log values. only names/keys. if you have any issues, please open an issue and we will try to help.