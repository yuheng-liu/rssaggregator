package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

// represents a RSSFeed item as shown in the website
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// converts xml in rssFeed into local struct
func urlToFeed(url string) (RSSFeed, error) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	// call to url and get raw response
	resp, err := httpClient.Get(url)
	if err != nil {
		return RSSFeed{}, err
	}
	// remember to close response at the end
	defer resp.Body.Close()
	// read body of response
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return RSSFeed{}, err
	}
	// unmarshal data bytes into rssFeed struct
	rssFeed := RSSFeed{}
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return RSSFeed{}, err
	}
	return rssFeed, nil
}
