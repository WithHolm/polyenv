package charm

// import (
// 	"log/slog"
// 	"os"

// 	"github.com/charmbracelet/bubbles/key"
// 	"github.com/charmbracelet/bubbles/list"
// 	tea "github.com/charmbracelet/bubbletea"
// )

// type listKeyMap struct {
// 	// toggleSpinner    key.Binding
// 	// toggleTitleBar   key.Binding
// 	// toggleStatusBar  key.Binding
// 	togglePagination key.Binding
// 	toggleHelpMenu   key.Binding
// 	// insertItem       key.Binding
// }

// func newListKeyMap() *listKeyMap {
// 	return &listKeyMap{
// 		// toggleSpinner: key.NewBinding(
// 		// 	key.WithKeys("s"),
// 		// 	key.WithHelp("s", "toggle spinner"),
// 		// ),
// 		// toggleTitleBar: key.NewBinding(
// 		// 	key.WithKeys("T"),
// 		// 	key.WithHelp("T", "toggle title"),
// 		// ),
// 		// toggleStatusBar: key.NewBinding(
// 		// 	key.WithKeys("S"),
// 		// 	key.WithHelp("S", "toggle status"),
// 		// ),
// 		togglePagination: key.NewBinding(
// 			key.WithKeys("P"),
// 			key.WithHelp("P", "toggle pagination"),
// 		),
// 		toggleHelpMenu: key.NewBinding(
// 			key.WithKeys("H"),
// 			key.WithHelp("H", "toggle help"),
// 		),
// 	}
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	var cmds []tea.Cmd

// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		h, v := appStyle.GetFrameSize()
// 		m.list.SetSize(msg.Width-h, msg.Height-v)

// 		switch {
// 		// case key.Matches(msg, m.keys.toggleSpinner):
// 		// 	cmd := m.list.ToggleSpinner()
// 		// 	return m, cmd

// 		// case key.Matches(msg, m.keys.toggleTitleBar):
// 		// 	v := !m.list.ShowTitle()
// 		// 	m.list.SetShowTitle(v)
// 		// 	m.list.SetShowFilter(v)
// 		// 	m.list.SetFilteringEnabled(v)
// 		// 	return m, nil

// 		// case key.Matches(msg, m.keys.toggleStatusBar):
// 		// 	m.list.SetShowStatusBar(!m.list.ShowStatusBar())
// 		// 	return m, nil

// 		case key.Matches(msg, m.keys.togglePagination):
// 			m.list.SetShowPagination(!m.list.ShowPagination())
// 			return m, nil

// 		case key.Matches(msg, m.keys.toggleHelpMenu):
// 			m.list.SetShowHelp(!m.list.ShowHelp())
// 			return m, nil
// 		}
// 	}

// 	// This will also call our delegate's update function.
// 	newListModel, cmd := m.list.Update(msg)
// 	m.list = newListModel
// 	cmds = append(cmds, cmd)

// 	return m, tea.Batch(cmds...)
// }

// // used by charm
// func (m model) View() string {
// 	return appStyle.Render(m.list.View())
// }

// // Run the model
// func (m model) Run() []listItem {
// 	var (
// 		// itemGenerator randomItemGenerator
// 		// delegateKeys = newDelegateKeyMap()
// 		listKeys = newListKeyMap()
// 	)
// 	// Setup list
// 	delegate := newItemDelegate(newDelegateKeyMap())
// 	// delegate := list.NewDefaultDelegate()
// 	m.list = list.New(m.items, delegate, 0, 0)
// 	m.list.Title = m.Title
// 	m.list.Styles.Title = titleStyle
// 	m.list.SetShowStatusBar(false)
// 	m.list.AdditionalFullHelpKeys = func() []key.Binding {
// 		return []key.Binding{
// 			// listKeys.toggleSpinner,
// 			// listKeys.insertItem,
// 			// listKeys.toggleTitleBar,
// 			// listKeys.toggleStatusBar,
// 			listKeys.togglePagination,
// 			listKeys.toggleHelpMenu,
// 		}
// 	}
// 	// tea.NewProgram(model tea.Model, opts ...tea.ProgramOption)
// 	k, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
// 	if err != nil {
// 		slog.Debug("Error running selection", m.Title, err)
// 		os.Exit(1)
// 	}
// 	out := make([]listItem, 0)
// 	for _, i := range k.(model).items {
// 		if i.(listItem).selected {
// 			out = append(out, i.(listItem))
// 		}

// 	}

// 	return out
// }
