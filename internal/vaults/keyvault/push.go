package keyvault

import (
	"context"
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

	res, err := c.client.SetSecret(context.Background(), s.RemoteKey, secretparam, nil)

	if err != nil {
		return err
	}
	slog.Info("Keyvault secret set", "secret", s.RemoteKey, "version", res.ID.Version())
	return nil
}

func (c *Client) PushElevate() error {
	slog.Debug("Keyvault PIM elevate not implemented yet")
	return nil
}
