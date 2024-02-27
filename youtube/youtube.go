package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"main/auth"
	"net/http"
	"net/url"
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

func (c *C) Search(query string) {
	client := auth.C{}
	ctx := client.ClientContext()
	httpclient := http.Client{}
	context, err := json.Marshal(map[string]any{"context": ctx.Context})
	accessToken := client.ParseTokens().AccessToken
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/search?", bytes.NewBuffer(context))
	v := url.Values{}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	v.Add("key", ctx.Key)
	v.Add("contentCheckOk", "true")
	v.Add("racyCheckOk", "true")
	v.Add("query", query)
	req.URL.RawQuery = v.Encode()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := httpclient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))

}
