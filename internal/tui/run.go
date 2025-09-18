// Package tui contains helper functions for the tui
package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/tools"
)

func RunHuh(f *huh.Form) {
	if f == nil {
		return
	}

	theme := huh.ThemeCatppuccin()
	theme.Focused.FocusedButton = theme.Blurred.FocusedButton.SetString("◉")
	theme.Focused.BlurredButton = theme.Blurred.BlurredButton.SetString("○")

	f = f.WithTheme(theme)

	if tools.AppConfig().Debug {
		f = f.WithProgramOptions(tea.WithAltScreen())
	}
	e := f.Run()
	if e != nil {
		if errors.Is(e, huh.ErrUserAborted) {
			fmt.Println("\nAborted.")
			os.Exit(0)
		}

		slog.Error("failed to run form", "error", e)
		fmt.Printf("%+v", e)
		panic(e.Error())
	}
}
