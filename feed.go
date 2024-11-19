package main

import (
    "context"
    "encoding/xml"
    "net/http"
    "io"
    "html"
)

type RSSFeed struct {
    Channel struct {
        Title       string    `xml:"title"`
        Link        string    `xml:"link"`
        Description string    `xml:"description"`
        Item        []RSSItem `xml:"item"`
    } `xml:"channel"`
}

type RSSItem struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    Description string `xml:"description"`
    PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Add("User-Agent", "gator")
    client := &http.Client{}

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read the body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Create a feed struct to hold our data
    var feed RSSFeed

    // Parse the XML into our struct
    err = xml.Unmarshal(body, &feed)
    if err != nil {
        return nil, err
    }

    feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
    feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

    // Unescape HTML entities in each item
    for i := range feed.Channel.Item {
        feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
        feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
    }

    return &feed, nil
}
