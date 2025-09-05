package plugin

import (
	"encoding/json"

	"github.com/withholm/polyenv/internal/model"
)

var bracketPairs = map[rune]rune{
	'(': ')',
	'{': '}',
	'[': ']',
}

type JsonFormatter struct {
	AsArray bool
}

func (f *JsonFormatter) Detect(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	lastChar := rune(data[len(data)-1])
	firstChar := rune(data[0])
	if firstChar != '{' && firstChar != '[' {
		return false
	}

	if lastChar != bracketPairs[firstChar] {
		return false
	}

	return json.Valid(data)
}

func (f *JsonFormatter) InputFormat(data []byte) (any, model.InputFormatType) {
	if data[0] == '{' {
		var out map[string]any
		json.Unmarshal(data, &out)
		return out, model.InputFormatMap
	}

	var out []string
	json.Unmarshal(data, &out)
	return out, model.InputFormatStrSlice
}

// format as json main
func (f *JsonFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	if f.AsArray {
		return f.OutputFormatAsArray(data)
	}

	return f.OutputFormatAsMap(data)
}

// format as json map
func (f *JsonFormatter) OutputFormatAsMap(data []model.StoredEnv) ([]byte, error) {
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
func (f *JsonFormatter) OutputFormatAsArray(data []model.StoredEnv) ([]byte, error) {
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
