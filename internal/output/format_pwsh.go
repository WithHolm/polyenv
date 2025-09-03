package output

import (
	"strings"

	"github.com/withholm/polyenv/internal/model"
)

type PwshFormatter struct {
}

func (f *PwshFormatter) Format(data []model.StoredEnv) ([]byte, error) {
	var out []string
	for _, v := range data {
		out = append(out, v.Key+"='"+v.Value+"'")
	}
	return []byte(strings.Join(out, ";")), nil
}
