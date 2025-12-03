package config

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request %w", err)
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d %s", resp.StatusCode, resp.Status)

	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading data from response %w", err)
	}

	var rss RSSFeed

	err = xml.Unmarshal(data, &rss)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling xml %w", err)
	}

	for i := range rss.Channel.Item {
		rss.Channel.Item[i].Title = html.UnescapeString(rss.Channel.Item[i].Title)
		rss.Channel.Item[i].Description = html.UnescapeString(rss.Channel.Item[i].Description)
	}

	channel_desc := html.UnescapeString(rss.Channel.Description)
	channel_title := html.UnescapeString(rss.Channel.Title)

	rss.Channel.Description = channel_desc
	rss.Channel.Title = channel_title

	return &rss, nil

}
