package state

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/faiface/beep"
	"log"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/faiface/beep/mp3"
)

type ProgramWindow string

const (
	PLAYLISTS       ProgramWindow = "Playlists"
	CHOOSE_PLAYLIST ProgramWindow = "Choose Playlist"
	CREATE_PLAYLIST ProgramWindow = "Create Playlist"
	SONGS           ProgramWindow = "Songs"
	PLAYER          ProgramWindow = "Player"
	SEARCH          ProgramWindow = "Search"
)

func getCurrentPlaylist() string {
	files, err := os.ReadDir("./playlists/dir")
	if err != nil {
		os.Mkdir("./playlists/dir", os.ModePerm)
		return ""
	}
	if len(files) == 0 {
		return ""
	} else {
		return files[0].Name()
	}
}

type Keys struct {
	Quit         string
	Submit       string
	NextSong     string
	PrevSong     string
	Delete       string
	MoveToLeft   string
	MoveToRight  string
	PauseSong    string
	VolumeUp     string
	VolumeDown   string
	GoBack       string
	VimMoveLeft  string
	VimMoveRight string
	ToggleHelp   string
}
type State struct {
	CurrentPlaylist string
	CurrentSong     string
	CurrentWindow   ProgramWindow
	SongList        []string
	mx              sync.Mutex
	Streamer        beep.StreamSeekCloser
	Keys            Keys
}

func New() *State {
	var keys Keys
	_, err := toml.DecodeFile("./config.toml", &keys)
	if err != nil {
		log.Fatal("config.toml not found or invalid.", err)
	}
	return &State{
		CurrentPlaylist: getCurrentPlaylist(),
		CurrentSong:     "",
		CurrentWindow:   PLAYLISTS,
		SongList:        []string{},
		mx:              sync.Mutex{},
		Keys:            keys,
	}
}

func (s *State) Lock() {
	s.mx.Lock()
}

func (s *State) Unlock() {
	s.mx.Unlock()
}

func (s *State) UpdateSongs() error {
	//if s.CurrentPlaylist == "" {
	//	return fmt.Errorf("no playlist selected")
	//}
	files, err := os.ReadDir(fmt.Sprintf("./playlists/dir/%s", s.CurrentPlaylist))
	if err != nil {
		return err
	}
	s.SongList = make([]string, len(files))
	for i, _ := range files {
		s.SongList[i] = files[i].Name()
	}
	if len(s.SongList) != 0 && s.CurrentSong == "" {
		s.CurrentSong = s.SongList[0]
	}
	return nil
}

func (s *State) SongsWithDuration() ([]table.Row, error) {
	var rows []table.Row
	for _, song := range s.SongList {
		f, err := os.Open(fmt.Sprintf("./playlists/dir/%s/%s", s.CurrentPlaylist, song))
		streamer, format, err := mp3.Decode(f)
		if err != nil {
			return rows, err
		}
		rows = append(rows, table.Row{song, format.SampleRate.D(streamer.Len()).Round(time.Second).String()})
		streamer.Close()
		f.Close()
	}
	return rows, nil
}
