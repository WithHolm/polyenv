package wizard

// for now, use the list model from the bubbletea library.
// we should do what we can with the alreay existing list model,
// and if this is too limited, write our own..

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/withholm/polyenv/internal/tools"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type (
	Select struct {
		Question    string                            // question to ask user
		Items       SelectChoices                     // static list of items for the list
		ItemsFunc   func() (SelectChoices, error)     // function to get items (If not provided by static list). param is data to populate
		OnReturn    func(SelectChoices, string) error // when question is answered, this is called
		Multiselect bool                              //is this a multiselect selection?
		list        list.Model                        // tea model for the list
	}

	SelectChoice struct {
		Display  string // displayed text
		Key      string // key used for reference back to vault. this will be returned from selection
		Selected bool   // highlight (default or previously selected etc..)
	}

	// its easier to check for any return data using update, if the slice is a type
	// we can call fn that returns this in init and then check for it in update, and handle it accordingly
	SelectChoices []SelectChoice

	onReturn func(SelectChoices, string) error
)

// check implements
var (
	_ Card      = &Select{}
	_ tea.Model = &Select{}
	_ list.Item = &SelectChoice{}
	_ tea.Msg   = &SelectChoices{}
)

// endregion

// region NEW
func NewSelect(q string) Select {
	return Select{
		Question:    q,
		Items:       SelectChoices{},
		ItemsFunc:   nil,
		OnReturn:    nil,
		Multiselect: false,
		list:        list.New(make([]list.Item, 0), list.NewDefaultDelegate(), 0, 0),
	}
}

// add items to the list
func (s Select) SetItems(i SelectChoices) Select {
	if s.ItemsFunc != nil {
		slog.Error("cannot set ItemsFunc and Items at the same time")
		os.Exit(1)
	}
	s.Items = i
	return s
}

func (s Select) SetItemsFunc(f func() (SelectChoices, error)) Select {
	if len(s.Items) > 0 {
		slog.Error("cannot set ItemsFunc and Items at the same time")
		os.Exit(1)
	}
	s.ItemsFunc = f
	return s
}

func (s Select) SetOnReturn(f onReturn) Select {
	s.OnReturn = f
	return s
}

//endregion

// region select func
func (s Select) validate() error {
	if s.Question == "" {
		return errors.New("question cannot be empty")
	}
	if s.Items == nil && s.ItemsFunc == nil {
		return errors.New("items have to be defined, either via Items or ItemsFunc")
	}

	if s.Items != nil && len(s.Items) == 0 {
		return errors.New("static items cannot be empty")
	}
	return nil
}

// return width of the most wide item on the list
func (s Select) Width() int {
	width := 0
	for _, a := range s.Items {
		width = tools.MathMax(len(a.Display), width)
	}
	return width
}

func (s Select) Height() int { return len(s.Items) }

func (s Select) Title() string { return s.Question }

func (s Select) Init() tea.Cmd {
	if s.ItemsFunc != nil {
		s.list.SetShowStatusBar(true)
		s.list.ToggleSpinner()
		return s.runItemsFunc
	}
	return nil
}

// on model init
func (s Select) startup() (Card, tea.Cmd) {
	cmds := []tea.Cmd{}
	slog.Debug("select start called", "title", s.Title())
	// s.list.Title = s.Question
	// if s.ItemsFunc != nil {
	// 	s.list.SetShowStatusBar(true)
	// 	// cmds = append(cmds, s.list.ToggleSpinner())
	// 	s.list.ToggleSpinner()
	// 	// cmds = append(cmds, s.list.NewStatusMessage("loading items.."))
	// 	s.list.NewStatusMessage("loading items..")
	// 	cmds = append(cmds, s.runItemsFunc)
	// 	slog.Debug("calling itemsfunc")
	// 	return s, tea.Batch(cmds...)
	// }

	slog.Debug("importing items")
	s.updateList()
	return s, tea.Batch(cmds...)
}

func (s Select) runItemsFunc() tea.Msg {
	items, err := s.ItemsFunc()
	if err != nil {
		slog.Error("failed to get items", "error", err.Error())
		os.Exit(1)
	}
	return items
}

// update bubble list model with our items
func (s Select) updateList() tea.Cmd {
	cmds := []tea.Cmd{}
	for i, item := range s.Items {
		if i < len(s.list.Items()) {
			continue
		}
		slog.Debug("insert item", "index", i, "item", item.Display)
		cmds = append(cmds, s.list.InsertItem(i, item))
	}
	return tea.Batch(cmds...)
}

// on keypress or events, figure out logic..
func (s Select) Update(msg tea.Msg) (model tea.Model, cmd tea.Cmd) {
	slog.Debug(fmt.Sprintf("select update '%s' called with msg %T", s.Title(), msg))
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		s.list.SetSize(msg.Width-h, msg.Height-v)
		return s, nil
	case tea.KeyMsg:
		if s.list.FilterState() == list.Filtering {
			break
		}

		slog.Debug("choice", "key", msg.String())
		switch keypress := msg.String(); keypress {
		case "enter":
			if s.Multiselect {
				i, ok := s.list.SelectedItem().(SelectChoice)
				if ok {
					i.Selected = !i.Selected
					return s, nil
				}
			}
			return s, tea.Quit
		}
	case SelectChoices:
		s.Items = msg
		cmds = append(cmds, s.updateList())
		s.list.StopSpinner()
		s.list.SetShowStatusBar(false)
		cmds = append(cmds, s.list.NewStatusMessage(""))
		return s, tea.Batch(cmds...)
	}

	s.list, cmd = s.list.Update(msg)
	cmds = append(cmds, cmd)
	return s, tea.Batch(cmds...)
}

func (s Select) View() string {
	return docStyle.Render(s.list.View())
}

// endregion

// region choice func
func (c SelectChoice) Title() string {
	return c.Display
}

func (c SelectChoice) Description() string { return "" }

func (c SelectChoice) FilterValue() string { return c.Display }

// endregion
