// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package keyvault

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/withholm/polyenv/internal/model"
)

func (cli *Client) Pull(s model.Secret) (model.SecretContent, error) {
	var sec model.SecretContent
	if cli.client == nil {
		return sec, fmt.Errorf("client not initialized. warmup first")
	}

	kvSecret, err := cli.client.GetSecret(context.Background(), s.RemoteKey, "", nil)
	if err != nil {
		return sec, fmt.Errorf("failed to read secret %s: %w", s.RemoteKey, err)
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

// var elevated = false

func (cli *Client) PullElevate() error {
	cli.pullElevateOnce.Do(func() {
		slog.Debug("Keyvault PIM elevate not implemented yet. returning no error")
	})
	return nil
}
