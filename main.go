package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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
	url string
	symbol string
	name string
	exchange string
	interest string
	floating string
	outstanding string
	industry string
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
	result := regexp.MustCompile(fmt.Sprintf(`(?i)\b(?:%s|%s)\b`, data.symbol, data.name))
	return result
}

func findShortsInWSB(data []ShortInterestData) map[string][]string {
	result := make(map[string][]string)
	matchingRegexes := getMatchingRegexes(data)

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
	log.Println(shortsData)
	result := findShortsInWSB(shortsData)
	log.Println(result)
}
