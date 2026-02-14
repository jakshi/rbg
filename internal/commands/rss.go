package commands

import (
	"context"
	"encoding/xml"
	"errors"
	"html"
	"io"
	"net/http"

	"github.com/jakshi/rbg/internal/database"

	"github.com/jakshi/rbg/internal/app"
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

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req) // <-- executes the request
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body) // reads entire response body
	if err != nil {
		return nil, err
	}

	feed := &RSSFeed{}
	err = xml.Unmarshal(body, feed)
	if err != nil {
		return nil, err
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return feed, nil
}

func agg(app *app.App, args []string) error {
	ctx := context.Background()
	feedURLs := []string{
		"https://www.wagslane.dev/index.xml",
	}

	for _, url := range feedURLs {
		feed, err := fetchFeed(ctx, url)
		if err != nil {
			return err
		}

		println("Feed Title:", feed.Channel.Title)
		println("Feed Description:", feed.Channel.Description)
		for _, item := range feed.Channel.Item {
			println("  Item Title:", item.Title)
			println("  Item Link:", item.Link)
			println("  Item Description:", item.Description)
			println("  Item PubDate:", item.PubDate)
			println()
		}
	}

	return nil
}
func addFeed(app *app.App, args []string) error {
	if len(args) < 2 {
		return errors.New("usage: addfeed <feed name> <feed URL>")
	}
	feedName := args[0]
	feedURL := args[1]

	currentUser := app.Config.CurrentUserName
	if currentUser == "" {
		return errors.New("no user logged in, please login first")
	}

	ctx := context.Background()
	user, exists, err := getUserByName(ctx, app, currentUser)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user not found")
	}

	_, err = app.DB.CreateFeed(ctx, database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	println("Feed added successfully")

	feed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	println("Fetched Feed Title:", feed.Channel.Title)
	println("Fetched Feed Description:", feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		println("  Item Title:", item.Title)
		println("  Item Link:", item.Link)
		println("  Item Description:", item.Description)
		println("  Item PubDate:", item.PubDate)
		println()
	}

	return nil

}
