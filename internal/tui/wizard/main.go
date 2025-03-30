package wizard

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type (
	// holds all the "cards" that are shown
	WizModel struct {
		cards    []Card // list of cards
		debugCmd tea.Model
		index    int // what card am i on
		width    int // width of the screen
		height   int // height of the screen
	}

	Card interface {
		tea.Model // tea model,
		validate() error
		startup() (Card, tea.Cmd)
		Width() int    // width of the card
		Height() int   // height of the card
		Title() string // title of the card
	}
)

var _ tea.Model = &WizModel{}

func NewWizard() *WizModel {
	return &WizModel{
		cards:  make([]Card, 0),
		index:  0,
		width:  0,
		height: 0,
	}
}

func (m *WizModel) card() Card {
	return m.cards[m.index]
}

func (m *WizModel) updateCard(c Card) {
	m.cards[m.index] = c
}

func (m *WizModel) nextCard() {
	m.index++
}

func (m *WizModel) hasMoreCards() bool {
	hasMore := m.index+1 < len(m.cards)
	slog.Debug("more cards", "index", m.index, "len", len(m.cards), "hasmore", hasMore)
	return hasMore
}

func (m *WizModel) AddCards(cards ...Card) {
	for _, c := range cards {
		e := c.validate()
		if e != nil {
			slog.Error("card failed validation", "card", c.Title(), "error", e.Error())
		}
	}
	m.cards = append(m.cards, cards...)
}

func (m *WizModel) Run() error {
	_, e := tea.NewProgram(m).Run()
	if e != nil {
		return e
	}
	return nil
}

func (m *WizModel) Init() tea.Cmd {
	slog.Debug("wizard init called.", "cards", len(m.cards))
	cmds := make([]tea.Cmd, 0)
	for _, c := range m.cards {
		c := c.Init()
		if c != nil {
			cmds = append(cmds, c)
		}
	}
	c, cmd := m.card().startup()
	m.updateCard(c)
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}

func (m *WizModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = m.card().Width()
		m.height = m.card().Height()
		slog.Debug(fmt.Sprintf("size. msg: %d x %d, card: %d x %d", msg.Width, msg.Height, m.width, m.height))

	case tea.KeyMsg:
		slog.Debug("key", "key", msg.String())

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	// call card to return cmd
	model, cmd := m.cards[m.index].Update(msg)
	card, ok := model.(Card)
	if !ok {
		slog.Error("item does not implement 'Card' interface")
		return m, tea.Quit
	}
	// if card has changed, update it
	m.updateCard(card)
	if cmd != nil {
		// switch incase there are other cmd's
		switch cmdMsg := cmd().(type) {
		case tea.QuitMsg:
			slog.Debug(fmt.Sprintf("card '%s' called %t", m.card().Title(), cmdMsg))
			if !m.hasMoreCards() {
				return m, cmd
			}

			m.nextCard()
			c, cmd := m.card().startup()
			m.updateCard(c)
			cmds = append(cmds, cmd)
		default:
			return m, cmd
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *WizModel) View() string {
	return m.card().View()
}

// region EnterKey
type delegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("[enter]", "confirm selection"),
		),
	}
}
