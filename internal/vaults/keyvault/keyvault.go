/*
notify:@withholm
*/
package keyvault

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	azsec "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type azsecretsClient interface {
	NewListSecretPropertiesPager(options *azsec.ListSecretPropertiesOptions) *runtime.Pager[azsec.ListSecretPropertiesResponse]
	GetSecret(ctx context.Context, name string, version string, options *azsec.GetSecretOptions) (azsec.GetSecretResponse, error)
	SetSecret(ctx context.Context, name string, parameters azsec.SetSecretParameters, options *azsec.SetSecretOptions) (azsec.SetSecretResponse, error)
}

type Client struct {
	//name of the keyvault
	// Name string `toml:"name"`

	//tenant of the keyvault
	Tenant string `toml:"tenant"`

	//uri of the keyvault
	Uri string `toml:"uri"`

	client azsecretsClient `toml:"-"`

	wiz Wizard `toml:"-"`
}

var ctx context.Context

func init() {
	// Only enable verbose logging if explicitly requested
	if os.Getenv("AZURE_SDK_GO_LOGGING") == "" && os.Getenv("DEBUG") == "true" {
		os.Setenv("AZURE_SDK_GO_LOGGING", "all")
	}
	azlog.SetListener(func(cls azlog.Event, msg string) {
		//remove package stack when posting err
		reqErr := strings.Contains(msg, "REQUEST ERROR")
		ipErr := strings.Contains(msg, "169.254.169.254")
		// config := tools.AppConfig()
		// slog.Debug("test", "req error", reqErr, "ip error", ipErr, "truncate debug", config.TruncateDebug)
		if reqErr && ipErr {
			msg, _, _ = strings.Cut(msg, "github.com")
			msg += "\n ...TRUNCATED: removed callstack..."
		}
		// slog.Debug("test", "req error", reqErr, "ip error", ipErr)
		slog.Debug(string(cls) + " -> " + msg)
	})
	azlog.SetEvents(azlog.EventRequest, azlog.EventResponse)
}

// validate that client implemts the vault interface -> done at vaults.go to avoid circular dependency

// returns the display name of the vault
func (cli *Client) DisplayName() string {
	return "Azure Key Vault"
}

func (cli *Client) ToString() string {
	return fmt.Sprintf("%s/%s", cli.Tenant, cli.Uri)
}

// Validate the secret name from input
func (cli *Client) ValidateSecretName(name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("should not be empty")
	}
	if strings.ToLower(name) != name {
		return cli.convertToKeyvaultName(name), fmt.Errorf("should be all lowercase")
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	if re.MatchString(name) {
		return cli.convertToKeyvaultName(name), fmt.Errorf("must only contain letters, numbers, and hyphens")
	}
	return name, nil
}

// converts a "name" to a name that can be used in keyvault
func (cli *Client) convertToKeyvaultName(value string) string {
	value = strings.ToLower(value)
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	result := re.ReplaceAllStringFunc(value, func(r string) string {
		return "-"
	})
	return result
}

func checkAzCliInstalled() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("az cli not installed. please install it and try again")
	}
	return nil
}

func (cli *Client) Marshal() map[string]any {
	return map[string]any{
		"type":   "keyvault",
		"tenant": cli.Tenant,
		"uri":    cli.Uri,
	}
}

func (cli *Client) Unmarshal(m map[string]any) error {
	tenant, ok := m["tenant"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing tenant")
	}
	uri, ok := m["uri"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing uri")
	}

	cli.Tenant = tenant
	cli.Uri = uri
	return nil
}

func (cli *Client) Warmup() error {
	slog.Debug("warming up vault client", "tenant", cli.Tenant, "uri", cli.Uri)
	err := checkAzCliInstalled()
	if err != nil {
		return err
	}

	if cli.Tenant == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: cli.Tenant,
	})
	if err != nil {
		return err
	}

	newCli, err := azsec.NewClient(cli.Uri, cred, nil)
	if err != nil {
		return err
	}
	cli.client = newCli

	return nil
}

func (cli *Client) ValidateConfig(options map[string]any) error {
	if options["type"] != "keyvault" {
		return fmt.Errorf("invalid vault type: %s. expecting keyvault", options["type"])
	}

	if options["tenant"] == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	//TODO: add url validation?
	if options["uri"] == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	return nil
}
