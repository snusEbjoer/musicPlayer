package ChoosePlaylist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"main/playlists"
)

type Modes int

const (
	DEFAULT Modes = iota
	CHOOSEN
)

type Model struct {
	table           table.Model
	currentPlaylist string
	defaultRows     []table.Row
	mode            Modes
	focused         bool
}

func (m *Model) SetFocused(b bool) {
	m.focused = b
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }
func (m Model) CurrPlaylist() string {
	return m.currentPlaylist
}
func DefaultPlaylist() Model {
	columns := []table.Column{{Title: "Playlists", Width: 10}}
	pl := playlists.P{}
	pls, err := pl.ShowAllPlaylists()
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
	return Model{table: t, defaultRows: rows, mode: DEFAULT, focused: false}
}

func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		if m.focused {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.table.View())
	case CHOOSEN:
		return baseStyle.Render(m.table.View()) + m.currentPlaylist
	}
	return baseStyle.Render(m.table.View()) + "m.currentPlaylist"
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
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.mode = CHOOSEN
				m.currentPlaylist = m.table.SelectedRow()[0]
			}
		case CHOOSEN:
			m.table.Blur()
		}

	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
