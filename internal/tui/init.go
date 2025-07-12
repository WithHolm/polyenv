package tui

import (
	"encoding/json"
	"fmt"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/withholm/polyenv/internal/vaults"
)

const maxWidth = 80

type wizState int

const (
	statePre wizState = iota
	stateVlt
	statePost
)

type InitModel struct {
	vaultType vaults.VaultType
	Vault     vaults.Vault
	state     wizState
	form      *huh.Form
	Completed bool
}

func NewInitModel(vaultType vaults.VaultType, opts map[string]string) *InitModel {
	if vaultType != "" {
		v, e := vaults.NewInitVault(string(vaultType))
		if e != nil {
			slog.Error("failed to create vault", "error", e)
			return nil
		}

		e = v.WizardWarmup(opts)
		if e != nil {
			slog.Error("failed to start vault wizard", "error", e)
		}

		return &InitModel{
			state:     stateVlt,
			vaultType: vaultType,
			Vault:     v,
			form:      v.WizardNext(),
		}
	}

	in := &InitModel{
		state:     statePre,
		vaultType: vaultType,
		Vault:     nil,
		form:      nil,
	}
	in.form = in.PreFormNext()
	return in
}

// region bubbletea
func (m InitModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// slog.Debug("msg", "str", fmt.Sprintf("%v", msg))
	if m.form.State == huh.StateCompleted {
		slog.Debug("form completed going to next", "state", m.state)
		//check if current model state is completed and move to next state
		switch m.state {
		case statePre:
			m.form = m.PreFormNext()
		case stateVlt:
			m.form = m.Vault.WizardNext()
		case statePost:
			m.form = m.PostFormNext()
		}
		cmd := m.form.Init()
		return m, cmd
		// cmds = append(cmds, cmd)
	}

	if m.form == nil {
		slog.Debug("going to next wiz state", "current", m.state, "next", m.state+1)
		m.state++
		switch m.state {
		case stateVlt:
			m.form = m.Vault.WizardNext()
		case statePost:
			m.form = m.PostFormNext()
		default:
			return m, tea.Quit
		}
		m.form.Init()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	default:
		j, e := json.Marshal(msg)
		if e != nil {
			slog.Error("failed to marshal", "error", e)
			return m, tea.Quit
		}
		slog.Debug("msg", "str", fmt.Sprintf("%s", j))
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m InitModel) View() string {
	if m.form == nil {
		return ""
	}
	if m.form.State != huh.StateCompleted {
		return m.form.View()
	}

	return ""
}

//endregion

//region other forms

var preFormGroup int
var postFormGroup int

func (m InitModel) PreFormNext() *huh.Form {
	func() { preFormGroup++ }()
	switch preFormGroup {
	case 0:
		return huh.NewForm(
			huh.NewGroup(
				vaults.VaultTypeSelector(&m.vaultType),
			),
		)
	}
	return nil
}

func (m InitModel) PostFormNext() *huh.Form {
	func() { postFormGroup++ }()
	switch postFormGroup {
	case 0:
		return huh.NewForm(huh.NewGroup(
			huh.NewNote().Title("WARNING").Description("add your dotenv file to .gitignore if you are going to pull to file!"),
		))
	}
	return nil
}
