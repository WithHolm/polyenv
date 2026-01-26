# Polyenv Feature Development Ideas

This document summarizes the proposed features and architectural changes for the `polyenv` tool, based on our recent discussions.

## 1. Unified & Context-Aware Environment Management

This proposal overhauls how environments are managed, moving from a dynamic command model to a more standard, unified, and intelligent system.

### Core Idea

The fundamental change is to **eliminate dynamic `!{env}` commands** and multiple `polyenv.*.toml` files in favor of a **single `polyenv.toml` configuration file** and a standard, flag-based CLI interaction. This system will be enhanced with intelligent, Git-based context detection to automate environment selection.

### Proposed `polyenv.toml` Structure

A single `polyenv.toml` will contain all configuration, structured into clear sections:

```toml
# --- 1. Define Your Vaults ---
# All available vaults are defined at the top level.
[vault.local_dev]
type = "local"

[vault.preview_ots]
# OneTimeSecret for ephemeral PR environments (requires manual TUI interaction)
type = "onetimesecret"

[vault.production_keyvault]
type = "keyvault"
url = "https://my-prod-app.vault.azure.net/"


# --- 2. Define Named Environment Profiles ---
# This section defines an "environment profile" and which vault it uses.
[environment.development]
vault = "local_dev"

[environment.preview]
vault = "preview_ots"

[environment.production]
vault = "production_keyvault"


# --- 3. Git Context Autodetection ---
# Maps git context to the named environments above.
# The tool checks this section from top to bottom and uses the first match.
[autodetection]
"ref:refs/heads/main"      = "production"
"ref:refs/heads/release/*" = "production" # For release branches
"pr:*"                     = "preview"    # For any pull request
"ref:refs/heads/*"         = "development"  # Fallback for any other branch

# --- 4. Secret Mappings ---
# Maps variables to their corresponding secret names in a vault.
[secret.DATABASE_URL]
remote_key = "POSTGRES_CONNECTION_STRING"

[secret.STRIPE_API_KEY]
# If remote_key is omitted, it defaults to the variable name itself.
```

### Environment Selection Hierarchy

To ensure predictable behavior, the active environment will be determined using the following order of precedence:

1.  **Explicit Flag (Highest Priority)**: `polyenv pull --env development` will always use the `development` environment.
2.  **System Environment Variable**: If the flag isn't set, the tool will check for `POLYENV_ENVIRONMENT=production`. This is ideal for CI/CD scripts.
3.  **Git Autodetection**: If neither of the above is set, `polyenv` will use the `[autodetection]` rules to determine the environment from the current Git branch or PR status.
4.  **Default Fallback**: If no rules match, it can fall back to a predefined default (e.g., `development`).

---

## 2. Interactive Onboarding via `.env.example`

This feature introduces a new `polyenv init` command to dramatically improve the onboarding experience for new developers.

### Core Idea

The `polyenv init` command will parse a structured `.env.example` file and launch an interactive wizard to guide a developer through setting up their local `.env` file, including the secure handling of secrets.

### Proposed `.env.example` Format

The `.env.example` file will be enhanced with special comments (`# polyenv:`) to provide metadata for the wizard, while keeping the `KEY=DEFAULT_VALUE` format intact.

```dotenv
# .env.example

# --- DATABASE ---
# The main connection string for the PostgreSQL database.
# polyenv:type=sec
DB_CONNECTION_STRING=

# The host where the database is running.
# polyenv:type=str,validate=hostname
DB_HOST=localhost

# The port for the database.
# polyenv:type=int,validate=port
DB_PORT=5432

# --- SITE CONFIGURATION ---
# Set to true to enable the experimental dashboard.
# polyenv:type=bool
ENABLE_DASHBOARD=true

# Email address for the site administrator.
# polyenv:type=str,validate=email
ADMIN_EMAIL=admin@example.com
```

### Wizard Functionality

The `polyenv init` command will launch a TUI wizard with the following features:

*   **Groups**: Comments like `# --- DATABASE ---` will be used as titles for logical groups of questions in the UI.
*   **Descriptions**: The plain comment line above each variable will serve as the description for the input field.
*   **Type-Specific Inputs**: The `polyenv:type` annotation will determine the input method:
    *   `bool`: A "Yes/No" confirmation.
    *   `int`: An input field that validates for integers.
    *   `str`: A standard text input field.
    *   `sec`: Triggers special secret-handling logic.
*   **Input Validation**: The `polyenv:validate` annotation will allow for format validation (e.g., `port`, `email`, `hostname`).
*   **Secret Handling**: When `type=sec` is encountered:
    1.  If the secret is already defined in `polyenv.toml`, the wizard will state that it's managed by `polyenv` and skip the question.
    2.  If it's a new secret, the wizard will ask the user if they want to **fetch it from a vault** or **enter the value manually**. If they choose the vault option, it will guide them to update the `polyenv.toml` with the new secret mapping.
*   **Output**: The wizard will generate a `.env.local` file with the user's inputs and, if necessary, an updated `polyenv.toml` file. It will conclude by prompting the user to run `polyenv pull`.