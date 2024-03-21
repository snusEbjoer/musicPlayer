package playlists

import (
	"os"
)

func CreatePlaylist(name string) {

	os.MkdirAll("./playlists/dir/"+name, os.ModePerm)
}
func ShowAllPlaylists() ([]string, error) {
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

func ShowAllSongs(name string) ([]string, error) {
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
func GetDefaultPlaylist() (string, error) {
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
