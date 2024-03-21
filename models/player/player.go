package player

import (
	ChoosePlaylist "main/models/ChoosePlaylist"
	"main/models/messages"
	"main/state"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Modes int

const (
	DEFAULT Modes = iota
	REPEAT
)

type Model struct {
	mode           Modes
	choosePlaylist ChoosePlaylist.Model
	focused        bool
	done           chan bool
	controlLocked  bool
	state          *state.State
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func DefaultPlaylist(state *state.State) Model {
	return Model{
		mode:          DEFAULT,
		done:          make(chan bool),
		controlLocked: false,
		state:         state,
	}
}

func (m Model) View() string {
	switch m.mode {
	case DEFAULT:
		if m.state.CurrentWindow == state.PLAYER {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(m.state.CurrentSong) + "\n"
	}
	return m.state.CurrentSong
}
func (m *Model) EndSong() {
	speaker.Lock()
	speaker.Close()
	speaker.Unlock()
}
func (m *Model) PlaySong() error {
	f, err := os.Open("./playlists/dir/" + m.state.CurrentPlaylist + "/" + m.state.CurrentSong)
	ch := make(chan bool)
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
		ch <- true
	})))
	<-ch

	defer streamer.Close()
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case messages.SongsUpdated:
		m.state.CurrentSong = m.state.SongList[0]
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.EndSong()
			go m.PlaySong()
		default:
		}
	}

	return m, cmd
}
