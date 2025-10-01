// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package plugin

import (
	"encoding/json"

	"github.com/withholm/polyenv/internal/model"
)

type PassthroughFormatter struct {
}

func (f *PassthroughFormatter) Name() string {
	return "passthrough"
}

func (f *PassthroughFormatter) Detect(data []byte) bool {
	return false
}

func (f *PassthroughFormatter) InputFormat(data []byte) (*model.InputData, error) {
	return nil, nil
}

func (f *PassthroughFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return out, nil
}
