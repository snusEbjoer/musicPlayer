package player

import (
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	ChoosePlaylist "main/models/ChoosePlaylist"
	"main/models/messages"
	"main/state"
	"math"
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
	progress       progress.Model
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m Model) Init() tea.Cmd { return nil }

func DefaultPlaylist(state *state.State) Model {
	progressBar := progress.New(progress.WithDefaultGradient())
	progressBar.ShowPercentage = false
	return Model{
		mode:          DEFAULT,
		done:          make(chan bool),
		controlLocked: false,
		state:         state,
		progress:      progressBar,
	}
}
func calcPersent(time int, length int) float64 {
	return float64((time * 100) / length)
}
func formatDuration(pos time.Duration, length time.Duration) (string, string) {
	posSec := pos.Round(time.Second).Seconds()
	lengthSec := length.Round(time.Second).Seconds()
	return fmt.Sprintf("%02v:%02v", math.Floor(posSec/60), int(posSec)%60), fmt.Sprintf("%02v:%02v", math.Floor(lengthSec/60), int(lengthSec)%60)
}
func (m Model) View() string {
	st := m.state.Streamer
	s := m.progress.ViewAs(0) + "00:00 / 00:00"

	if st != nil {
		length := (time.Second * time.Duration(st.Len()) / time.Duration(44100)).Round(time.Second)
		pos := (time.Second * time.Duration(st.Position()) / time.Duration(44100)).Round(time.Second)
		curr, end := formatDuration(pos, length)
		s = fmt.Sprintf("%v %v  %v", curr,
			m.progress.ViewAs(calcPersent(st.Position(), st.Len())/100),
			end)
	}

	switch m.mode {
	case DEFAULT:
		if m.state.CurrentWindow == state.PLAYER {
			baseStyle.BorderForeground(lipgloss.Color("229"))
		} else {
			baseStyle.BorderForeground(lipgloss.Color("240"))
		}
		return baseStyle.Render(fmt.Sprintf("%s\n", lipgloss.JoinVertical(lipgloss.Center, m.state.CurrentSong, "\n", s)) + "\n")
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
		if len(m.state.SongList) == 0 {
			return m, cmd
		}
		m.state.CurrentSong = m.state.SongList[0]
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
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
	if m.state.Streamer == nil {
		return m, cmd
	}
	if m.state.Streamer.Position() == m.state.Streamer.Len() {
		return m, func() tea.Msg {
			return messages.SongEnded(true)
		}
	}
	return m, cmd
}
