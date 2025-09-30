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

func (cli *Client) Push(s model.SecretContent) error {
	secretparam := azsecrets.SetSecretParameters{
		Value: &s.Value,
	}
	if s.ContentType != "" {
		secretparam.ContentType = &s.ContentType
	}

	if cli.client == nil {
		return fmt.Errorf("client not initialized. warmup first")
	}
	res, err := cli.client.SetSecret(context.Background(), s.RemoteKey, secretparam, nil)

	if err != nil {
		return fmt.Errorf("failed to set secret %s: %w", s.RemoteKey, err)
	}
	slog.Info("Keyvault secret set", "secret", s.RemoteKey, "version", res.ID.Version())
	return nil
}

func (cli *Client) PushElevate() error {
	cli.pushElevateOnce.Do(func() {
		slog.Debug("Keyvault PIM elevate not implemented yet")
	})
	return nil
}
