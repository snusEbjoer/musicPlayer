package searchsong

import (
	"fmt"
	"log"
	// "log"
	PlaylistsTable "main/models/Playlists"
	"main/state"
	"main/youtube"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Modes int

const (
	DEFAULT Modes = iota
	SEARCH
	CHOOSE
)

type DownloadMessage struct {
	Option youtube.SearchResult
}

type Model struct {
	textInput     textinput.Model
	table         table.Model
	defaultRows   []table.Row
	mode          Modes
	playlistTable PlaylistsTable.Model
	query         string
	options       []youtube.SearchResult
	state         *state.State
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

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
func DefaultSearchSong(state *state.State) Model {
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
	return Model{table: t, textInput: ti, defaultRows: rows, mode: DEFAULT, state: state}
}

func (m Model) View() string {
	var styles lipgloss.Style
	if m.state.CurrentWindow == state.SEARCH {
		styles = baseStyle.BorderForeground(lipgloss.Color("229"))
	} else {
		styles = baseStyle.BorderForeground(lipgloss.Color("240"))
	}
	return styles.Render(m.table.View())
}

func (m Model) getOption(title string) (youtube.SearchResult, error) {
	for _, op := range m.options {
		if op.Title == title {
			return op, nil
		}
	}
	return youtube.SearchResult{}, fmt.Errorf("option not found")
}

func (m Model) Update(msg tea.Msg, currentPlaylist string) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case DEFAULT:
			{
				switch msg.String() {
				case "esc":
					{
						m.textInput.SetValue("")
						m.table.SetRows(m.defaultRows)
					}
				case "enter":
					{
						m.query = m.textInput.Value()
						if len(m.textInput.Value()) == 0 {
							return m, cmd
						}
						m.textInput.SetValue("")
						m.mode = CHOOSE
						pl := youtube.C{}
						options, err := pl.Search(m.query)
						if err != nil {
							log.Fatal(err)
						}
						if len(options) == 0 {
							m.table.SetRows([]table.Row{{"No results, press ESC to go back."}})
							return m, cmd
						}
						m.options = options
						var rows []table.Row
						for _, op := range options {
							rows = append(rows, table.Row{op.Title})
						}
						m.table.SetRows(rows)
					}
				default:
					{
						m.textInput, cmd = m.textInput.Update(msg)
						m.table.SetRows([]table.Row{{m.textInput.View()}})
					}
				}
			}
		case CHOOSE:
			{
				switch msg.String() {
				case "esc":
					{
						m.textInput.SetValue("")
						m.mode = DEFAULT
						m.table.SetRows(m.defaultRows)
					}
				case "enter":
					{
						currOp := m.table.SelectedRow()[0]
						option, err := m.getOption(currOp)
						if err != nil {
							log.Fatal(err)
						}
						m.mode = DEFAULT
						m.textInput.SetValue("")
						m.table.SetRows(m.defaultRows)
						return m, func() tea.Msg {
							return DownloadMessage{
								Option: option,
							}
						}
					}
				default:
					{
						m.table, cmd = m.table.Update(msg)
					}
				}
			}

		}

	}
	return m, cmd
}
