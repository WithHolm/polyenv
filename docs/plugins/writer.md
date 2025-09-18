# Using Writers

A Writer in Polyenv determines the **destination** of your formatted secrets. While the default is to simply print to the screen, you can use writers to send secrets to other places, like GitHub Actions outputs or a one-time secret sharing service.

You can specify a writer using the `--to` flag.

```shell
polyenv export --to <writer_name>
```

Below is a list of the available writers and what they do.

---

## `stdout`

This is the **default** writer. It simply prints the formatted output directly to your terminal (standard output). It's the writer you'll use most often for local development, for loading environment variables, or for redirecting secrets to a file.

**Usage:**
```shell
# Print secrets to the screen (implicitly uses --to stdout)
polyenv export

# Load secrets into your shell
eval "$(polyenv export --as posix --to stdout)"
```

---

## `github-env`

This writer is for use inside **GitHub Actions**. It appends the formatted secrets to the `$GITHUB_ENV` file, which is the standard method for setting environment variables that will be available to all subsequent steps in the same job.

This writer only accepts the `dotenv` format, which it uses by default.

**Usage (in a GitHub Actions workflow):**
```yaml
- name: Load Secrets into Job Environment
  run: polyenv export --to github-env

- name: Use the Secrets
  run: |
    echo "The secret API key is $API_KEY"
```

---

## `github-out`

This writer is also for use inside **GitHub Actions**. It sets the formatted secrets as **step outputs**. This is useful if you need to pass secrets to a different job or make them available in the context of a specific step.

**Usage (in a GitHub Actions workflow):**
```yaml
- name: Load Secrets as Step Outputs
  id: load_secrets
  run: polyenv export --to github-out

- name: Use an Output from a Previous Step
  run: |
    echo "The secret API key is ${{ steps.load_secrets.outputs.API_KEY }}"
```

---

## `ots` (One-Time Secret)

This writer sends your formatted secrets to a one-time secret sharing service and provides you with a secret link. **This link can only be visited once**, after which the secret is permanently deleted. This is a secure way to share sensitive information with another person.

By default, this writer uses the `pick` formatter so you can interactively choose which secret to share.

to protect against automated attacks this will always go through a TUI form

**Usage:**
```shell
# This will prompt you to pick a secret and then return a one-time URL
polyenv export --to ots

# You can also share multiple secrets by specifying a different formatter
polyenv export --to ots --as dotenv
```
