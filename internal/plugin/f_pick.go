package plugin

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tui"
)

type PickFormatter struct {
}

func (f *PickFormatter) Name() string {
	return "pick"
}

func (f *PickFormatter) Detect(data []byte) bool {
	return false
}

func (f *PickFormatter) InputFormat(data []byte) (*model.InputData, error) {
	return nil, nil
}

func (f *PickFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	outputFormats := make([]string, 0)
	for k := range OutputFormatters {
		if AcceptsFormat(SelectedWriter, k) {
			if k == "pick" {
				continue
			}
			outputFormats = append(outputFormats, k)
		}
	}
	if len(outputFormats) == 0 {
		slog.Error("no output formats available for writer except for 'pick'. this shouldnt happen..", "writer", SelectedWriter)
		os.Exit(1)
	}

	selectedEnv := make([]model.StoredEnv, 0)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[model.StoredEnv]().
				Title("Select values to include").
				OptionsFunc(func() []huh.Option[model.StoredEnv] {
					var out []huh.Option[model.StoredEnv]
					for _, v := range data {
						out = append(out, huh.Option[model.StoredEnv]{
							Value: v,
							Key:   v.Key,
						})
					}
					return out
				}, nil).Value(&selectedEnv),
		),
	)
	tui.RunHuh(form)
	slog.Debug("selected", "env", len(selectedEnv))

	selectedFormat := outputFormats[0]
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select output format").
				Description("Select the format to output the environment variables in. the list is dependent on the writer you have selected").
				OptionsFunc(func() []huh.Option[string] {
					var out []huh.Option[string]
					for _, k := range outputFormats {
						out = append(out, huh.Option[string]{
							Value: k,
							Key:   k,
						})
					}
					return out
				}, nil).Value(&selectedFormat),
		).WithHide(len(outputFormats) == 1),
	)
	tui.RunHuh(form)

	slog.Debug("selected", "format", selectedFormat)

	//grab formatter
	formatter := OutputFormatters[selectedFormat]()
	outbytes, err := formatter.OutputFormat(selectedEnv)
	if err != nil {
		slog.Error("failed to format output", "error", err)
		os.Exit(1)
	}
	if selectedFormat == "dotenv" {
		prepend := []string{
			"# generated via 'pick' plugin for polyenv (github.com/withholm/polyenv)",
		}
		var secrets []model.StoredEnv
		for _, v := range selectedEnv {
			if v.IsSecret {
				secrets = append(secrets, v)
				continue
			}
			ok, _ := v.DetectSecret()
			if ok {
				secrets = append(secrets, v)
			}
		}
		if len(secrets) > 0 {
			prepend = append(prepend, []string{
				"# the file contains secrets and should be treated with care",
				"# {value} -> {reason}",
			}...)
			for _, v := range secrets {
				_, reason := v.DetectSecret()
				val := fmt.Sprintf("# %s -> %s", v.Key, reason)
				prepend = append(prepend, val)
			}
		}
		prepend = append(prepend, "")
		prepend = append(prepend, "")
		// outprepend = strings.Join(prepend, "\n")
		outbytes = append([]byte(strings.Join(prepend, "\n")), outbytes...)
	}
	return outbytes, nil
}
