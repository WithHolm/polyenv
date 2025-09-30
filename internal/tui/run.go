// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package tui contains helper functions for the tui
package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/term"
	"github.com/withholm/polyenv/internal/tools"
	// "golang.org/x/term"
)

func IsTTY() bool {
	return term.IsTerminal(uintptr(os.Stdout.Fd()))
}

func RunHuh(f *huh.Form) {
	if f == nil {
		return
	}
	if !IsTTY() {
		slog.Error("cannot run interractive content in a non interractive terminal")
		os.Exit(1)
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
