// Sample Go code for user authorization

package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	PlaylistsTable "main/models/Playlists"
	"main/models/SearchSong"
	"os"
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
	searchSong       SearchSong.Model
	//Songs
	//player
	currentPlaylist string
	//currentSong
	//isPause
}

func (m model) Init() tea.Cmd { return nil }

func initialModel() model {
	playlists := PlaylistsTable.Model{}.DefaultPlaylist()
	return model{mode: NORMAL, playlist: playlists, cursor: 0, searchSong: SearchSong.Model{}.DefaultSearchSong(playlists.GetCurrPlaylist())}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case NORMAL:
			switch msg.String() {
			case "left":
				if m.cursor > 0 {
					m.cursor--
				}
			case "right":
				if m.cursor < 2 {
					m.cursor++
				}
			case "enter":
				m.focusedWindowIdx = Windows(m.cursor)
				m.mode = INPUT
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		case INPUT:
			switch m.focusedWindowIdx {
			case PLAYLISTS:
				switch msg.String() {
				case "esc":
					if m.playlist.Focused() {
						m.playlist.Blur()
					} else {
						m.playlist.Focus()
					}
				case "q", "ctrl+c":
					return m, tea.Quit
				default:
					m.playlist, _ = m.playlist.Update(msg)
					m.currentPlaylist = m.playlist.GetCurrPlaylist()
					if m.currentPlaylist != "" {
						m.mode = NORMAL
					}
				}
			case SEARCHSONG:
				switch msg.String() {
				case "esc":
					if m.playlist.Focused() {
						m.playlist.Blur()
					} else {
						m.playlist.Focus()
					}
				case "q", "ctrl+c":
					return m, tea.Quit
				default:
					m.searchSong, cmd = m.searchSong.Update(msg, m.currentPlaylist)

				}
			}
		}
	}
	return m, cmd
}

func (m model) View() string {
	s := "\n deeez player \n\n"
	s += fmt.Sprintf("%s%s \n %d %d %d %s", m.playlist.View(), m.searchSong.View(), m.cursor, m.focusedWindowIdx, m.mode, m.currentPlaylist)
	return s

}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	//yt := youtube.C{}
	//videoId, err := yt.Search("квинка слоумо")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(videoId)
	//dlUrl, err := yt.DownloadVideo(videoId[0])
	//if err != nil {
	//	fmt.Sprintf("cry about it")
	//}
	//fmt.Println(dlUrl)

}
