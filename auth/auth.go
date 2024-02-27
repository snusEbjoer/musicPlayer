package auth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type Client interface {
	ClientCredentials() ClientCredentials
	ClientContext() ClientCtx
	AccessToken()
	RefreshToken()
	FetchToken()
}
type C struct{}

type ClientCredentials struct {
	ClientSecret string
	ClientId     string
}
type UserAgent struct {
	UserAgent string `json:"User-Agent"`
}
type ClientCtx struct {
	Context ClientContx `json:"context"`
	Header  UserAgent   `json:"header"`
	Key     string      `json:"key"`
}
type ClientContx struct {
	Client Context `json:"client"`
}
type Context struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	// AndroidSdkVersion int    `json:"androidSdkVersion"`
}
type FetchData struct {
	Scope    string `json:"scope"`
	ClientId string `json:"client_id"`
}
type RespData struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	VerificationURL string `json:"verification_url"`
}
type TokenData struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DeviceCode   string `json:"device_code"`
	GrantType    string `json:"grant_type"`
}
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}
type Tokens struct {
	AccessToken  string `json:"access_token"`
	Expires      string `json:"expires"`
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenPayload struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

func (c *C) ClientContext() ClientCtx {
	ctx := ClientCtx{Context: ClientContx{Client: Context{ClientName: "WEB", ClientVersion: "2.20200720.00.02"}}, Header: UserAgent{UserAgent: "Mozilla/5.0"}, Key: "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"}
	return ctx
}

func (c *C) ClientCred() ClientCredentials {
	cred := ClientCredentials{ClientId: "861556708454-d6dlm3lh05idd8npek18k6be8ba3oc68.apps.googleusercontent.com", ClientSecret: "SboVhoG9s0rNafixCSGGKXAT"}
	return cred
}
func (c *C) ParseTokens() Tokens {
	jsonFile, err := os.Open("token.json")
	if err != nil {
		fmt.Println("File not found")
		return Tokens{}
	}
	defer jsonFile.Close()
	var fileData Tokens
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(byteValue, &fileData)
	return fileData
}
func calcExpirationTime(expiresIn int) string {
	seconds := expiresIn % 60
	minutes := expiresIn % 3600 / 60
	hours := math.Floor(float64(expiresIn / 3600))
	expires := time.Now().Local().Add(time.Second*time.Duration(seconds) + time.Minute*time.Duration(minutes) + time.Hour*time.Duration(hours))
	return expires.String()
}
func (c *C) saveTokens(data []byte) {
	var tokenDat TokenResponse
	json.Unmarshal(data, &tokenDat)
	file, err := json.MarshalIndent(Tokens{AccessToken: tokenDat.AccessToken, Expires: calcExpirationTime(tokenDat.ExpiresIn), RefreshToken: tokenDat.RefreshToken}, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("token.json", file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
func (c *C) RequestWithAuth(req *http.Request) (*http.Response, error) {
	client := http.Client{}
	tokens := c.ParseTokens()
	expires, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", tokens.Expires)
	if err != nil {
		log.Print(err)
	}
	if time.Now().After(expires) {
		c.RefreshToken()
		tokens = c.ParseTokens()
	}
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	return resp, err
}
func (c *C) FetchToken() {
	r := bufio.NewReader(os.Stdin)
	creds := c.ClientCred()
	ctx := c.ClientContext()
	data := FetchData{Scope: "https://www.googleapis.com/auth/youtube", ClientId: creds.ClientId}
	m, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{}
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/device/code", bytes.NewBuffer(m))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("User-Agent", ctx.Header.UserAgent)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	var dat RespData
	json.Unmarshal(body, &dat)
	fmt.Printf("Go to %s, and enter your user code: %s \n", dat.VerificationURL, dat.UserCode)
	fmt.Println("Press Enter afrer you do deeez...")
	r.ReadString('\n')
	tokenData, err := json.Marshal(TokenData{ClientId: creds.ClientId, ClientSecret: creds.ClientSecret, DeviceCode: dat.DeviceCode, GrantType: "urn:ietf:params:oauth:grant-type:device_code"})
	if err != nil {
		log.Fatal(err)
	}
	reqToken, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBuffer(tokenData))
	if err != nil {
		log.Fatal(err)
	}
	reqToken.Header.Add("Content-Type", "application/json")
	respToken, err := client.Do(reqToken)
	if err != nil {
		log.Fatal(err)
	}
	bodyToken, err := io.ReadAll(respToken.Body)
	respToken.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	c.saveTokens(bodyToken)
}
func (c *C) RefreshToken() {
	creds := c.ClientCred()
	tokens := c.ParseTokens()
	data, err := json.Marshal(RefreshTokenPayload{ClientId: creds.ClientId, ClientSecret: creds.ClientSecret, RefreshToken: tokens.RefreshToken})
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{}
	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyToken, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	c.saveTokens(bodyToken)
}
