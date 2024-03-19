package player

import (
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
	REPEAT
)

type Model struct {
	mode            Modes
	choosePlaylist  ChoosePlaylist.Model
	currentPlaylist string
	focused         bool
	songs           []string
	currentSong     string
	done            chan bool
	//progress        progress.Model
	//createPlaylist
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func DefaultPlaylist() (Model, error) {
	pl := playlists.P{}
	currPls, _ := pl.GetDefaultPlaylist()

	return Model{
		mode:            DEFAULT,
		currentPlaylist: currPls,
		focused:         false,
		currentSong:     "No song",
		done:            make(chan bool),
	}, nil
}

func (m *Model) SetFocused(b bool) {
	m.focused = b
}
func (m *Model) SetCurrentSong(song string) {
	m.currentSong = song
}
func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		if m.focused {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.currentSong) + "\n"
	}
	return m.currentSong
}
func (m *Model) EndSong() {
	_, ok := <-m.done
	if ok == false {
		m.done <- true
	}
	speaker.Close()
}
func (m *Model) PlaySong() error {
	f, err := os.Open("./playlists/dir/" + m.currentPlaylist + "/" + m.currentSong)
	//fmt.Println("./playlists/dir/" + m.currentPlaylist + "/" + songName)
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

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		_, ok := <-m.done
		if ok == false {
			m.done <- true
		}
	})))
	_, ok := <-m.done
	if ok {
		<-m.done
	}

	speaker.Close()
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
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			go m.EndSong()
			go m.PlaySong()
		case "ctrl+right":
			go m.EndSong()
			go m.PlaySong()
		case "ctrl+left":
			go m.EndSong()
			go m.PlaySong()
		default:
		}
	}

	return m, cmd
}
