package chooseplaylist

import (
	"fmt"
	"log"
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
	CHOOSEN
)

type Model struct {
	table       table.Model
	defaultRows []table.Row
	mode        Modes
	state       *state.State
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func New(state *state.State) Model {
	columns := []table.Column{{Title: "Choose playlist", Width: 40}}
	pls, err := playlists.ShowAllPlaylists()
	if err != nil {
		fmt.Println(err)
	}
	var rows []table.Row
	for _, el := range pls {
		rows = append(rows, table.Row{el})
	}
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
	return Model{table: t, defaultRows: rows, mode: DEFAULT, state: state}
}

func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		if m.state.CurrentWindow == state.PLAYLISTS {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.table.View())
	case CHOOSEN:
		if m.state.CurrentWindow == state.PLAYLISTS {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.table.View())
	}
	return baseStyle.Render(m.table.View()) + "m.currentPlaylist"
}
func (m *Model) UpdatePlaylist() {
	pls, err := playlists.ShowAllPlaylists()
	if err != nil {
		log.Fatal(err)
	}
	var rows []table.Row
	for _, el := range pls {
		rows = append(rows, table.Row{el})
	}
	m.table.SetRows(rows)
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
		switch m.mode {
		case DEFAULT:
			switch msg.String() {
			case m.state.Keys.Submit:
				m.state.CurrentPlaylist = m.table.SelectedRow()[0]
				m.state.UpdateSongs()
				return m, func() tea.Msg {
					return messages.SongsUpdated(true)
				}
			}
		}

	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
