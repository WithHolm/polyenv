// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package plugin

import (
	"github.com/joho/godotenv"
	"github.com/withholm/polyenv/internal/model"
)

type DotenvFormatter struct {
}

func (f *DotenvFormatter) Name() string {
	return "dotenv"
}

func (f *DotenvFormatter) Detect(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	_, err := godotenv.UnmarshalBytes(data)
	return err == nil
}

func (f *DotenvFormatter) InputFormat(data []byte) (*model.InputData, error) {
	out, err := godotenv.UnmarshalBytes(data)
	if err != nil {
		return nil, err
	}
	return &model.InputData{IsMap: true, Value: out}, nil
}

func (f *DotenvFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	outmap := make(map[string]string)
	for _, v := range data {
		outmap[v.Key] = v.Value
	}
	str, err := godotenv.Marshal(outmap)
	if err != nil {
		return nil, err
	}

	//append newline if not empty
	if len(str) > 0 {
		str = str + "\n"
	}
	return []byte(str), nil
}

//k
