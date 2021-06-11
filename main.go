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

type RedditRootNode struct {
	Kind string `json:"kind"`
	Data RedditDataNode `json:"data"`
}

type RedditDataNode struct {
	Modhash string `json:"modhash"`
	Dist int32 `json:"dist"`
	After string `json:"after"`
	Before string `json:"before"`
	Children []RedditChildNode `json:"children"`
}

type RedditChildNode struct {
	Kind string `json:"kind"`
	Data RedditChildDataNode `json:"data"`
}

type RedditChildDataNode struct {
	Subreddit string `json:"subreddit"`
	Selftext string `json:"selftext"`
	AuthorFullname string `json:"author_fullname"`
	ModReasonTitle string `json:"mod_reason_title"`
	Title string `json:"title"`
	Ups uint32 `json:"ups"`
	Downs uint32 `json:"downs"`
	TotalAwardsReceived uint32 `json:"total_awards_received"`
	Score uint32 `json:"score"`
	SelftextHtml string `json:"selftext_html"`
	Url string `json:"url"`
	Unknown map[string]interface{} `json:"-"`
}

const LIMIT uint8 = 50

func getWallStreetBets() []RedditChildNode {
	client := getRedditHttpClient()
	// time.Sleep(5 * time.Second)
	resp, _ := client.Get(fmt.Sprintf("https://oauth.reddit.com/r/wallstreetbets/hot?limit=%d", LIMIT))
	body, _ := io.ReadAll(resp.Body)
	hotJson := &RedditRootNode{}
	json.Unmarshal(body, hotJson)
	return hotJson.Data.Children
}

type WsbStockResults struct {
	MatchedShortUrls map[string][]string `json:"matched_short_urls"`
	UnmatchedUrls []string `json:"unmatched_urls"`
}

func getHtmlTagRegex() *regexp.Regexp {
	return regexp.MustCompile(`&lt;.+?&gt;`)
}

func removeHtmlTags(redditHtmlText string) string {
	return getHtmlTagRegex().ReplaceAllString(redditHtmlText, "")
}

func findShortsInWSB(shortInterestData []ShortInterestData, wsbData []RedditChildNode) WsbStockResults {
	result := WsbStockResults{
		MatchedShortUrls: make(map[string][]string),
		UnmatchedUrls: make([]string, 0),
	}
	matchingRegexes := getMatchingRegexes(shortInterestData)
	for _, childNode := range wsbData {
		atleastOne := false
		childNodeData := childNode.Data
		for index, regex := range matchingRegexes {
			shortInterestStock := shortInterestData[index]
			if regex.MatchString(childNodeData.Title) ||
				regex.MatchString(removeHtmlTags(childNodeData.SelftextHtml)) ||
				regex.MatchString(childNodeData.Selftext) ||
				regex.MatchString(childNodeData.ModReasonTitle) {
				result.MatchedShortUrls[shortInterestStock.Symbol] = append(result.MatchedShortUrls[shortInterestStock.Symbol], childNodeData.Url)
				atleastOne = true
			}
		}
		if (!atleastOne) {
			result.UnmatchedUrls = append(result.UnmatchedUrls, childNodeData.Url)
		}
	}
	return result
}

func getMatchingRegex(data ShortInterestData) *regexp.Regexp {
	result := regexp.MustCompile(fmt.Sprintf(`(?i)\b(?:%s|%s)\b`, data.Symbol, data.Name))
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

func main() {
	shortsData := getShortInterest()
	wsbData := getWallStreetBets()
	result := findShortsInWSB(shortsData, wsbData)
	prettyPrint, _ := json.MarshalIndent(result, "", "  ")
	log.Println(fmt.Sprintf("%s", prettyPrint))
}
