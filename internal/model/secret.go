package model

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

type Secret struct {
	Vault       string `toml:"vault"`
	ContentType string `toml:"content_type"`
	Enabled     bool   `toml:"-"`
	RemoteKey   string `toml:"remote_key"`
	LocalKey    string `toml:"-"`
}

// Used when pushing or pulling secrets
type SecretContent struct {
	ContentType string
	Value       string
	RemoteKey   string
	LocalKey    string
}

// data coming from dotenv file
type StoredEnv struct {
	Value    string
	Key      string
	File     string
	IsSecret bool
}

type InputEnv struct {
	IsMap  bool
	Map    map[string]any
	String string
}

func (s Secret) ToString() string { return fmt.Sprintf("%s (%s)", s.RemoteKey, s.ContentType) }

// Gets the secret content from the vault
func (s Secret) GetContent(v Vault) (string, error) {
	err := v.PullElevate()
	if err != nil {
		return "", fmt.Errorf("failed to elevate permissions: %w", err)
	}

	ret, er := v.Pull(s)
	if er != nil {
		return "", fmt.Errorf("failed to pull secret: %w", er)
	}
	return ret.Value, nil
}

// Pushes the secret content to the vault
func (s Secret) SetContent(v Vault, content SecretContent) error {
	err := v.PushElevate()
	if err != nil {
		return fmt.Errorf("failed to elevate permissions: %w", err)
	}

	er := v.Push(content)
	if er != nil {
		return fmt.Errorf("failed to push secret: %w", er)
	}
	return nil
}

// saves stored env variable to dotenv file.
// will update if key already exists, or create new if not
func (st StoredEnv) Save() error {
	mp, e := godotenv.Read(st.File)
	if e != nil {
		return fmt.Errorf("failed to parse dotenv file: %w", e)
	}
	currentvalue, ok := mp[st.Key]
	if ok {
		if currentvalue == st.Value {
			return nil
		}
	}

	mp[st.Key] = st.Value
	slog.Debug("saving env", "key", st.Key, "file", st.File)
	e = godotenv.Write(mp, st.File)
	if e != nil {
		return fmt.Errorf("failed to write to dotenv file: %w", e)
	}
	return nil
}

// removes stored env variable from dotenv file
func (st StoredEnv) Remove() error {
	mp, e := godotenv.Read(st.File)
	if e != nil {
		if os.IsNotExist(e) {
			return nil
		}
		return fmt.Errorf("failed to parse dotenv file: %w", e)
	}

	_, ok := mp[st.Key]
	if ok {
		delete(mp, st.Key)
	}

	e = godotenv.Write(mp, st.File)
	if e != nil {
		return fmt.Errorf("failed to write to dotenv file: %w", e)
	}
	return nil
}

//region secret detection

var (
	secretKeywords = []string{
		"API_KEY", "APITOKEN", "ACCESS_TOKEN", "AUTH_TOKEN",
		"SECRET", "PASSWORD", "PASS", "PWORD",
		"PRIVATE_KEY", "CLIENT_SECRET", "DB_PASSWORD",
	}
	sensitiveFilenames = []string{
		"secret", "credential", ".pem", ".key",
	}
	// secretRegexes is a map of a secret's description to its regex pattern.
	secretRegexes = map[string]*regexp.Regexp{
		"JWT":                  regexp.MustCompile(`ey[A-Za-z0-9_-]{10,}.[A-Za-z0-9_-]{10,}.[A-Za-z0-9_-]{10,}`),
		"GitHub Token":         regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		"Slack Token":          regexp.MustCompile(`xoxp-[0-9a-zA-Z-]{20,}`),
		"AWS Access Key ID":    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		"PEM private key":      regexp.MustCompile(`-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----`),
		"Generic Base64 (40+)": regexp.MustCompile(`[A-Za-z0-9+/=]{40,}`),
		"Generic Hex (40+)":    regexp.MustCompile(`[a-fA-F0-9]{40,}`),
	}
)

// calculateEntropy calculates the Shannon entropy of a string.
func calculateEntropy(s string) float64 {
	if s == "" {
		return 0.0
	}
	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range counts {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// DetectSecret checks if a given key/value pair is likely a secret.
// It returns true if it is a secret, along with the reason for the detection.
func (st StoredEnv) DetectSecret() (bool, string) {
	if st.Value == "" {
		return false, ""
	}

	// 1. Key Name Analysis
	upperKey := strings.ToUpper(st.Key)
	for _, keyword := range secretKeywords {
		if strings.Contains(upperKey, keyword) {
			return true, fmt.Sprintf("Keyword '%s' in key", keyword)
		}
	}

	// 2. File Context Analysis can be added here later

	// 3. Pattern Matching on Value
	for reason, re := range secretRegexes {
		if re.MatchString(st.Value) {
			return true, reason
		}
	}

	// 4. Entropy Calculation
	if calculateEntropy(st.Value) > 4.5 {
		return true, "High entropy value"
	}

	return false, ""
}
