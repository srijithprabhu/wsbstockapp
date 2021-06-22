package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"sort"
	"strings"
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
	resp, _ := http.Get(HighShortInterestEndpoint)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
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

const HighShortInterestEndpoint = "https://www.highshortinterest.com/all/"
const RedditAccessTokenEndpoint = "https://www.reddit.com/api/v1/access_token"
const RedditWsbEndpoint = "https://oauth.reddit.com/r/wallstreetbets/top"
const RedditWsbTimeFilter string = "day"
const RedditWsbResultLimit uint8 = 50

func getRedditHttpClient(params map[string]interface{}) *http.Client {
	clientId := params["reddit_client_id"].(string)
	secret := params["reddit_secret_token"].(string)
	username := params["reddit_username"].(string)
	password := params["reddit_password"].(string)

	authReq, _ := http.NewRequest("POST", RedditAccessTokenEndpoint, nil)
	authReq.SetBasicAuth(clientId, secret)
	authReq.Header.Add("User-Agent", "MyAuthBot/0.0.1")

	values := authReq.URL.Query()
	values.Add("grant_type", "password")
	values.Add("username", username)
	values.Add("password", password)
	authReq.URL.RawQuery = values.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	res, _ := client.Do(authReq)
	defer res.Body.Close()
	authResp := &RedditAuthResponse{}
	result, _ := ioutil.ReadAll(res.Body)
	json.Unmarshal(result, authResp)

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
	Ups int `json:"ups"`
	Downs int `json:"downs"`
	TotalAwardsReceived uint32 `json:"total_awards_received"`
	Score uint32 `json:"score"`
	SelftextHtml string `json:"selftext_html"`
	Url string `json:"url"`
	Permalink string `json:"permalink"`
	Unknown map[string]interface{} `json:"-"`
}

func getWallStreetBets(params map[string]interface{}) []RedditChildNode {
	client := getRedditHttpClient(params)
	resp, _ := client.Get(fmt.Sprintf("%s?limit=%d&t=%s", RedditWsbEndpoint, RedditWsbResultLimit, RedditWsbTimeFilter))
	body, _ := ioutil.ReadAll(resp.Body)
	hotJson := &RedditRootNode{}
	json.Unmarshal(body, hotJson)
	return hotJson.Data.Children
}

type WsbSummary struct {
	Url string `json:"url"`
	Title string `json:"title"`
	Upvotes int `json:"upvotes"`
	MatchingShorts []string `json:"matching_shorts"`
	Text string `json:"text"`
	Permalink string `json:"permalink"`
}

type WsbStockResults struct {
	MatchedShortUrls []WsbSummary `json:"matched_short_urls"`
	AdditionalInterestingUrls []WsbSummary `json:"additional_interesting_urls"`
	UnmatchedUrls []WsbSummary `json:"unmatched_urls"`
}

func getHtmlTagRegex() *regexp.Regexp {
	return regexp.MustCompile(`&lt;.+?&gt;`)
}

func removeHtmlTags(redditHtmlText string) string {
	return getHtmlTagRegex().ReplaceAllString(redditHtmlText, "")
}

func findShortsInWSB(shortInterestData []ShortInterestData, wsbData []RedditChildNode) WsbStockResults {
	result := WsbStockResults{
		MatchedShortUrls: make([]WsbSummary, 0),
		UnmatchedUrls: make([]WsbSummary, 0),
	}
	matchingRegexes := getMatchingRegexes(shortInterestData)
	interestingRegex := getInterestingRedditRegex()
	for _, childNode := range wsbData {
		atleastOne := false
		childNodeData := childNode.Data
		summary := WsbSummary{childNodeData.Url, childNodeData.Title, childNodeData.Ups, make([]string, 0), childNodeData.Selftext, childNodeData.Permalink}
		for index, regex := range matchingRegexes {
			if matchesRegex(regex, childNodeData) {
				summary.MatchingShorts = append(summary.MatchingShorts, shortInterestData[index].Symbol)
				atleastOne = true
			}
		}
		if atleastOne {
			result.MatchedShortUrls = append(result.MatchedShortUrls, summary)
		} else {
			if matchesRegex(interestingRegex, childNodeData) {
				result.AdditionalInterestingUrls = append(result.AdditionalInterestingUrls, summary)
			} else {
				result.UnmatchedUrls = append(result.UnmatchedUrls, summary)
			}
		}
	}
	return result
}

