package plugin

import (
	"fmt"
	"log/slog"
	"os"
)

type GithubWriter struct {
}

func (e *GithubWriter) Name() string {
	return "github"
}

func (e *GithubWriter) AcceptedFormats() (accepted []string, deny []string) {
	return []string{"dotenv"}, []string{}
}

func (e *GithubWriter) Write(data []byte) error {
	//validate that env is set
	envFile := os.Getenv("GITHUB_ENV")
	if envFile == "" {
		slog.Error("no GITHUB_ENV set. are you running this in a github action?")
		os.Exit(1)
	}
	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		slog.Error("failed to open GITHUB_ENV file", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("failed to close GITHUB_ENV file", "error", err)
		}
	}()

	//append lineshift to end of data
	data = append(data, []byte("\n")...)

	if _, err := f.Write(data); err != nil {
		slog.Error("failed to write to GITHUB_ENV file", "error", err)
		os.Exit(1)
	}

	fmt.Println("Wrote environment variable to GITHUB_ENV")
	return nil
}
