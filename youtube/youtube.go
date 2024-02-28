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
	Title    string
	VideoUrl string
}

func (c C) mapSearchResponseToSearchResult(jsonResponse []byte) ([]SearchResult, error) {
	query, err := gojq.Parse(`
		.contents .twoColumnSearchResultsRenderer .primaryContents 
		.sectionListRenderer .contents[0] .itemSectionRenderer .contents 
		| map(select(.videoRenderer).videoRenderer) 
		| map(pick(.videoId,.title.runs[0])) 
		| map(setpath(["title"]; .title.runs[0].text)) 
		| map(setpath(["videoId"]; "https://youtube.com/watch?v=" + .videoId))`)
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
			Title:    title,
			VideoUrl: videoId,
		})
	}

	return searchResults, nil
}
func (c *C) Download(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	os.WriteFile("asdsda.webm", respBody, 0644)
	return nil
}
func (c *C) DownloadVideo(videoId string) ([]SearchResult, error) {
	client := auth.C{}
	context := client.ClientContext()
	ctx := client.ClientDownloadContext(videoId)
	body, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Add("key", context.Key)
	req, err := http.NewRequest("POST", "https://youtubei.googleapis.com/youtubei/v1/player?", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()
	resp, err := client.RequestWithAuth(req)
	if err != nil {
		return nil, err
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(respBody))
	return []SearchResult{}, nil
}

// Returns result list
func (c *C) Search(query string) ([]SearchResult, error) {
	client := auth.C{}
	ctx := client.ClientContext()
	context, err := json.Marshal(map[string]any{"context": ctx.Context})
	if err != nil {
		return nil, nil
	}
	tokens, err := client.ParseTokens()
	if err != nil {
		return nil, nil
	}
	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/search?", bytes.NewBuffer(context))
	if err != nil {
		return nil, nil
	}

	// TODO: MB move to separate function
	// Something like `prepareSearchRequest`
	req.Header.Add("Authorization", "Bearer "+tokens.AccessToken)
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
