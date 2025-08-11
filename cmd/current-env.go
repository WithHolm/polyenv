package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

var stats bool
var output string

var outputs = []string{
	"json",
	"azdevops",
	"github",
}

func generateEnvCommand() *cobra.Command {
	var envCmd = &cobra.Command{
		Use:   "env",
		Short: "list all current environment vairables read from .env files",
		Long: `
		list all current environment vairables read from .env files from git root and all sub-folders
	`,
		Run: listEnv,
	}
	envCmd.Flags().BoolVar(&stats, "stats", false, "output stats for current dotenv key-val pairs (no values)")
	envCmd.Flags().StringVarP(&output, "output", "o", outputs[0], fmt.Sprintf("outputs env variables to a given format: %s", outputs))
	err := envCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputs, cobra.ShellCompDirectiveKeepOrder
	})
	if err != nil {
		slog.Error("failed to add completion on output", "err", err)
	}
	return envCmd
}

func listEnv(cmd *cobra.Command, args []string) {
	list, err := PolyenvFile.AllDotenvKeys()
	if err != nil {
		slog.Error("failed to list env", "error", err)
		os.Exit(1)
	}

	if stats {
		listEnvStats(list)
		return
	}

	out := make(map[string]interface{})
	for _, v := range list {
		out[v.Key] = v.Value
	}

	switch output {
	case "json":
		jsonBytes, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			slog.Error("failed to marshal json", "error", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	case "azdevops":
		for k, v := range out {
			_, isSecret := PolyenvFile.Secrets[k]
			fmt.Printf("##vso[task.setvariable variable=%s;issecret=%v]%s\n", k, isSecret, v)
		}
	case "github":
		//i could just godotenv.write.. but i dont know if the file has any other content
		//so im just gonna append dotenv content to the file
		envFile := os.Getenv("GITHUB_ENV")
		if envFile == "" {
			slog.Error("no GITHUB_ENV set. are you running this in a github action?")
			os.Exit(1)
		}

		stringOut := make(map[string]string)
		for k, v := range out {
			stringOut[k] = fmt.Sprintf("%v", v)
			slog.Info("setting env", "key", k)
		}

		out, err := godotenv.Marshal(stringOut)
		if err != nil {
			slog.Error("failed to marshal env", "error", err)
			os.Exit(1)
		}

		f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err := f.WriteString(out + "\n"); err != nil {
			panic(err)
		}

		fmt.Println("Wrote environment variable to GITHUB_ENV")
	default:
		slog.Error("unhandeled output case", "case", output)
	}

}

// list stats for currently stored env variables
func listEnvStats(l []model.StoredEnv) {
	cwd, err := tools.GetGitRootOrCwd()
	if err != nil {
		slog.Error("failed to get git root", "error", err)
		os.Exit(1)
	}

	rows := make([]table.Row, len(l))
	longestPath := 0
	longestName := 0
	longestVault := 0
	for i, v := range l {
		tags := []string{}
		sec, isSecret := PolyenvFile.Secrets[v.Key]
		rel := strings.TrimPrefix(v.File, cwd)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if isSecret {
			tags = append(tags, "sec")
		}
		count := 0
		for _, subval := range l {
			if subval.Key == v.Key {
				count++
			}
		}
		if count > 1 {
			tags = append(tags, "dup")
		}

		rows[i] = table.Row{v.Key, fmt.Sprintf("%s", tags), rel, sec.Vault}
		longestPath = max(longestPath, len(rel))
		longestName = max(longestName, len(v.Key))
		longestVault = max(longestVault, len(sec.Vault))
		slog.Debug("row", "row", rows[i])
	}

	columns := []table.Column{
		{Title: "Name", Width: longestName},
		{Title: "tags", Width: 8},
		{Title: "Path", Width: longestPath},
		{Title: "SecretVault", Width: max(longestVault, len("SecretVault"))},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithFocused(false),
	)

	t.View()
	fmt.Print(t.View() + "\n")
}
