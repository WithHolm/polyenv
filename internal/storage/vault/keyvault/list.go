// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
			if secret.ID == nil {
				slog.Warn("skipping secret with nil ID", "secret", secret)
				continue
			}

			var ctype string
			if secret.ContentType != nil {
				ctype = *secret.ContentType
			}

			enabled := false
			if secret.Attributes != nil && secret.Attributes.Enabled != nil {
				enabled = *secret.Attributes.Enabled
			}

			out = append(out, model.Secret{
				ContentType: ctype,
				Enabled:     enabled,
				RemoteKey:   secret.ID.Name(),
			})
		}
	}
	return out, nil
}

func (cli *Client) ListElevate() error {
	cli.listElevateOnce.Do(func() {
		slog.Debug("Keyvault PIM elevate not implemented yet")
	})
	return nil
}
