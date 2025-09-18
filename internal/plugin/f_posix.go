package plugin

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/withholm/polyenv/internal/model"
)

type PosixFormatter struct{}

func (f *PosixFormatter) Name() string {
	return "posix"
}

// Detect is not applicable for an output-only formatter.
func (f *PosixFormatter) Detect(data []byte) bool {
	return false
}

// InputFormat is not applicable for an output-only formatter.
func (f *PosixFormatter) InputFormat(data []byte) (*model.InputData, error) {
	return nil, nil
}

// OutputFormat converts secrets to the `export KEY="VALUE"` format.
func (f *PosixFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	// Convert the StoredEnv slice to a map for godotenv
	envMap := make(map[string]string, len(data))
	for _, v := range data {
		envMap[v.Key] = v.Value
	}

	// Use godotenv.Marshal to get a correctly formatted and escaped KEY=VALUE string
	dotenvString, err := godotenv.Marshal(envMap)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(dotenvString, "\n")
	var outputLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			outputLines = append(outputLines, "export "+line)
		}
	}

	return []byte(strings.Join(outputLines, "\n")), nil
}
