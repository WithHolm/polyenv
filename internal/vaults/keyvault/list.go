package keyvault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/withholm/polyenv/internal/model"
)

// List all secrets
func (cli *Client) List() (out []model.Secret, err error) {
	opts := azsecrets.ListSecretPropertiesOptions{}
	if cli.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	pager := cli.client.NewListSecretPropertiesPager(&opts)

	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets: %w", err)
		}
		for _, secret := range page.Value {
			if secret.ContentType == nil || secret.Attributes == nil || secret.Attributes.Enabled == nil || secret.ID == nil {
				slog.Warn("skipping secret with nil fields", "secret", secret)
				continue
			}
			out = append(out, model.Secret{
				ContentType: *secret.ContentType,
				Enabled:     *secret.Attributes.Enabled,
				RemoteKey:   secret.ID.Name(),
			})
		}
	}
	return out, nil
}

func (cli *Client) ListElevate() error {
	slog.Debug("Keyvault PIM elevate not implemented yet")
	return nil
}
