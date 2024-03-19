package main

import (
	"fmt"
	"log"
	PlaylistsTable "main/models/Playlists"
	"main/models/SearchSong"
	"main/models/Songs"
	"main/playlists"
	"math"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Mode int
type Windows int

const (
	PLAYLISTS Windows = iota
	SEARCHSONG
	SONGS
	PLAYER
)
const (
	NORMAL Mode = iota
	INPUT
)

type model struct {
	mode             Mode
	cursor           int
	focusedWindowIdx Windows
	playlist         PlaylistsTable.Model
	searchSong       searchsong.Model
	currentPlaylist  string
	songs            Songs.Model
}

func (m model) Init() tea.Cmd { return nil }

func initialModel() model {
	playlistsTable := PlaylistsTable.DefaultPlaylist()
	pl := playlists.P{}
	currentPlaylist, err := pl.GetDefaultPlaylist()
	if err != nil {
		currentPlaylist = ""
	}
	songs, _ := Songs.DefaultSongs()
	return model{
		mode:            NORMAL,
		playlist:        playlistsTable,
		cursor:          0,
		searchSong:      searchsong.DefaultSearchSong(playlistsTable.GetCurrPlaylist()),
		currentPlaylist: currentPlaylist,
		songs:           songs,
	}
}

func (m *model) BlurWindow() {
	switch m.focusedWindowIdx {
	case PLAYLISTS:
		m.playlist.SetFocused(false)
	case SEARCHSONG:
		m.searchSong.SetFocused(false)
	case SONGS:
		m.songs.SetFocused(false)
	}
}

func (m *model) SwitchFocus() {
	switch m.focusedWindowIdx {
	case PLAYLISTS:
		m.playlist.SetFocused(true)
	case SEARCHSONG:
		m.searchSong.SetFocused(true)
	case SONGS:
		m.songs.SetFocused(true)
	}
}

func (m *model) FocusTable() {
	switch m.focusedWindowIdx {
	case PLAYLISTS:
		m.playlist.Focus()
	case SEARCHSONG:
		m.searchSong.Focus()
	case SONGS:
		m.songs.Focus()
	}

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.cursor > 0 {
				m.BlurWindow()
				m.cursor--
				m.focusedWindowIdx = Windows(m.cursor)
				m.SwitchFocus()
			}
		case "right":
			if m.cursor < 2 {
				m.BlurWindow()
				m.cursor++
				m.focusedWindowIdx = Windows(m.cursor)
				m.SwitchFocus()
			}

		case "q", "ctrl+c":
			return m, tea.Quit
		}
		switch m.focusedWindowIdx {
		case PLAYLISTS:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "enter":
				m.playlist, cmd = m.playlist.Update(msg)
				m.currentPlaylist = m.playlist.GetCurrPlaylist()
				m.songs.SetCurrPlaylist(m.currentPlaylist)
			default:
				m.playlist, _ = m.playlist.Update(msg)
				m.currentPlaylist = m.playlist.GetCurrPlaylist()
			}
		case SEARCHSONG:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				m.searchSong, cmd = m.searchSong.Update(msg, m.currentPlaylist)

			}
		case SONGS:
			switch msg.String() {
			default:
				m.songs, cmd = m.songs.Update(msg)
			}
		}
	}
	return m, cmd
}

func NormalizeRight(s string, l int) string {
	if len(s) < l {
		return s + strings.Repeat(" ", l-len(s))
	}
	return s
}

func MaxRowLength(rows []string) int {
	max := 0
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}

func mergeViewsInRow(view1, view2 string) string {
	var sb strings.Builder
	rows1 := strings.Split(view1, "\n")
	rows2 := strings.Split(view2, "\n")
	maxLen := math.Max(float64(len(rows1)), float64(len(rows2)))
	maxRowLen := MaxRowLength(rows1)
	for i := 0; i < int(maxLen); i++ {
		if i < len(rows1) {
			sb.WriteString(rows1[i])
		} else {
			sb.WriteString(strings.Repeat(" ", maxRowLen))
		}
		if i < len(rows2) {
			sb.WriteString(rows2[i])
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m model) View() string {
	s := "\n deeez player \n\n"
	s += fmt.Sprintf("%s \n %s \n %d %d %d %s", mergeViewsInRow(m.playlist.View(), m.searchSong.View()), m.songs.View(), m.cursor, m.focusedWindowIdx, m.mode, m.currentPlaylist)
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("ASASA, there's been an error: %v", err)
	}
}

func (m *model) PlaySong(songName string) error {
	f, err := os.Open("./playlists/dir/" + m.currentPlaylist + "/" + songName)
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return err
	}
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	speaker.Lock()
	go speaker.Play(streamer)
	defer speaker.Unlock()
	defer streamer.Close()
	return nil
}
