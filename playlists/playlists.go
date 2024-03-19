package playlists

import (
	"os"
)

type Playlists interface {
	CreatePlaylist()
	AddSongToPlaylist()
	DeletePlaylist()
	DeleteFromPlaylist()
	ShowAllPlaylists()
	ShowSongsInPlaylist(name string)
}

type P struct{}

func (p *P) CreatePlaylist(name string) {

	os.MkdirAll("./playlists/dir/"+name, os.ModePerm)
}
func (p *P) ShowAllPlaylists() ([]string, error) {
	files, err := os.ReadDir("./playlists/dir/")
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(files))
	for _, v := range files {
		res = append(res, v.Name())
	}
	return res, nil
}

func (p *P) ShowAllSongs(name string) ([]string, error) {
	songs, err := os.ReadDir("./playlists/dir/" + name)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	res := make([]string, len(songs))
	for i, _ := range songs {
		res[i] = songs[i].Name()
	}
	return res, nil
}
func (p *P) GetDefaultPlaylist() (string, error) {
	files, err := os.ReadDir("./playlists/dir")
	if len(files) == 0 {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var playlist []os.DirEntry
	for _, f := range files {
		count, err := os.ReadDir("./playlists/dir/" + f.Name())
		if err != nil {
			return "", err
		}
		if len(count) >= len(playlist) {
			playlist = count
		}
	}

	return files[0].Name(), nil
}
