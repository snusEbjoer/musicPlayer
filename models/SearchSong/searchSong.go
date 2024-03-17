package searchsong

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	PlaylistsTable "main/models/Playlists"
	"main/youtube"
)

type Modes int

const (
	DEFAULT Modes = iota
	SEARCH
	CHOOSE
)

type Model struct {
	textInput       textinput.Model
	table           table.Model
	defaultRows     []table.Row
	mode            Modes
	playlistTable   PlaylistsTable.Model
	currentPlaylist string
	query           string
	options         []youtube.SearchResult
	focused         bool
}

func (m *Model) SetFocused(b bool) {
	m.focused = b
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
func DefaultSearchSong(currPlaylist string) Model {
	columns := []table.Column{{Title: "Search song", Width: 50}}

	ti := textinput.New()
	ti.Placeholder = "Search song"
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
	return Model{table: t, textInput: ti, defaultRows: rows, mode: DEFAULT, focused: false}
}

func DefineMode(name string) Modes {
	switch name {
	case "Choose playlist":
		return CHOOSE
	case "Create playlist":
		return CHOOSE
	}
	return DEFAULT
}

func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		if m.focused {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.table.View()) + "\n"
	case CHOOSE:
		if m.focused {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.table.View()) + "\n"
	}
	return m.table.View()
}

func (m Model) getOption(title string) youtube.SearchResult {
	for _, op := range m.options {
		if op.Title == title {
			return op
		}
	}
	return youtube.SearchResult{}
}

func (m Model) Update(msg tea.Msg, currentPlaylist string) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case DEFAULT:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.query = m.textInput.Value()
				m.textInput.SetValue("")
				m.mode = CHOOSE
				pl := youtube.C{}
				options, err := pl.Search(m.query)
				if err != nil {
					return m, tea.Quit
				}
				m.options = options
				var rows []table.Row
				for _, op := range options {
					rows = append(rows, table.Row{op.Title})
				}
				m.table.SetRows(rows)
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				m.table.SetRows([]table.Row{{m.textInput.View()}})
			}
		case CHOOSE:
			switch msg.String() {
			case "esc":
				m.table.SetRows(m.defaultRows)
			case "enter":
				currOp := m.table.SelectedRow()[0]
				option := m.getOption(currOp)
				yt := youtube.C{}
				dlUrl, err := yt.DownloadVideo(option)
				if err != nil {
					fmt.Println("cry about it") // forhead reasons
				}
				m.table.SetRows(m.defaultRows)
				go yt.Download(dlUrl.DownloadUrl, option.Title, currentPlaylist)
			default:
				m.table, cmd = m.table.Update(msg)
			}

		}

	}
	return m, cmd
}
