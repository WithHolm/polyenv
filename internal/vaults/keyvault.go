package vaults

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type KeyvaultClient struct {
	name       string //name of the keyvault
	tenant     string //tenant of the keyvault
	envNameTag string //name of the tag that contains the env key
	client     *azsecrets.Client
}

// init initializes the keyvault client
func (c *KeyvaultClient) Init() error {

	if c.tenant == "" {
		return fmt.Errorf("tenant cannot be empty")
	}

	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
		TenantID: c.tenant,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s.vault.azure.net", c.name)
	cli, err := azsecrets.NewClient(url, cred, nil)
	if err != nil {
		return err
	}
	c.client = cli

	c.envNameTag = "envKey"

	return nil
}

func ConvertToKeyvaultName(value string) string {
	value = strings.ToLower(value)

	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, ":", "-")
	value = strings.ReplaceAll(value, "_", "-")

	return value
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

// Pull pulls all secrets from keyvault
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
			val, err := c.client.GetSecret(context.Background(), secret.ID.Name(), secret.ID.Version(), nil)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret: %s", err)
			}

			// try to get val from tags, else just use secret name
			n := *val.Secret.Tags[c.envNameTag]
			if n == "" {
				fmt.Println("no env key found in tags, using secret name")
				n = secret.ID.Name()
			}
			out[n] = *val.Secret.Value
		}
	}

	return out, nil
}

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

func (c *KeyvaultClient) Opsie() error {
	return fmt.Errorf("not implemented, yet..")
}
