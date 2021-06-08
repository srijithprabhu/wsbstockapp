package main

import (
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

func getStockRegex() *regexp.Regexp {
	return regexp.MustCompile(`<tr>\s*<td[^>]*><a\s.*?href="(.*?)"[^>]*?>([A-Z]+)</a></td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*<td[^>]*>(.+?)</td>\s*</tr>`)
}

func main() {
	resp, err := http.Get("https://www.highshortinterest.com/all/")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	short_stock_body := string(body)
	log.Println(short_stock_body)

	regex := getStockRegex()
	submatches := regex.FindAllStringSubmatch(short_stock_body, -1)

	for _, submatch := range submatches {
		url := submatch[URL]
		symbol := submatch[SYMBOL]
		name := submatch[NAME]
		exchange := submatch[EXCHANGE]
		interest := submatch[INTEREST]
		floating := submatch[FLOATING]
		outstanding := submatch[OUTSTANDING]
		industry := submatch[INDUSTRY]
		log.Println(url, symbol, name, exchange, interest, floating, outstanding, industry)
	}
}
