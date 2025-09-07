package main

import (
	_ "embed"
	"log/slog"
	"os"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/withholm/polyenv/cmd"
	"github.com/withholm/polyenv/internal/tools"
)

//go:embed CONTRIBUTORS
var Contributors string

func init() {
	appconfig := tools.AppConfig()
	appconfig.Debug = slices.Contains(os.Args, "--debug")
	appconfig.TruncateDebug = !slices.Contains(os.Args, "--disable-truncate-debug")

	opts := log.Options{
		Level:        log.InfoLevel,
		ReportCaller: tools.AppConfig().Debug,
	}
	if tools.AppConfig().Debug {
		opts.Level = log.DebugLevel
	}

	handler := log.NewWithOptions(os.Stderr, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	cmd.SetContributors(Contributors)
	cmd.Execute()
}

//trigger ci
