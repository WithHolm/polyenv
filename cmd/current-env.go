package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
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
	"azas",
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
	list, err := PolyenvFile.AllDotenvValues()
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

	slog.Debug("output as", "type", output)

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
			slog.Info("setting env", "key", k, "isSecret", isSecret)
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

		dotenvContent, err := godotenv.Marshal(stringOut)
		if err != nil {
			slog.Error("failed to marshal env", "error", err)
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
		// defer f.Close()

		if _, err := f.WriteString(dotenvContent + "\n"); err != nil {
			slog.Error("failed to write to GITHUB_ENV file", "error", err)
			os.Exit(1)
		}

		fmt.Println("Wrote environment variable to GITHUB_ENV")
	case "azas":
		asAzOut := map[string]string{}
		for k, v := range out {
			val := fmt.Sprintf("%v", v)
			sec, isSecret := PolyenvFile.Secrets[k]
			if isSecret {

				// vault, ok := PolyenvFile.Vaults[sec.Vault]
				vlt, ok := PolyenvFile.VaultMap[sec.Vault]
				if !ok {
					slog.Error("vault not found", "vault", sec.Vault)
					os.Exit(1)
				}
				if vlt["type"] != "keyvault" {
					slog.Error("only keyvault references is supported for azas output", "secret", k, "vault", sec.Vault)
					os.Exit(1)
				}
				uri, ok := vlt["uri"]
				if !ok {
					slog.Error("keyvault uri not found", "vault", sec.Vault)
					os.Exit(1)
				}
				secretUri, err := url.JoinPath(uri.(string), "secrets", sec.RemoteKey)
				if err != nil {
					slog.Error("failed to join path", "error", err)
					os.Exit(1)
				}
				val = fmt.Sprintf("@Microsoft.KeyVault(%s)", secretUri)
			}
			asAzOut[k] = val
		}
		jsonBytes, err := json.MarshalIndent(asAzOut, "", "  ")
		if err != nil {
			slog.Error("failed to marshal json", "error", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
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

		rows[i] = table.Row{v.Key, strings.Join(tags, ","), rel, sec.Vault}
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
