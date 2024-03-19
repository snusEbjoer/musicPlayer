package CreatePlaylist

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"main/models/ChoosePlaylist"
	"main/playlists"
)

type Modes int

type Model struct {
	table           table.Model
	currentPlaylist string
	textInput       textinput.Model
	defaultRows     []table.Row
	focused         bool
	choosePlaylist  ChoosePlaylist.Model
}

func (m *Model) SetFocused(b bool) {
	m.focused = b
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return textinput.Blink }
func (m Model) CurrPlaylist() string {
	return m.currentPlaylist
}
func DefaultPlaylist() Model {
	columns := []table.Column{{Title: "Create playlist", Width: 30}}
	ti := textinput.New()
	ti.Placeholder = "Create playlist"
	ti.CharLimit = 156
	ti.Width = 20
	ti.Focus()
	rows := []table.Row{{ti.View()}}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(4),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	choosePlaylist := ChoosePlaylist.DefaultPlaylist()
	return Model{table: t, defaultRows: rows, focused: false, textInput: ti, choosePlaylist: choosePlaylist}
}

func (m Model) View() string {
	if m.focused {
		baseStyle.BorderForeground(lipgloss.Color("229"))
	} else {
		baseStyle.BorderForeground(lipgloss.Color("240"))
	}
	return baseStyle.Render(m.table.View())
}
func (m Model) Focus() {
	m.table.Focus()
}
func (m Model) Blur() {
	m.table.Blur()
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "enter":
			pl := playlists.P{}
			plName := m.textInput.Value()
			if len(plName) == 0 {
				return m, cmd
			}
			pl.CreatePlaylist(plName)
		default:
			m.textInput, cmd = m.textInput.Update(msg)
			m.table.SetRows([]table.Row{{m.textInput.View()}})
		}

	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
