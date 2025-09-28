// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package plugin

import (
	"fmt"
	"log/slog"

	"github.com/withholm/polyenv/internal/model"
)

type AzDevopsFormatter struct {
}

func (f *AzDevopsFormatter) Name() string {
	return "azdevops"
}

func (f *AzDevopsFormatter) Detect(data []byte) bool {
	return false
}

func (f *AzDevopsFormatter) InputFormat(data []byte) (*model.InputData, error) {
	return nil, nil
}

func (f *AzDevopsFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	out := make([]byte, 0)
	for _, v := range data {
		secret := v.IsSecret
		if !secret {
			var reason string
			secret, reason = v.DetectSecret()
			if reason != "" {
				slog.Info("secret detected. marking as secret in output", "reason", reason, "key", v.Key)
			}
		}
		slog.Info("setting env", "key", v.Key, "isSecret", secret)
		line := fmt.Sprintf("##vso[task.setvariable variable=%s;issecret=%v]%s\n", v.Key, secret, v.Value)
		out = append(out, []byte(line)...)
	}

	return out, nil
}
