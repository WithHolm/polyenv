package keyvault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/withholm/polyenv/internal/model"
)

func (c *Client) Push(s model.SecretContent) error {
	secretparam := azsecrets.SetSecretParameters{
		Value: &s.Value,
	}
	if s.ContentType != "" {
		secretparam.ContentType = &s.ContentType
	}

	if c.client == nil {
		return fmt.Errorf("client not initialized. warmup first")
	}
	res, err := c.client.SetSecret(context.Background(), s.RemoteKey, secretparam, nil)

	if err != nil {
		return fmt.Errorf("failed to set secret %s: %w", s.RemoteKey, err)
	}
	slog.Info("Keyvault secret set", "secret", s.RemoteKey, "version", res.ID.Version())
	return nil
}

func (c *Client) PushElevate() error {
	c.pushElevateOnce.Do(func() {
		slog.Debug("Keyvault PIM elevate not implemented yet")
	})
	return nil
}
