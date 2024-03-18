package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"main/auth"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/itchyny/gojq"
)

type C struct{}

type Youtube interface {
	Search(query string)
	Download(videoId string)
}
type BaseParams struct {
	Key            string `json:"key"`
	ContentCheckOk bool   `json:"contentCheckOk"`
	RacyCheckOk    bool   `json:"racyCheckOk"`
	Query          string `json:"query"`
}

type SearchResult struct {
	Title   string
	VideoId string
}
type DownloadUrl struct {
	Title       string
	DownloadUrl string
}

func (c *C) mapSearchResponseToSearchResult(jsonResponse []byte) ([]SearchResult, error) {
	query, err := gojq.Parse(`
		.contents .twoColumnSearchResultsRenderer .primaryContents 
		.sectionListRenderer .contents[0] .itemSectionRenderer .contents 
		| map(select(.videoRenderer).videoRenderer) 
		| map(pick(.videoId,.title.runs[0])) 
		| map(setpath(["title"]; .title.runs[0].text)) 
		| map(setpath(["videoId"]; .videoId))`)
	if err != nil {
		return nil, err
	}

	var m map[string]any

	err = json.Unmarshal(jsonResponse, &m)
	if err != nil {
		return nil, err
	}

	results, ok := query.Run(m).Next()
	if !ok {
		return nil, fmt.Errorf("can't parse response JSON into SearchResults")
	}

	resultList, ok := results.([]interface{})

	if !ok {
		return nil, fmt.Errorf("can't parse response JSON into SearchResults")
	}

	searchResults := make([]SearchResult, 0, len(resultList))

	for _, result := range resultList {
		v := result.(map[string]interface{})
		title := v["title"].(string)
		videoId := v["videoId"].(string)
		searchResults = append(searchResults, SearchResult{
			Title:   title,
			VideoId: videoId,
		})
	}

	return searchResults, nil
}
func (c *C) Download(dlUrl string, title string, playlistPath string) error {
	req, err := http.NewRequest("GET", dlUrl, nil)
	if err != nil {
		return err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	f, _ := os.Create(fmt.Sprintf("./playlists/dir/%s/%s.mp4", playlistPath, "sample"))
	defer resp.Body.Close()
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	cmd := exec.Command(
		"ffmpeg",
		"-i",
		fmt.Sprintf(
			"./playlists/dir/%s/%s.mp4",
			playlistPath, "sample"),
		"-map", "0:a", "-c", "copy",
		fmt.Sprintf("./playlists/dir/%s/%s.mp4", playlistPath, title))
	transformCmd := exec.Command(
		"ffmpeg",
		"-i",
		fmt.Sprintf("./playlists/dir/%s/%s.mp4", playlistPath, title),
		fmt.Sprintf("./playlists/dir/%s/%s.mp3", playlistPath, title),
	)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	cmd.Run()
	transformCmd.Run()
	defer os.Remove(fmt.Sprintf("./playlists/dir/%s/sample.mp4", playlistPath))
	defer os.Remove(fmt.Sprintf("./playlists/dir/%s/%s.mp4", playlistPath, title))

	return nil
}

func (c *C) DownloadVideo(sr SearchResult) (DownloadUrl, error) {
	client := auth.C{}
	context := client.ClientContext()
	querry, err := gojq.Parse(`.streamingData.formats[0].url`)
	if err != nil {
		return DownloadUrl{}, err
	}
	ctx := client.ClientDownloadContext(sr.VideoId)
	body, err := json.Marshal(ctx)
	if err != nil {
		return DownloadUrl{}, err
	}
	v := url.Values{}
	v.Add("key", context.Key)
	//v.Add("html5", "1")
	//v.Add("eurl", "https://youtube.googleapis.com/v/"+sr.VideoId)
	//v.Add("c", "TVHTML5")
	//v.Add("contentCheckOk", "true")
	//v.Add("racyCheckOk", "true")
	// v.Add("cver", "6.20180913")
	//v.Add("videoId", sr.VideoId)

	req, err := http.NewRequest("POST", "https://youtubei.googleapis.com/youtubei/v1/player?", bytes.NewBuffer(body))
	if err != nil {
		return DownloadUrl{}, err
	}
	req.URL.RawQuery = v.Encode()
	resp, err := client.RequestWithAuth(req)

	if err != nil {
		return DownloadUrl{}, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return DownloadUrl{}, err
	}
	var respData map[string]any
	json.Unmarshal(respBody, &respData)
	results, ok := querry.Run(respData).Next()
	if !ok {
		fmt.Errorf("can't parse response JSON into SearchResults")
	}
	resultsList, ok := results.(string)
	if !ok {
		fmt.Errorf("can't parse response JSON into SearchResults")
	}
	return DownloadUrl{DownloadUrl: resultsList, Title: sr.Title}, nil
}

// Returns result list
func (c *C) Search(query string) ([]SearchResult, error) {
	client := auth.C{}
	ctx := client.ClientContext()
	context, err := json.Marshal(map[string]any{"context": ctx.Context})
	if err != nil {
		return nil, nil
	}
	if err != nil {
		return nil, nil
	}
	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/search?", bytes.NewBuffer(context))
	if err != nil {
		return nil, nil
	}

	// TODO: MB move to separate function
	// Something like `prepareSearchRequest`
	v := url.Values{}
	v.Add("key", ctx.Key)
	v.Add("contentCheckOk", "true")
	v.Add("racyCheckOk", "true")
	v.Add("query", query)
	req.URL.RawQuery = v.Encode()
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	resp, err := client.RequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return c.mapSearchResponseToSearchResult(body)
}
