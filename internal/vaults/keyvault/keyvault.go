/*
notify:@withholm
*/
package keyvault

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type KeyvaultClient struct {
	//name of the keyvault
	name string

	//tenant of the keyvault
	tenant string

	//name of the tag that contains the env key
	envNameTag string

	//uri of the keyvault
	uri string

	//subscription of the keyvault
	subscription string

	//include keys and certificates
	includeCert bool

	//az client
	client *azsecrets.Client

	//wizard
	wiz *wizard
}

// validate that client implemts the vault interface -> done at vaults.go to avoid circular dependency

// returns the display name of the vault
func (cli *KeyvaultClient) DisplayName() string {
	return "Azure Key Vault"
}

// set attributes for the client. used by repository init
func (cli *KeyvaultClient) SetOptions(options map[string]string) error {
	cli.envNameTag = options["ENV_NAME_TAG"]
	cli.uri = options["URI"]
	cli.name = options["NAME"]
	cli.tenant = options["TENANT"]
	// cli.style = options["STYLE"]
	cli.includeCert = options["INCLUDE_CERTANDKEYS"] == "true"
	if cli.envNameTag == "" {
		return fmt.Errorf("env name tag cannot be empty")
	}
	if cli.uri == "" {
		return fmt.Errorf("uri for keyvault cannot be empty")
	}
	if cli.name == "" {
		return fmt.Errorf("name of keyvault cannot be empty")
	}
	if cli.tenant == "" {
		return fmt.Errorf("tenant for keyvault cannot be empty")
	}

	return nil
}

func (cli *KeyvaultClient) GetOptions() map[string]string {
	return map[string]string{
		"VAULT_TYPE":          "keyvault",
		"NAME":                cli.name,
		"TENANT":              cli.tenant,
		"URI":                 cli.uri,
		"ENV_NAME_TAG":        cli.envNameTag,
		"INCLUDE_CERTANDKEYS": fmt.Sprintf("%t", cli.includeCert),
	}
}

// Converts a string to keyvault name
func ConvertToKeyvaultName(value string) string {
	value = strings.ToLower(value)
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	result := re.ReplaceAllStringFunc(value, func(r string) string {

		return "-"
	})
	return result
}

// Push pushes a single secret to keyvault
func (cli *KeyvaultClient) Push(name string, value string) error {
	contentType := "text/plain"
	secretparam := azsecrets.SetSecretParameters{
		Value:       &value,
		ContentType: &contentType,
		Tags: map[string]*string{
			cli.envNameTag: &name,
		},
	}
	sName := ConvertToKeyvaultName(name)
	_, err := cli.client.SetSecret(context.Background(), sName, secretparam, nil)
	if err != nil {
		return fmt.Errorf("failed to push secret: %s", err)
	}

	return nil
}

// Pull  all secrets from keyvault
func (cli *KeyvaultClient) Pull() (map[string]string, error) {
	out := make(map[string]string)

	list, err := cli.List()
	if err != nil {
		return nil, err
	}
	for _, secret := range list {
		val, err := cli.client.GetSecret(context.Background(), secret, "", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret %s: %s", secret, err)
		}
		// try to get val from tags, else just use secret name
		var n string
		if val.Secret.Tags[cli.envNameTag] == nil {
			slog.With("secret", secret).Debug("no env key found in tags, using secret name")
			n = secret
		} else {
			n = *val.Secret.Tags[cli.envNameTag]
		}
		out[n] = *val.Secret.Value
	}

	return out, nil
}

// List all secrets
func (cli *KeyvaultClient) List() (out []string, err error) {
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := cli.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %s", err)
		}
		for _, secret := range page.Value {
			if !cli.includeCert && *secret.ContentType == "application/x-pkcs12" {
				continue
			}
			// slog.Debug(tools.ToIndentedJson(secret))
			if !*secret.Attributes.Enabled {
				slog.Debug("secret is not enabled, skipping", "secret", secret.ID.Name())
				continue
			}
			out = append(out, secret.ID.Name())
		}
	}
	return out, nil
}

// region flush
// Flush flushes a single secret from keyvault
func (cli *KeyvaultClient) Flush(key []string) error {
	list, err := cli.List()
	if err != nil {
		return err
	}
	for _, name := range list {
		if slices.Contains(key, name) {
			_, err := cli.client.DeleteSecret(context.Background(), name, nil)
			if err != nil {
				return fmt.Errorf("failed to delete secret: %s", err)
			}
		}
	}
	return nil
}

// endregion flush
func (cli *KeyvaultClient) Opsie() error {
	return fmt.Errorf("not implemented yet")
}

func checkAzCliInstalled() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("az cli not installed. please install it and try again")
	}
	return nil
}

func (cli *KeyvaultClient) Warmup() error {
	err := checkAzCliInstalled()
	if err != nil {
		return err
	}

	if cli.tenant == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: cli.tenant,
	})
	if err != nil {
		return err
	}

	newCli, err := azsecrets.NewClient(cli.uri, cred, nil)
	if err != nil {
		return err
	}
	cli.client = newCli

	return nil
}

// WizardWarmup is used to get questions for the wizard
func (cli *KeyvaultClient) WizardWarmup() error {
	err := checkAzCliInstalled()
	if err != nil {
		return err
	}

	cli.wiz = newWizard()
	cli.wiz.Run()
	return nil
}

// cleanup after wizard is done
func (cli *KeyvaultClient) WizardComplete() map[string]string {

	return map[string]string{
		"NAME":                cli.wiz.selectedRes.Name,
		"TENANT":              cli.wiz.selectedTenant.Id,
		"URI":                 cli.wiz.selectedRes.VaultUri,
		"STYLE":               "nocomments", // TODO: make this a setting. supported to be if i want to support comments in env settings.. mabye?
		"ENV_NAME_TAG":        "dotenvKey",
		"INCLUDE_CERTANDKEYS": fmt.Sprintf("%t", cli.wiz.IncludeCert),
	}
}
