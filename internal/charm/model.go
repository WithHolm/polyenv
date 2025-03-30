package charm

// import (
// 	"fmt"

// 	"github.com/charmbracelet/bubbles/list"
// )

// type listItem struct {
// 	title       string
// 	description string
// 	selected    bool
// }

// func (i listItem) Key() string         { return i.title }
// func (i listItem) Title() string       { return i.title }
// func (i listItem) Description() string { return i.description }
// func (i listItem) FilterValue() string { return fmt.Sprint(i.title, i.description) }

// type model struct {
// 	list  list.Model
// 	items []list.Item
// 	Title string

// 	keys         *listKeyMap
// 	delegateKeys *delegateKeyMap
// }

// // Create a new model for execution of the app
// func New() model {
// 	var (
// 		delegateKeys = newDelegateKeyMap()
// 		listKeys     = newListKeyMap()
// 	)

// 	mod := model{
// 		keys:         listKeys,
// 		delegateKeys: delegateKeys,
// 	}

// 	return mod
// }

// // add item to the model
// func (m *model) AddItem(name string, desc string) {
// 	m.items = append(m.items, listItem{title: name, description: desc, selected: false})
// }
