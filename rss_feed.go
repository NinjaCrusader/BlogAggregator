package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

//RSSFeed struct and methods

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

//RSSItem struct and methods

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

//start of helper functions

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	newRequest, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("there was an err creating new request: %v\n", err)
	}

	newRequest.Header.Add("User-Agent", "gator")

	client := &http.Client{}
	doRequest, err := client.Do(newRequest)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("there was an err getting reqeust information: %v\n", err)
	}
	defer doRequest.Body.Close()

	body, err := io.ReadAll(doRequest.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("there was an err reading request content: %v\n", err)
	}

	Feed := &RSSFeed{}

	err = xml.Unmarshal(body, Feed)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("there was an error Unmarshalling data: %v\n", err)
	}

	Feed.Channel.Title = html.UnescapeString(Feed.Channel.Title)
	Feed.Channel.Description = html.UnescapeString(Feed.Channel.Description)

	for i := 0; i < len(Feed.Channel.Item); i++ {
		Feed.Channel.Item[i].Title = html.UnescapeString(Feed.Channel.Item[i].Title)
		Feed.Channel.Item[i].Description = html.UnescapeString(Feed.Channel.Item[i].Description)
	}

	return Feed, nil
}
