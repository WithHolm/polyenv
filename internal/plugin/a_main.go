// Package plugin contains the polyenv plugin
package plugin

import (
	"log/slog"
	"os"
	"slices"

	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

var SelectedSource model.Source
var SelectedWriter model.Writer
var SelectedFormatter model.Formatter

var Sources = map[string]func() model.Source{
	// "ots": func() model.Source { return &OneTimeSecretSource{} },
	// "env": func() model.Source { return &EnvSource{} },
}

var InputFormatters = map[string]func() model.Formatter{
	"json":    func() model.Formatter { return &JSONFormatter{} },
	"jsonArr": func() model.Formatter { return &JSONFormatter{AsArray: true} },
	"dotenv":  func() model.Formatter { return &DotenvFormatter{} },
}

var OutputFormatters = map[string]func() model.Formatter{
	"json":     func() model.Formatter { return &JSONFormatter{AsArray: false} },
	"jsonArr":  func() model.Formatter { return &JSONFormatter{AsArray: true} },
	"pwsh":     func() model.Formatter { return &PwshFormatter{} },
	"stats":    func() model.Formatter { return &StatsFormatter{} },
	"dotenv":   func() model.Formatter { return &DotenvFormatter{} },
	"azdevops": func() model.Formatter { return &AzDevopsFormatter{} },
	"pick":     func() model.Formatter { return &PickFormatter{} },
	"posix":    func() model.Formatter { return &PosixFormatter{} },
	"bash":     func() model.Formatter { return &PosixFormatter{} },
}

var Writers = map[string]func() model.Writer{
	"stdout":     func() model.Writer { return &StdOutWriter{} },
	"github-env": func() model.Writer { return &GithubWriter{typ: GithubToEnv} },
	"github-out": func() model.Writer { return &GithubWriter{typ: GithubToOutput} },
	// "ots":        func() model.Writer { return &OtsWriter{} },
}

func init() {
	//validate writers
	for k, v := range Writers {
		accept, deny := v().AcceptedFormats()
		acceptAll := slices.Contains(accept, "*")
		denyAll := slices.Contains(deny, "*")

		if len(accept) == 0 {
			slog.Error("writer does not support any formats. it should allow atleast one format or '*' to accept all formats", "writer", k)
			os.Exit(1)
		}

		if acceptAll && accept[len(accept)-1] != "*" {
			slog.Error("* acceptance must be the last item in the accept list", "writer", k)
			os.Exit(1)
		}

		if denyAll {
			slog.Error("writer cannot deny all formats ('*' is not allowed in deny list)", "writer", k)
			os.Exit(1)
		}

		allformatters := append(tools.MapKeySlice(OutputFormatters), tools.MapKeySlice(InputFormatters)...)
		slices.Sort(allformatters)
		allformatters = slices.Compact(allformatters)
		if acceptAll && len(deny) >= len(allformatters) {
			slog.Error("writer cannot deny all formats.", "writer", k)
			os.Exit(1)
		}

		//validate that all formats are in the registry
		for _, f := range accept {
			if f == "*" {
				continue
			}
			if !slices.Contains(allformatters, f) {
				slog.Error("plugin error: accepted format not in registry", "writer", k, "format", f)
				os.Exit(1)
			}
		}

		for _, f := range deny {
			if f == "*" {
				continue
			}
			if !slices.Contains(allformatters, f) {
				slog.Error("plugin error: denied format not in registry", "writer", k, "format", f)
				os.Exit(1)
			}
		}
	}
}

// check if writer accepts a given format
func AcceptsFormat(writer model.Writer, format string) bool {
	whitelist, blacklist := writer.AcceptedFormats()
	if slices.Contains(whitelist, "*") {
		return true
	}

	if !slices.Contains(blacklist, format) && slices.Contains(whitelist, format) {
		return true
	}

	return slices.Contains(whitelist, format)
}

// returns a format that is accepted by the writer.
// used when user-defined format is 'auto'.
func AutoOutputFormat(writer model.Writer) (string, error) {
	acceptedformatters, denyformatters := writer.AcceptedFormats()
	slog.Debug("writer allow/denies:", "allowed", acceptedformatters, "deny", denyformatters)
	if slices.Contains(acceptedformatters, "*") && len(acceptedformatters) > 1 {
		slog.Debug("writer has * and more than one accepted format. selecting first one", "selected", acceptedformatters[0])
		//if it has * and more than one 'accepted' item, select the first one
		return acceptedformatters[0], nil
	} else if slices.Contains(acceptedformatters, "*") {
		slog.Debug("writer has just *, selecting first formatter in registry that is not in deny list")
		//if it has just *, select first formatter in registry that is not in deny list
		for k := range OutputFormatters {
			if !slices.Contains(denyformatters, k) {
				return k, nil
			}
		}
	}

	return acceptedformatters[0], nil
}
