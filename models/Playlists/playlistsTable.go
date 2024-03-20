package PlaylistsTable

import (
	"fmt"
	ChoosePlaylist "main/models/ChoosePlaylist"
	"main/models/CreatePlaylist"
	"main/models/messages"
	"main/playlists"
	"main/state"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Modes int

const (
	DEFAULT Modes = iota
	CREATE
	CHOOSE
)

type Model struct {
	table          table.Model
	defaultRows    []table.Row
	mode           Modes
	choosePlaylist ChoosePlaylist.Model
	styles         lipgloss.Style
	createPlaylist CreatePlaylist.Model
	state          *state.State
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) Focus() {
	m.table.Focus()
}

func (m *Model) Blur() {
	m.table.Blur()
}
func (m *Model) Focused() bool {

	return m.table.Focused()
}
func DefaultPlaylist(state *state.State) Model {
	pl := playlists.P{}
	columns := []table.Column{{Title: "Playlists", Width: 30}}
	rows := []table.Row{{"Create playlist"}, {"Choose playlist"}}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(4),
	)

	s := table.DefaultStyles()

	defaultStyles := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

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
	choosePlaylist := ChoosePlaylist.DefaultPlaylist(state)
	createPlaylist := CreatePlaylist.DefaultPlaylist(state)
	currPls, err := pl.GetDefaultPlaylist()
	if err != nil {
		currPls = ""
	}
	fmt.Println(currPls)
	return Model{
		table:          t,
		defaultRows:    rows,
		mode:           DEFAULT,
		choosePlaylist: choosePlaylist,
		styles:         defaultStyles,
		createPlaylist: createPlaylist,
		state:          state,
	}
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
		if m.state.CurrentWindow == state.PLAYLISTS {
			m.styles.BorderForeground(lipgloss.Color("229"))
		} else {
			m.styles.BorderForeground(lipgloss.Color("240"))
		}
		return m.styles.Render(m.table.View())
	case CHOOSE:
		return m.choosePlaylist.View()
	case CREATE:
		return m.createPlaylist.View()
	}
	return m.table.View()
}
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.SongsUpdated:
		return m, func() tea.Msg {
			return messages.SongsUpdated(true)
		}
	case tea.KeyMsg:
		switch m.mode {
		case DEFAULT:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.mode = DefineMode(m.table.SelectedRow()[0])
			default:
				m.table, cmd = m.table.Update(msg)
			}
		case CHOOSE:
			switch msg.String() {
			case "esc":
				m.mode = DEFAULT
				m.choosePlaylist, cmd = m.choosePlaylist.Update(msg)
			case "enter":
				m.mode = DEFAULT
				m.choosePlaylist, cmd = m.choosePlaylist.Update(msg)
				return m, cmd
			default:
				m.choosePlaylist, cmd = m.choosePlaylist.Update(msg)
			}
		case CREATE:
			switch msg.String() {
			case "esc":
				m.mode = DEFAULT
			case "enter":
				m.createPlaylist, cmd = m.createPlaylist.Update(msg)
				m.mode = DEFAULT
				m.choosePlaylist.UpdatePlaylist()
			default:
				m.createPlaylist, cmd = m.createPlaylist.Update(msg)
			}

		}

	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
