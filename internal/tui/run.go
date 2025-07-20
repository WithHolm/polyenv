package tui

import (
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func RunHuh(f *huh.Form) {
	if f == nil {
		return
	}

	theme := huh.ThemeCatppuccin()
	theme.Focused.FocusedButton = theme.Blurred.FocusedButton.SetString("◉")
	theme.Focused.BlurredButton = theme.Blurred.BlurredButton.SetString("○")

	e := f.WithTheme(theme).WithProgramOptions(tea.WithAltScreen()).Run()

	if e != nil {
		slog.Error("failed to run form", "error", e.Error())
		// fmt.Fprintf(os.Stderr, "failed to run wizard: %s\n", e.Error())
		os.Exit(1)
	}
}
