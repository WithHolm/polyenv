package plugin

import (
	"strings"

	"github.com/withholm/polyenv/internal/model"
)

type PwshFormatter struct {
}

func (f *PwshFormatter) Detect(data []byte) bool {
	return false
}

func (f *PwshFormatter) InputFormat(data []byte) (any, model.InputFormatType) {
	return nil, 0
}

func (f *PwshFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	var out []string
	for _, v := range data {
		out = append(out, v.Key+"='"+v.Value+"'")
	}
	return []byte(strings.Join(out, ";")), nil
}
