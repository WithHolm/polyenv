package keyvault

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/withholm/polyenv/internal/model"
)

func (c *Client) Pull(s model.Secret) (model.SecretContent, error) {
	var sec model.SecretContent
	kvSecret, err := c.client.GetSecret(context.Background(), s.RemoteKey, "", nil)
	if err != nil {
		return sec, fmt.Errorf("failed to read secret %s: %s", s.RemoteKey, err)
	}

	if kvSecret.ContentType != nil {
		sec.ContentType = *kvSecret.ContentType
	}
	if kvSecret.Value != nil {
		sec.Value = *kvSecret.Value
	}
	sec.RemoteKey = s.RemoteKey
	sec.LocalKey = s.LocalKey

	return sec, nil
}

var elevated = false

func (c *Client) PullElevate() error {
	slog.Debug("Keyvault PIM elevate not implemented yet. returning no error")
	if elevated {
		return nil
	}
	elevated = true
	time.Sleep(time.Second * 2)
	return nil
}
