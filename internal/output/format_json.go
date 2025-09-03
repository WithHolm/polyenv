package output

import (
	"encoding/json"

	"github.com/withholm/polyenv/internal/model"
)

type JsonFormatter struct {
	AsArray bool
}

// format as json main
func (f *JsonFormatter) Format(data []model.StoredEnv) ([]byte, error) {
	if f.AsArray {
		return f.FormatAsArray(data)
	}

	return f.FormatAsMap(data)
}

// format as json map
func (f *JsonFormatter) FormatAsMap(data []model.StoredEnv) ([]byte, error) {
	out := make(map[string]any)
	for _, v := range data {
		out[v.Key] = v.Value
	}
	o, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return o, nil
}

// format as json array
func (f *JsonFormatter) FormatAsArray(data []model.StoredEnv) ([]byte, error) {
	var out []map[string]any
	for _, v := range data {
		out = append(out, map[string]any{
			"key":   v.Key,
			"value": v.Value,
		})
	}
	o, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, err
	}
	return o, nil
}
