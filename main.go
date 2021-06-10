package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

const (
	_ = iota
	URL
	SYMBOL
	NAME
	EXCHANGE
	INTEREST
	FLOATING
	OUTSTANDING
	INDUSTRY
)

type ShortInterestData struct {
	URL         string
	Symbol      string
	Name        string
	Exchange    string
	Interest    string
	Floating    string
	Outstanding string
	Industry    string
}

func getStockRegex() *regexp.Regexp {
	return regexp.MustCompile(`<tr>\s*<td[^>]*><a\s.*?href="(.*?)"[^>]*?>([A-Z]+)</a></td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>([0-9\.]+?)%</td>\s*<td[^>]*>([0-9A-Za-z\.]+?)</td>\s*<td[^>]*>([0-9A-Za-z\.]+?)</td>\s*<td[^>]*>(.+?)</td>\s*</tr>`)
}

func getShortInterest() []ShortInterestData {
	resp, err := http.Get("https://www.highshortinterest.com/all/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	shortStockBody := string(body)

	regex := getStockRegex()
	submatches := regex.FindAllStringSubmatch(shortStockBody, -1)

	var shortsData []ShortInterestData

	for _, submatch := range submatches {
		data := ShortInterestData{
			submatch[URL],
			submatch[SYMBOL],
			submatch[NAME],
			submatch[EXCHANGE],
			submatch[INTEREST],
			submatch[FLOATING],
			submatch[OUTSTANDING],
			submatch[INDUSTRY],
		}
		shortsData = append(shortsData, data)
	}
	return shortsData
}

func getMatchingRegex(data ShortInterestData) *regexp.Regexp {
	result := regexp.MustCompile(fmt.Sprintf(`(?i)\b(?:%s|%s)\b`, data.Symbol, data.Name))
	return result
}

func findShortsInWSB(data []ShortInterestData) map[string][]string {
	result := make(map[string][]string)
	// matchingRegexes := getMatchingRegexes(data)

	return result
}

func getMatchingRegexes(data []ShortInterestData) []*regexp.Regexp {
	var result []*regexp.Regexp
	for _, element := range data {
		regex := getMatchingRegex(element)
		result = append(result, regex)
	}
	return result
}
// {"access_token": "18056296-dE8dyX3W7oTEmtCARM5xkK1ifhoyow", "token_type": "bearer", "expires_in": 3600, "scope": "*"}
type RedditAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn uint16 `json:"expires_in"`
	Scope string `json:"scope"`
}

func getRedditHttpClient() *http.Client {
	clientId := os.Getenv("REDDIT_CLIENT_ID")
	secret := os.Getenv("REDDIT_SECRET_TOKEN")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")

	authReq, _ := http.NewRequest("POST","https://www.reddit.com/api/v1/access_token", nil)
	authReq.SetBasicAuth(clientId, secret)
	authReq.Header.Add("User-Agent", "MyAuthBot/0.0.1")

	values := authReq.URL.Query()
	values.Add("grant_type", "password")
	values.Add("username", username)
	values.Add("password", password)
	authReq.URL.RawQuery = values.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	log.Println(authReq.RequestURI)
	res, _ := client.Do(authReq)
	defer res.Body.Close()
	authResp := &RedditAuthResponse{}
	result, _ := io.ReadAll(res.Body)
	json.Unmarshal(result, authResp)
	log.Println(authResp)

	client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport {
			Proxy: func (request *http.Request) (*url.URL, error) {
				request.Header.Set("Authorization", fmt.Sprintf("%s %s", authResp.TokenType, authResp.AccessToken))
				request.Header.Set("User-Agent", "MyBot/0.0.1")
				return nil, nil
			},
		},
	}

	return client
}

func getWallStreetBets() []string {
	result := make([]string, 0)
	client := getRedditHttpClient()
	time.Sleep(5 * time.Second)
	resp, _ := client.Get("https://oauth.reddit.com/r/wallstreetbets/hot")
	body, _ := io.ReadAll(resp.Body)
	log.Println(fmt.Sprintf("%s", body))
	return result
}

func main() {
	shortsData := getShortInterest()
	wsbData := getWallStreetBets()
	log.Println(shortsData)
	log.Println(wsbData)
	result := findShortsInWSB(shortsData)
	log.Println(result)
}
