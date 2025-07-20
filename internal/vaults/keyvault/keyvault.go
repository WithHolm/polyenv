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
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
)

type Client struct {
	//name of the keyvault
	Name string `toml:"name"`

	//tenant of the keyvault
	Tenant string `toml:"tenant"`

	//uri of the keyvault
	Uri string `toml:"uri"`

	client *azsecrets.Client `toml:"-"`

	wiz Wizard `toml:"-"`
}

var ctx context.Context

func Init() {
	os.Setenv("AZURE_SDK_GO_LOGGING", "all")
	azlog.SetListener(func(cls azlog.Event, msg string) {
		slog.Debug(msg)
	})
	azlog.SetEvents(azlog.EventRequest, azlog.EventResponse)
}

// validate that client implemts the vault interface -> done at vaults.go to avoid circular dependency

// returns the display name of the vault
func (cli *Client) DisplayName() string {
	return "Azure Key Vault"
}

// region new wiz
func (cli *Client) WizWarmup(m map[string]any) error {
	cli.wiz = Wizard{
		Tenant:       "",
		Subscription: "",
		Uri:          "",
		Name:         "",
		state:        0,
	}

	err := checkAzCliInstalled()
	if err != nil {
		return err
	}
	return nil
}

func (cli *Client) WizNext() *huh.Form {
	return cli.wiz.Next()
}

func (cli *Client) WizComplete() map[string]any {
	return map[string]any{
		"tenant": cli.wiz.Tenant,
		"uri":    cli.wiz.Uri,
		"name":   cli.wiz.Name,
	}
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
	if !re.MatchString(name) {
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

// List all secrets
func (cli *Client) List() (out []model.Secret, err error) {
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := cli.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %s", err)
		}
		for _, secret := range page.Value {
			out = append(out, model.Secret{
				ContentType: *secret.ContentType,
				Enabled:     *secret.Attributes.Enabled,
				RemoteKey:   secret.ID.Name(),
			})
		}
	}
	return out, nil
}

// region flush
// Flush flushes a single secret from keyvault
func (cli *Client) Flush(key []string) error {
	return fmt.Errorf("Not implemented yet")
	// list, err := cli.List()
	// if err != nil {
	// 	return err
	// }
	// for _, name := range list {
	// 	if slices.Contains(key, name) {
	// 		_, err := cli.client.DeleteSecret(context.Background(), name, nil)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to delete secret: %s", err)
	// 		}
	// 	}
	// }
	// return nil
}

// endregion flush
func (cli *Client) Opsie() error {
	return fmt.Errorf("not implemented yet")
}

func checkAzCliInstalled() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("az cli not installed. please install it and try again")
	}
	return nil
}

func (cli *Client) Warmup() error {
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

	newCli, err := azsecrets.NewClient(cli.Uri, cred, nil)
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
