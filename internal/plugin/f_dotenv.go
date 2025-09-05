package plugin

import (
	"github.com/joho/godotenv"
	"github.com/withholm/polyenv/internal/model"
)

type DotenvFormatter struct {
}

func (f *DotenvFormatter) Detect(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	_, err := godotenv.UnmarshalBytes(data)
	return err == nil
}

func (f *DotenvFormatter) InputFormat(data []byte) (any, model.InputFormatType) {
	out, err := godotenv.UnmarshalBytes(data)
	if err != nil {
		return err, model.InputFormatError
	}
	return out, model.InputFormatMap
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
	return []byte(str), nil
}
