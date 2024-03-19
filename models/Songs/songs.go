package Songs

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	ChoosePlaylist "main/models/ChoosePlaylist"
	"main/playlists"
	"os"
	"time"
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
	focused         bool
	songs           []string
	currentSong     string
	//createPlaylist
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
func DefaultSongs() (Model, error) {
	pl := playlists.P{}
	currPls, err := pl.GetDefaultPlaylist()
	if err != nil {
		currPls = ""
	}
	songs, err := pl.ShowAllSongs(currPls)
	if err != nil {
		return Model{}, err
	}
	columns := []table.Column{{Title: "Songs", Width: 50}, {Title: "", Width: 5}}
	var rows []table.Row
	for _, song := range songs {
		f, err := os.Open("./playlists/dir/" + currPls + "/" + song)
		streamer, format, err := mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
		}
		rows = append(rows, table.Row{song, format.SampleRate.D(streamer.Len()).Round(time.Second).String()})
		streamer.Close()
		f.Close()
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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
	return Model{
		table:           t,
		defaultRows:     rows,
		mode:            DEFAULT,
		choosePlaylist:  choosePlaylist,
		currentPlaylist: currPls,
		currentSong:     songs[0],
	}, nil
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

func (m *Model) SetFocused(b bool) {
	m.focused = b
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
	}
	return m.table.View()
}
func (m *Model) SetCurrPlaylist(newPlaylist string) {
	m.currentPlaylist = newPlaylist
	pl := playlists.P{}
	songs, _ := pl.ShowAllSongs(newPlaylist)
	var rows []table.Row
	for _, song := range songs {
		f, err := os.Open("./playlists/dir/" + m.currentPlaylist + "/" + song)
		streamer, format, err := mp3.Decode(f)
		if err != nil {
			fmt.Println(err)
		}
		rows = append(rows, table.Row{song, format.SampleRate.D(streamer.Len()).Round(time.Second).String()})
		streamer.Close()
		f.Close()
	}
	m.table.SetRows(rows)
}
func (m *Model) SetCurrentSong(song string) {
	m.currentSong = song
}
func (m *Model) GetCurrentSong() string {
	return m.currentSong
}
func (m *Model) NextSong() string {
	pl := playlists.P{}
	songs, _ := pl.ShowAllSongs(m.currentPlaylist)
	for i := range songs {
		if songs[i] == m.currentSong {
			if i == len(songs)-1 {
				m.table.SetCursor(0)
				return songs[0]
			}
			m.table.SetCursor(i + 1)
			return songs[i+1]
		}
	}
	return songs[0]
}
func (m *Model) PrevSong() string {
	pl := playlists.P{}
	songs, _ := pl.ShowAllSongs(m.currentPlaylist)
	for i := range songs {
		if songs[i] == m.currentSong {
			if i == 0 {
				m.table.SetCursor(len(songs) - 1)
				return songs[len(songs)-1]
			}
			m.table.SetCursor(i - 1)
			return songs[i-1]
		}
	}
	return songs[0]
}
func (m *Model) PlaySong(songName string) error {
	f, err := os.Open("./playlists/dir/" + m.currentPlaylist + "/" + songName)
	//fmt.Println("./playlists/dir/" + m.currentPlaylist + "/" + songName)
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	fmt.Println(format.SampleRate.D(streamer.Len()), "\n")
	if err != nil {
		return err
	}
	done := make(chan bool)
	speaker.Lock()
	go speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
		speaker.Unlock()
	})))

	<-done
	//speaker.Lock()
	//go speaker.Play(streamer)
	//defer speaker.Unlock()
	defer streamer.Close()
	return nil
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.currentSong = m.table.SelectedRow()[0]
			//go m.PlaySong(m.table.SelectedRow()[0])
		default:
			m.table, cmd = m.table.Update(msg)
		}
	}

	return m, cmd
}
