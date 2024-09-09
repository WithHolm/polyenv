package charmselect

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

type item struct {
	title       string
	description string
	selected    bool
}

func (i item) Key() string         { return i.title }
func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return fmt.Sprint(i.title, i.description) }

type model struct {
	list  list.Model
	items []list.Item
	Title string

	keys         *listKeyMap
	delegateKeys *delegateKeyMap
}

// Create a new model for execution of the app
func New() model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	return model{
		keys:         listKeys,
		delegateKeys: delegateKeys,
	}
}

func (m *model) AddItem(name string, desc string) {
	m.items = append(m.items, item{title: name, description: desc, selected: false})
}
