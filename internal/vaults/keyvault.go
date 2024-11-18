package vaults

import (
	"context"
	"dotenv-myvault/internal/tools"
	"dotenv-myvault/internal/vaults/keyvault"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

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
	cred        *azidentity.DefaultAzureCredential
	wizHelper   keyvault.Wizard
}

// set attributes for the client. used by repository init
func (c *KeyvaultClient) SetOptions(options map[string]string) error {
	c.envNameTag = options["ENV_NAME_TAG"]
	c.uri = options["URI"]
	c.name = options["NAME"]
	c.tenant = options["TENANT"]
	c.style = options["STYLE"]
	c.includeCert = options["INCLUDE_CERTANDKEYS"] == "true"

	if c.envNameTag == "" {
		return fmt.Errorf("env name tag cannot be empty")
	}
	if c.uri == "" {
		return fmt.Errorf("uri for keyvault cannot be empty")
	}
	if c.name == "" {
		return fmt.Errorf("name of keyvault cannot be empty")
	}
	if c.tenant == "" {
		return fmt.Errorf("tenant for keyvault cannot be empty")
	}

	return nil
}

func (c *KeyvaultClient) GetOptions() map[string]string {
	return map[string]string{
		"VAULT_TYPE":          "keyvault",
		"NAME":                c.name,
		"TENANT":              c.tenant,
		"URI":                 c.uri,
		"STYLE":               c.style,
		"ENV_NAME_TAG":        c.envNameTag,
		"INCLUDE_CERTANDKEYS": fmt.Sprintf("%t", c.includeCert),
	}
}

func (c *KeyvaultClient) setTenant(tenant string) error {
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: tenant,
	})
	if err != nil {
		return fmt.Errorf("failed to set tenant: %s", err)
	}

	c.cred = cred
	c.tenant = tenant

	return nil
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
func (c *KeyvaultClient) Push(name string, value string) error {

	contentType := "text/plain"
	secretparam := azsecrets.SetSecretParameters{
		Value:       &value,
		ContentType: &contentType,
		Tags: map[string]*string{
			c.envNameTag: &name,
		},
	}
	sName := ConvertToKeyvaultName(name)
	_, err := c.client.SetSecret(context.Background(), sName, secretparam, nil)
	if err != nil {
		return fmt.Errorf("failed to push secret: %s", err)
	}

	return nil
}

// Pull  all secrets from keyvault
func (c *KeyvaultClient) Pull() (map[string]string, error) {
	out := make(map[string]string)

	//list all secrets in vault
	opts := azsecrets.ListSecretPropertiesOptions{}
	pager := c.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %s", err)
		}
		// for each secret on page
		for _, secret := range page.Value {
			slog.Debug(tools.ToIndentedJson(secret))

			if *secret.Attributes.Enabled == false {
				slog.Debug("secret is not enabled, skipping", "secret", secret.ID.Name())
				continue
			}

			val, err := c.client.GetSecret(context.Background(), secret.ID.Name(), secret.ID.Version(), nil)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret %s: %s", secret.ID.Name(), err)
			}

			// try to get val from tags, else just use secret name
			var n string
			if val.Secret.Tags[c.envNameTag] == nil {
				slog.Debug("no env key found in tags, using secret name")
				n = secret.ID.Name()
			} else {
				n = *val.Secret.Tags[c.envNameTag]
			}
			out[n] = *val.Secret.Value
		}
	}

	return out, nil
}

// region flush
// Flush flushes a single secret from keyvault
func (c *KeyvaultClient) Flush(key string) error {
	_, err := c.client.DeleteSecret(context.Background(), key, nil)
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

func (c *KeyvaultClient) Opsie() error {
	return fmt.Errorf("not implemented, yet..")
}

func (c *KeyvaultClient) Warmup() error {
	if c.tenant == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: c.tenant,
	})
	if err != nil {
		return err
	}

	cli, err := azsecrets.NewClient(c.uri, cred, nil)
	if err != nil {
		return err
	}
	c.client = cli

	return nil
}

// WizardWarmup is used to get questions for the wizard
func (c *KeyvaultClient) WizardWarmup() {
	c.wizHelper = keyvault.Wizard{}
	c.wizHelper.StartGetTenants()

}

// return next question for wizard. will block if wizard is not warmed up
// this will also receive all channels from tenant, subscription and keyvaults receives
func (c *KeyvaultClient) WizardNext() VaultWizardCard {
	c.wizHelper.Current++
	slog.Debug("wizard current:", "index", c.wizHelper.Current)
	switch c.wizHelper.Current {
	case 1:
		// select tenant
		slog.Info("waiting for tenants")
		c.wizHelper.Tenants = <-c.wizHelper.TenantChannel
		c.wizHelper.StartGetSubscriptions()
		slog.Info("waiting for subscriptions")
		for i := 0; i < len(c.wizHelper.Tenants); i++ {
			slog.Debug("waiting for subscription lookup channel", "index", i)
			k := <-c.wizHelper.SubscriptionChannel
			c.wizHelper.Subscriptions = append(c.wizHelper.Subscriptions, k...)
		}

		slog.Info("tenants:", "count", len(c.wizHelper.Tenants))
		slog.Info("subscriptions:", "count", len(c.wizHelper.Subscriptions))

		slog.Debug("starting keyvaults")
		c.wizHelper.StartGetKeyvaults()
		q := make([]VaultWizardSelection, 0)
		for _, t := range c.wizHelper.Tenants {
			if !c.wizHelper.TenantHasSub(t) {
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
			Callback:  c.wizHelper.AnswerTenant,
		}
	case 2:
		// select keyvault
		for i := 0; i < c.wizHelper.Tenantcount; i++ {
			slog.Debug("waiting for resource graph channel", "index", i)
			k := <-c.wizHelper.ResGraphChannel
			c.wizHelper.ResGraphItems = append(c.wizHelper.ResGraphItems, k...)
		}

		q := make([]VaultWizardSelection, 0)
		for _, item := range c.wizHelper.ResGraphItems {
			if !item.InTenant(c.wizHelper.Tenant) {
				continue
			}
			q = append(q, VaultWizardSelection{
				Key:         item.Name,
				Description: fmt.Sprintf("%s(%s)", c.wizHelper.GetSubName(item.SubscriptionId), item.SubscriptionId),
			})
		}

		return VaultWizardCard{
			Title:     "What vault do you want to use?",
			Questions: q,
			Callback:  c.wizHelper.AnswerKeyvault,
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
					c.wizHelper.IncludeCert = false
				} else {
					c.wizHelper.IncludeCert = true
				}
				return nil
			},
		}
	}

	return VaultWizardCard{}
}

// cleanup after wizard is done
func (c *KeyvaultClient) WizardComplete() map[string]string {
	ret := c.wizHelper.GetWizardMap()
	return ret
}
