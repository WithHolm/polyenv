package plugin

import (
	"encoding/json"

	"github.com/withholm/polyenv/internal/model"
)

type PassthroughFormatter struct {
}

func (f *PassthroughFormatter) Detect(data []byte) bool {
	return false
}

func (f *PassthroughFormatter) InputFormat(data []byte) (any, model.InputFormatType) {
	return nil, 0
}

func (f *PassthroughFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return out, nil
}
