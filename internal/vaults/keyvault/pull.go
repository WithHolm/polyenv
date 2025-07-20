package keyvault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/withholm/polyenv/internal/model"
)

func (c *Client) Pull(s model.Secret) (model.SecretContent, error) {
	var sec model.SecretContent
	kvSecret, err := c.client.GetSecret(context.Background(), s.RemoteKey, "", nil)
	if err != nil {
		return sec, fmt.Errorf("failed to read secret %s: %s", s.RemoteKey, err)
	}

	sec.ContentType = *kvSecret.ContentType
	sec.Value = *kvSecret.Value
	sec.RemoteKey = s.RemoteKey
	sec.LocalKey = s.LocalKey

	return sec, nil
}

func (c *Client) PullElevate() error {
	slog.Debug("Keyvault PIM elevate not implemented yet")
	return nil
}
