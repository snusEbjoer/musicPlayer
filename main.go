package main

import (
	"fmt"
	"log"
	"main/auth"
	"main/models/Songs"
	"main/models/messages"
	"main/models/player"
	"main/models/playlists_table"
	"main/models/search_song"
	"main/state"
	"main/youtube"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var _ = speaker.Init(44100, 4410)
var ctrl = beep.Ctrl{}
var volume = effects.Volume{Base: 2, Volume: 0}

var m = initialModel()
var program = tea.NewProgram(m)

const debounceTime = 300 * time.Millisecond

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
	focusedWindowIdx Windows
	state            *state.State

	searchSong  searchsong.Model
	playlist    PlaylistsTable.Model
	songs       songs.Model
	player      player.Model
	Scheduler   *time.Ticker
	SongPlaying bool
	err         string
	toggleHelp  bool
}

func (m model) Init() tea.Cmd { return nil }

func initialModel() model {
	state := state.New()
	playlistsTable := PlaylistsTable.New(state)
	err := state.UpdateSongs()
	if err != nil {
		log.Fatal(err)
	}
	songs, _ := songs.New(state)
	player := player.New(state)

	return model{
		mode:             NORMAL,
		playlist:         playlistsTable,
		focusedWindowIdx: PLAYLISTS,
		searchSong:       searchsong.New(state),
		songs:            songs,
		player:           player,
		state:            state,
		Scheduler:        time.NewTicker(debounceTime),
		SongPlaying:      true,
		err:              "",
		toggleHelp:       false,
	}
}

func (m *model) StartSong() {
	f, err := os.Open("./playlists/dir/" + m.state.CurrentPlaylist + "/" + m.state.CurrentSong)
	if err != nil {
		log.Fatal(err)
	}
	streamer, _, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	speaker.Lock()
	ctrl.Streamer = streamer
	ctrl.Paused = false
	m.state.Streamer = streamer
	speaker.Unlock()
	volume.Streamer = &ctrl
	speaker.Clear()
	speaker.Play(&volume)
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
func (m *model) SetNextSong() {
	if len(m.state.SongList) == 1 {
		m.state.CurrentSong = m.state.SongList[0]
		return
	}
	for i := range m.state.SongList {
		if m.state.SongList[i] == m.state.CurrentSong {
			m.state.CurrentSong = m.state.SongList[i+1]
			return
		}
		if m.state.CurrentSong == m.state.SongList[len(m.state.SongList)-1] {
			m.state.CurrentSong = m.state.SongList[0]
			return
		}
	}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.TryPlaySound:
		if !m.SongPlaying {
			m.StartSong()
			m.SongPlaying = true
		}
	case messages.SongsUpdated:
		m.state.UpdateSongs()
		m.songs, _ = m.songs.Update(messages.SongsUpdated(true))
		m.player, cmd = m.player.Update(messages.SongsUpdated(true))
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
				program.Send(messages.SongsUpdated(true))
			}()
		}
	case tea.KeyMsg:
		switch msg.String() {
		case m.state.Keys.ToggleHelp:
			m.toggleHelp = !m.toggleHelp
		case m.state.Keys.MoveToLeft, m.state.Keys.VimMoveLeft:
			if m.focusedWindowIdx > 0 {
				m.focusedWindowIdx--
				m.SwitchFocus()
			}
		case m.state.Keys.MoveToRight, m.state.Keys.VimMoveRight:
			if m.focusedWindowIdx < 3 {
				m.focusedWindowIdx++
				m.SwitchFocus()
			}
		case m.state.Keys.Quit:
			return m, tea.Quit
		}
		switch m.focusedWindowIdx {
		case PLAYLISTS:
			switch msg.String() {
			case m.state.Keys.Submit:
				m.playlist, cmd = m.playlist.Update(msg)
			default:
				m.playlist, cmd = m.playlist.Update(msg)
				return m, cmd
			}
		case SEARCHSONG:
			switch msg.String() {
			default:
				m.searchSong, cmd = m.searchSong.Update(msg)

			}
		case SONGS:
			switch msg.String() {
			case m.state.Keys.Submit:
				m.songs, cmd = m.songs.Update(msg)
			default:
				m.songs, cmd = m.songs.Update(msg)
			}
		case PLAYER:
			switch msg.String() {
			case m.state.Keys.NextSong:
				m.songs.NextSong()
				m.SongPlaying = false
				m.Scheduler.Reset(debounceTime)
			case m.state.Keys.PrevSong:
				m.songs.PrevSong()
				m.SongPlaying = false
				m.Scheduler.Reset(debounceTime)
			case m.state.Keys.Submit:
				m.StartSong()
				m.SongPlaying = true
			case m.state.Keys.PauseSong:
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			case m.state.Keys.VolumeDown:
				speaker.Lock()
				volume.Volume -= 0.1
				speaker.Unlock()
			case m.state.Keys.VolumeUp:
				speaker.Lock()
				volume.Volume += 0.1
				speaker.Unlock()
			default:
				m.player, cmd = m.player.Update(msg)
			}
		}
	}
	if m.state.Streamer != nil && m.state.Streamer.Position() == m.state.Streamer.Len() {
		m.songs.NextSong()
		m.SongPlaying = false
		m.StartSong()
	}
	return m, cmd
}
func (m *model) currPlaylistView() string {
	if m.state.CurrentPlaylist == "" {
		return "Create playlist than choose"
	}
	return m.state.CurrentPlaylist
}
func hilight(str string) string {

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})
	return keyStyle.Render(str)
}
func (m *model) HelpView() string {

	if !m.toggleHelp {
		return fmt.Sprintf("%s toggle help   %s  quit", hilight("(?)"), hilight("(ctrl+c)"))
	}
	return fmt.Sprintf("%s move up %s move down %s move left %s move right %s quit %s submit",
		hilight("(↑/k)"),
		hilight("(↓/j)"),
		hilight("(←/h)"),
		hilight("(→/l)"),
		hilight("(ctrl+c)"),
		hilight("(enter)"),
	)
}
func (m model) View() string {
	s := "\ndeeez player\n" + fmt.Sprintf("Current playlist: %s\n", m.currPlaylistView())
	s += fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.playlist.View(),
			m.searchSong.View(),
		),
		m.err,
		m.songs.View(),
		m.player.View(),
		m.HelpView(),
	)
	return s
}

func EnsureToken() {
	_, err := os.Stat("token.json")
	if os.IsNotExist(err) {
		a := auth.C{}
		err := a.FetchToken()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	EnsureToken()
	go func() {
		for _ = range m.Scheduler.C {
			program.Send(messages.TryPlaySound(true))
		}
	}()
	if _, err := program.Run(); err != nil {
		log.Fatalf("ASASA, there's been an error: %v", err)
	}
}
