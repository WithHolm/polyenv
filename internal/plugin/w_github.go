package plugin

import (
	"fmt"
	"log/slog"
	"os"
)

type GithubWriterType string

const (
	GithubToEnv    GithubWriterType = "env"
	GithubToOutput GithubWriterType = "output"
)

type GithubWriter struct {
	typ GithubWriterType
}

func (e *GithubWriter) Name() string {
	return "github"
}

func (e *GithubWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"dotenv"}, []string{}
}

func (e *GithubWriter) Write(data []byte) error {
	var outputEnv string
	switch e.typ {
	case GithubToEnv:
		outputEnv = "GITHUB_ENV"
	case GithubToOutput:
		outputEnv = "GITHUB_OUTPUT"
	}

	//validate that env is set
	envFile := os.Getenv(outputEnv)
	if envFile == "" {
		slog.Error("no env set. are you running this in a github action?", "env", outputEnv)
		os.Exit(1)
	}
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("failed to open github file", "env", outputEnv, "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("failed to close github file", "env", outputEnv, "error", err)
		}
	}()

	//append lineshift to end of data
	data = append(data, []byte("\n")...)

	if _, err := f.Write(data); err != nil {
		slog.Error("failed to write to github file", "output", outputEnv, "error", err)
		os.Exit(1)
	}

	fmt.Println("Wrote environment variable to", outputEnv)
	return nil
}
