package PlaylistsTable

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ChoosePlaylist "main/models/ChoosePlaylist"
)

type Modes int

const (
	DEFAULT Modes = iota
	CREATE
	CHOOSE
)

type Model struct {
	table           table.Model
	defaultRows     []table.Row
	mode            Modes
	choosePlaylist  ChoosePlaylist.Model
	currentPlaylist string
	//createPlaylist
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Focus() {
	m.table.Focus()
}
func (m Model) Blur() {
	m.table.Blur()
}
func (m Model) Focused() bool {

	return m.table.Focused()
}
func (m Model) DefaultPlaylist() Model {
	columns := []table.Column{{Title: "Playlists", Width: 20}}
	rows := []table.Row{{"Create playlist"}, {"Choose playlist"}}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(20),

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
	choosePlaylist := m.choosePlaylist.DefaultPlaylist()
	return Model{table: t, defaultRows: rows, mode: DEFAULT, choosePlaylist: choosePlaylist, currentPlaylist: ""}
}

func DefineMode(name string) Modes {
	switch name {
	case "Choose playlist":
		return CHOOSE
	case "Create playlist":
		return CREATE
	}
	return DEFAULT
}

func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		return baseStyle.Render(m.table.View()) + "\n"
	case CHOOSE:
		return m.choosePlaylist.View() + "\n" + m.currentPlaylist
	}
	return m.table.View()
}
func (m Model) GetCurrPlaylist() string {
	return m.currentPlaylist
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case DEFAULT:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.mode = DefineMode(m.table.SelectedRow()[0])
			default:
				m.table, cmd = m.table.Update(msg)
			}
		case CHOOSE:
			switch msg.String() {
			case "enter":
				m.choosePlaylist, cmd = m.choosePlaylist.Update(msg)
				m.currentPlaylist = m.choosePlaylist.CurrPlaylist()
				m.mode = DEFAULT
				return m, nil
			default:
				m.choosePlaylist, cmd = m.choosePlaylist.Update(msg)
			}

		}

	}

	return m, cmd
}