func matchesRegex(regex *regexp.Regexp, childNodeData RedditChildDataNode) bool {
	return regex.MatchString(childNodeData.Title) ||
		regex.MatchString(removeHtmlTags(childNodeData.SelftextHtml)) ||
		regex.MatchString(childNodeData.Selftext) ||
		regex.MatchString(childNodeData.ModReasonTitle)
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

func getInterestingRedditRegex() *regexp.Regexp {
	interestingRedditWords := []string{
		"rocket",
		"moon",
		"short",
		"hold",
		"holding",
		"hodl",
		"hodling",
		"green",
		"suppress",
		"suppressed",
		"meme",
		"undervalue",
		"under value",
		"undervalued",
		"under valued",
		"bounce",
		"ape",
		"apes",
		"yolo",
		"🚀",
		"💎",
		"🙌",
		"🦍",
		"🐵",
		"💪",
		"🤚🏾",
		"✋🏾",
	}
	result := regexp.MustCompile(fmt.Sprintf(`(?i)(?:%s)`, strings.Join(interestingRedditWords, "|")))
	return result
}

func aggregateListHtml(summaries []WsbSummary) string {
	sort.Slice(summaries, func(i, j int) bool {
		// reverse sort on length of matching shorts
		var diff = summaries[i].Upvotes - summaries[j].Upvotes
		return diff > 0
	})
	result := "<ul>\n"
	for _ , summary := range summaries {
		postfix := fmt.Sprintf("<li><a href=\"%s\">%s</a>(Shorts: %s)(%d Upvotes)</li>\n",
			fmt.Sprintf("https://www.reddit.com%s", summary.Permalink),
			summary.Title,
			strings.Join(summary.MatchingShorts,","),
			summary.Upvotes,
		)
		result = result + postfix
	}
	result = result + "</ul>"
	return result
}

func generateEmailHtml(params map[string]interface{}, results WsbStockResults) string {
	result := fmt.Sprintf("<h3>High Short Interest stock discussions</h3>\n%s\n", aggregateListHtml(results.MatchedShortUrls))
	result = result + fmt.Sprintf("<h3>Other interesting discussions</h3>\n%s\n", aggregateListHtml(results.AdditionalInterestingUrls))
	result = result + fmt.Sprintf("<h3>Miscellaneous Discussion (Hopefully)</h3>\n%s\n", aggregateListHtml(results.UnmatchedUrls))
	return result
}

func sendEmail(params map[string]interface{}, emailHtml string) {
	username := params["email_address"].(string)
	password := params["email_password"].(string)
	server := params["email_smtp_host"].(string)
	port := params["email_smtp_port"].(string)

	var to []string
	for _, emailAddress := range params["email_addresses"].([]interface{}) {
		to = append(to, emailAddress.(string))
	}

	date := time.Now()

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("Subject: WallstreetBets Results for %v\n%s\n%s", date.Format("Mon Jan 02 2006"), mime, emailHtml)

	smtpAddress := fmt.Sprintf("%s:%s", server, port)
	auth := smtp.PlainAuth("", username, password, server)
	err := smtp.SendMail(smtpAddress, auth, username, to, []byte(body))

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func Main(params map[string]interface{}) map[string]interface{} {
	shortsData := getShortInterest()
	wsbData := getWallStreetBets(params)
	wsbStockResult := findShortsInWSB(shortsData, wsbData)
	emailHtml := generateEmailHtml(params, wsbStockResult)
	sendEmail(params, emailHtml)
	return map[string]interface{}{"status":"ok"}
}