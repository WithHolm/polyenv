package vaults

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/withholm/dotenv-myvault/internal/tools"
	"github.com/withholm/dotenv-myvault/internal/vaults/keyvault"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type KeyvaultClient struct {
	name        string //name of the keyvault
	tenant      string //tenant of the keyvault
	envNameTag  string //name of the tag that contains the env key
	uri         string //uri of the keyvault
	style       string //style of storage
	includeCert bool   //include keys and certificates
	client      *azsecrets.Client
	// cred        *azidentity.DefaultAzureCredential
	wizHelper keyvault.Wizard
}

// set attributes for the client. used by repository init
func (cli *KeyvaultClient) SetOptions(options map[string]string) error {
	cli.envNameTag = options["ENV_NAME_TAG"]
	cli.uri = options["URI"]
	cli.name = options["NAME"]
	cli.tenant = options["TENANT"]
	cli.style = options["STYLE"]
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
		"STYLE":               cli.style,
		"ENV_NAME_TAG":        cli.envNameTag,
		"INCLUDE_CERTANDKEYS": fmt.Sprintf("%t", cli.includeCert),
	}
}

// func (cli *KeyvaultClient) setTenant(tenant string) error {
// 	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
// 		TenantID: tenant,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to set tenant: %s", err)
// 	}

// 	// cli.cred = cred
// 	cli.tenant = tenant

// 	return nil
// }

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

	//list all secrets in vault
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := cli.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %s", err)
		}
		// for each secret on page
		for _, secret := range page.Value {
			slog.Debug(tools.ToIndentedJson(secret))

			if !*secret.Attributes.Enabled {
				slog.Debug("secret is not enabled, skipping", "secret", secret.ID.Name())
				continue
			}

			val, err := cli.client.GetSecret(context.Background(), secret.ID.Name(), secret.ID.Version(), nil)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret %s: %s", secret.ID.Name(), err)
			}

			// try to get val from tags, else just use secret name
			var n string
			if val.Secret.Tags[cli.envNameTag] == nil {
				slog.Debug("no env key found in tags, using secret name")
				n = secret.ID.Name()
			} else {
				n = *val.Secret.Tags[cli.envNameTag]
			}
			out[n] = *val.Secret.Value
		}
	}

	return out, nil
}

// region flush
// Flush flushes a single secret from keyvault
func (cli *KeyvaultClient) Flush(key string) error {
	_, err := cli.client.DeleteSecret(context.Background(), key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %s", err)
	}
	return nil
}

// FlushAll flushes all secrets from keyvault
func (c *KeyvaultClient) FlushAll() error {
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := c.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list secrets: %s", err)
		}
		// for each secret on page
		for _, secret := range page.Value {
			err := c.Flush(secret.ID.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// endregion flush

func (cli *KeyvaultClient) Opsie() error {
	return fmt.Errorf("not implemented yet")
}

func (cli *KeyvaultClient) Warmup() error {
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
func (cli *KeyvaultClient) WizardWarmup() {
	cli.wizHelper = keyvault.Wizard{}
	cli.wizHelper.StartGetTenants()

}

// return next question for wizard. will block if wizard is not warmed up
// this will also receive all channels from tenant, subscription and keyvaults receives
func (cli *KeyvaultClient) WizardNext() VaultWizardCard {
	cli.wizHelper.Current++
	slog.Debug("wizard current:", "index", cli.wizHelper.Current)
	switch cli.wizHelper.Current {
	case 1:
		// select tenant
		slog.Info("waiting for tenants")
		cli.wizHelper.Tenants = <-cli.wizHelper.TenantChannel
		cli.wizHelper.StartGetSubscriptions()
		slog.Info("waiting for subscriptions")
		for i := 0; i < len(cli.wizHelper.Tenants); i++ {
			slog.Debug("waiting for subscription lookup channel", "index", i)
			k := <-cli.wizHelper.SubscriptionChannel
			cli.wizHelper.Subscriptions = append(cli.wizHelper.Subscriptions, k...)
		}

		slog.Info("tenants:", "count", len(cli.wizHelper.Tenants))
		slog.Info("subscriptions:", "count", len(cli.wizHelper.Subscriptions))

		slog.Debug("starting keyvaults")
		cli.wizHelper.StartGetKeyvaults()
		q := make([]VaultWizardSelection, 0)
		for _, t := range cli.wizHelper.Tenants {
			if !cli.wizHelper.TenantHasSub(t) {
				continue
			}

			q = append(q, VaultWizardSelection{
				Key:         t.Id,
				Description: t.Description(),
			})
		}

		return VaultWizardCard{
			Title:     "What tenant do you want to use?",
			Questions: q,
			Callback:  cli.wizHelper.AnswerTenant,
		}
	case 2:
		// select keyvault
		for i := 0; i < cli.wizHelper.Tenantcount; i++ {
			slog.Debug("waiting for resource graph channel", "index", i)
			k := <-cli.wizHelper.ResGraphChannel
			cli.wizHelper.ResGraphItems = append(cli.wizHelper.ResGraphItems, k...)
		}

		q := make([]VaultWizardSelection, 0)
		for _, item := range cli.wizHelper.ResGraphItems {
			if !item.InTenant(cli.wizHelper.Tenant) {
				continue
			}
			q = append(q, VaultWizardSelection{
				Key:         item.Name,
				Description: fmt.Sprintf("%s(%s)", cli.wizHelper.GetSubName(item.SubscriptionId), item.SubscriptionId),
			})
		}

		return VaultWizardCard{
			Title:     "What vault do you want to use?",
			Questions: q,
			Callback:  cli.wizHelper.AnswerKeyvault,
		}
	case 3:
		return VaultWizardCard{
			Title: "Some services also uses secrets to save certificates. do you want to include these (also handles items with contentType 'x-pkcs12')?",
			Questions: []VaultWizardSelection{
				{
					Key:         "secrets",
					Description: "Only secrets (exclude items with 'x-pkcs12' contentType)",
				},
				{
					Key:         "all",
					Description: "Secrets and Certificates (includes items with 'x-pkcs12' contentType)",
				},
			},
			Callback: func(s string) error {
				if s == "secrets" {
					cli.wizHelper.IncludeCert = false
				} else {
					cli.wizHelper.IncludeCert = true
				}
				return nil
			},
		}
	}

	return VaultWizardCard{}
}

// cleanup after wizard is done
func (cli *KeyvaultClient) WizardComplete() map[string]string {
	ret := cli.wizHelper.GetWizardMap()
	return ret
}
