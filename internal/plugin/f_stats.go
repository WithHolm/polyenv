package plugin

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/withholm/polyenv/internal/model"
	"github.com/withholm/polyenv/internal/polyenvfile"
	"github.com/withholm/polyenv/internal/tools"
)

type StatsFormatter struct {
	PolyenvFile *polyenvfile.File
}

func (f *StatsFormatter) Detect(data []byte) bool {
	return false
}

func (f *StatsFormatter) InputFormat(data []byte) (any, model.InputFormatType) {
	return nil, 0
}

func (f *StatsFormatter) OutputFormat(data []model.StoredEnv) ([]byte, error) {
	slices.SortFunc(data, func(a, b model.StoredEnv) int {
		return strings.Compare(a.Key, b.Key)
	})
	cwd, err := tools.GetGitRootOrCwd()
	if err != nil {
		slog.Error("failed to get git root", "error", err)
		os.Exit(1)
	}

	rows := make([]table.Row, len(data))
	longestPath := 0
	longestName := 0
	longestVault := 0
	longestReason := 0
	for i, v := range data {
		var reason string
		tags := []string{}
		sec, isSecret := f.PolyenvFile.Secrets[v.Key]
		if !isSecret {
			isSecret, reason = v.DetectSecret()
			if reason != "" {
				longestReason = max(longestReason, len(reason))
			}
		}
		rel := strings.TrimPrefix(v.File, cwd)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if isSecret {
			tags = append(tags, "sec")
		}

		//check for duplicates
		count := 0
		for _, subval := range data {
			if subval.Key == v.Key {
				count++
			}
		}
		if count > 1 {
			tags = append(tags, "dup")
		}

		rows[i] = table.Row{v.Key, strings.Join(tags, ","), rel, sec.Vault, reason}
		longestPath = max(longestPath, len(rel))
		longestName = max(longestName, len(v.Key))
		longestVault = max(longestVault, len(sec.Vault))
		slog.Debug("row", "row", rows[i])
	}

	columns := []table.Column{
		{Title: "Name", Width: longestName},
		{Title: "Tags", Width: 8},
		{Title: "Path", Width: longestPath},
		{Title: "SecretVault", Width: max(longestVault, len("SecretVault"))},
		{Title: "SecretReason", Width: max(longestReason, len("SecretReason"))},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
		table.WithFocused(false),
	)

	t.Blur()
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Cell
	t.SetStyles(s)

	return []byte(t.View()), nil
}
