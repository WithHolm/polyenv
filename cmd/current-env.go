package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/cobra"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/tools"
)

var stats bool
var output string

var outputs = []string{
	"json",
	"pwshdevops",
	"pwsh",
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

	jsonBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		slog.Error("failed to marshal json", "error", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
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

	// theme := huh.ThemeCatppuccin()
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithFocused(false),
	)

	t.View()
	// s := table.DefaultStyles()
	// style := lipgloss.NewStyle().
	// 	BorderStyle(lipgloss.RoundedBorder()).
	// style := lipgloss.NewStyle().
	// 	BorderStyle(lipgloss.NormalBorder()).
	// 	Background(theme.Focused.Base.GetBackground())
	// style.
	// Foreground(lipgloss.Color("205")).
	// BorderStyle(lipgloss.RoundedBorder()).
	// BorderForeground(lipgloss.Color("205"))

	// t.SetStyle(style)

	fmt.Print(t.View() + "\n")
	// fmt.Print(style.Render(t.View()) + "\n")

	// t.BorderStyle = lipgloss.RoundedBorder()
	// t.TitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	// t.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	// t.SetColumnStyle(0, lipgloss.NewStyle().Foreground(lipgloss.Color("205")))
	// t.SetColumnStyle(1, lipgloss.NewStyle().Foreground(lipgloss.Color("205")))
	// t.SetColumnStyle(2, lipgloss.NewStyle().Foreground(lipgloss.Color("205")))

	// t.
	// // lipgloss.NewRenderer(lipgloss.ANSI256).Render(t.View())
}
