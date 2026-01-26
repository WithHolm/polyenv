package model

// File represents the entire structure of the polyenv.toml configuration file.
type File struct {
	Vaults           map[string]map[string]any        `toml:"vault"`        // Maps [vault.<name>] sections
	EnvironmentProfiles map[string]EnvironmentProfileConfig `toml:"environment"` // Maps [environment.<name>] sections
	Autodetection    map[string]string                `toml:"autodetection"` // Maps [autodetection] section (git context patterns to environment profile names)
	Secrets          map[string]SecretConfig          `toml:"secret"`       // Maps [secret.<name>] sections
}

// EnvironmentProfileConfig defines the configuration for a named environment profile.
// It will be used to unmarshal the [environment.<name>] sections from polyenv.toml.
type EnvironmentProfileConfig struct {
    Vault string `toml:"vault"` // The default vault to use for this environment profile
    // Add other environment-specific settings here as needed, e.g., default_source_env_file.
}

// SecretConfig defines how a specific secret variable is mapped to a vault.
// It will be used to unmarshal the [secret.<name>] sections from polyenv.toml.
type SecretConfig struct {
	RemoteKey string `toml:"remote_key"` // The name of the secret in the remote vault
	Vault     string `toml:"vault"`      // Optional: override the default vault for this specific secret
}