package main

import (
	"fmt"
	"log"
	PlaylistsTable "main/models/Playlists"
	"main/models/SearchSong"
	"main/models/Songs"
	"main/models/player"
	"main/state"
	"main/youtube"
	"math"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var program = tea.NewProgram(initialModel())

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

type SongsUpdatedMsg bool

type model struct {
	mode             Mode
	cursor           int
	focusedWindowIdx Windows
	state            *state.State

	searchSong searchsong.Model
	playlist   PlaylistsTable.Model
	songs      Songs.Model
	player     player.Model
}

func (m model) Init() tea.Cmd { return nil }

func initialModel() model {
	state := state.New()
	playlistsTable := PlaylistsTable.DefaultPlaylist(state)
	err := state.UpdateSongs()
	if err != nil {
		log.Fatal(err)
	}
	songs, _ := Songs.DefaultSongs(state)
	player := player.DefaultPlaylist(state)

	return model{
		mode:       NORMAL,
		playlist:   playlistsTable,
		cursor:     0,
		searchSong: searchsong.DefaultSearchSong(state),
		songs:      songs,
		player:     player,
		state:      state,
	}
}

func (m *model) SwitchFocus() {
	switch m.focusedWindowIdx {
	case PLAYLISTS:
		m.state.CurrentWindow = state.PLAYLISTS
	case SEARCHSONG:
		m.state.CurrentWindow = state.SEARCH
	case SONGS:
		m.state.CurrentWindow = state.SONGS
	case PLAYER:
		m.state.CurrentWindow = state.PLAYER
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
	case Songs.SongsUpdatedMsg:
		m.state.UpdateSongs()
		m.songs, _ = m.songs.Update(Songs.SongsUpdatedMsg(true))
		return m, nil
	case searchsong.DownloadMessage:
		{
			yt := youtube.C{}
			dlUrl, err := yt.DownloadVideo(msg.Option)
			if err != nil {
				log.Fatal(err)
			}
			go func() {
				yt.Download(dlUrl.DownloadUrl, msg.Option.Title, m.state.CurrentPlaylist)
				program.Send(Songs.SongsUpdatedMsg(true))
			}()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if m.cursor > 0 {
				m.cursor--
				m.focusedWindowIdx = Windows(m.cursor)
				m.SwitchFocus()
			}
		case "right":
			if m.cursor < 3 {
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
				m.songs.SetCurrPlaylist(m.state.CurrentPlaylist)
			default:
				m.playlist, cmd = m.playlist.Update(msg)
				return m, cmd
			}
		case SEARCHSONG:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				m.searchSong, cmd = m.searchSong.Update(msg, m.state.CurrentPlaylist)

			}
		case SONGS:
			switch msg.String() {
			case "enter":
				m.songs, cmd = m.songs.Update(msg)
			default:
				m.songs, cmd = m.songs.Update(msg)
			}
		case PLAYER:
			switch msg.String() {
			case "alt+right":
				m.songs.NextSong()
				m.player.EndSong()
				go m.player.PlaySong()
			case "alt+left":
				m.songs.PrevSong()
				m.player.EndSong()
				go m.player.PlaySong()
			default:
				m.player, cmd = m.player.Update(msg)
			}

		}
	}
	return m, cmd
}

func SendUpdatedMsg() {
	program.Send(SongsUpdatedMsg(true))
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
	s += fmt.Sprintf(
		"%s\n%s\n%s %d %d %d %s",
		mergeViewsInRow(
			m.playlist.View(),
			m.searchSong.View(),
		),
		m.songs.View(),
		m.player.View(),
		m.cursor,
		m.focusedWindowIdx,
		m.mode,
		m.state.CurrentPlaylist,
	)
	return s
}

func main() {
	if _, err := program.Run(); err != nil {
		log.Fatalf("ASASA, there's been an error: %v", err)
	}
}

func (m *model) PlaySong(songName string) error {
	f, err := os.Open("./playlists/dir/" + m.state.CurrentPlaylist + "/" + songName)
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
