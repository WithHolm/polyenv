package plugin

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

type PwshFormatter struct {
}

func (f *PwshFormatter) Name() string {
	return "pwsh"
}

func (f *PwshFormatter) Detect(data []byte) bool {
	return false
}

func (f *PwshFormatter) InputFormat(data []byte) (*model.InputData, error) {
	return nil, nil
}

func (f *PwshFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	var out []string
	for _, v := range data {
		var val string
		//if value ir only numbers
		if _, err := fmt.Sscanf(v.Value, "%d", &v.Value); err == nil {
			val = v.Value
		} else if ok, _ := tools.InequalFindInStrSlice([]string{"true", "false"}, v.Value); ok {
			val = fmt.Sprintf("$%s", v.Value)
		} else {
			val = fmt.Sprintf("\"%s\"", v.Value)
		}

		slog.Info("Setting", "env", val)
		out = append(out, fmt.Sprintf("Set-Item \"env:%s\" -value %s", v.Key, val))
	}
	return []byte(strings.Join(out, "\n")), nil
}
