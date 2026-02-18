package commands

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
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

func listFeeds(app *app.App, args []string) error {
	ctx := context.Background()
	feeds, err := app.DB.ListFeeds(ctx)
	if err != nil {
		return err
	}

	if len(feeds) == 0 {
		println("No feeds found.")
		return nil
	}

	println("Your Feeds:")
	for _, feed := range feeds {
		user, err := app.DB.GetUser(ctx, feed.UserID)
		if err != nil {
			return err
		}
		println("- " + feed.Name + " (" + feed.Url + ")")
		println("  Added by:", user.Name)
	}

	return nil
}

func addFeed(app *app.App, args []string, user database.User) error {
	if len(args) < 2 {
		return errors.New("usage: addfeed <feed name> <feed URL>")
	}
	feedName := args[0]
	feedURL := args[1]

	ctx := context.Background()
	createdFeed, err := app.DB.CreateFeed(ctx, database.CreateFeedParams{
		Name:   feedName,
		Url:    feedURL,
		UserID: user.ID,
	})
	if err != nil {
		return err
	}

	println("Feed added successfully")

	fetched, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	println("Fetched Feed Title:", fetched.Channel.Title)
	println("Fetched Feed Description:", fetched.Channel.Description)
	for _, item := range fetched.Channel.Item {
		println("  Item Title:", item.Title)
		println("  Item Link:", item.Link)
		println("  Item Description:", item.Description)
		println("  Item PubDate:", item.PubDate)
		println()
	}

	// Create a feed follow for the user who added the feed
	_, err = app.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: createdFeed.ID,
	})
	if err != nil {
		return err
	}

	println("You are now following the feed")

	return nil
}

func followFeed(app *app.App, args []string, user database.User) error {
	if len(args) < 1 {
		return errors.New("usage: follow <feed URL>")
	}
	feedURL := args[0]

	ctx := context.Background()
	feed, err := app.DB.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return err
	}

	result, err := app.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println("Feed Title:", result.FeedName)
	fmt.Println("User", user.Name, "is now following the feed")

	return nil
}

func listFollowing(app *app.App, args []string, user database.User) error {
	ctx := context.Background()

	follows, err := app.DB.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return err
	}

	if len(follows) == 0 {
		println("You are not following any feeds.")
		return nil
	}

	println("Feeds you are following:")
	for _, follow := range follows {
		println("- " + follow.FeedName + " (" + follow.FeedUrl + ")")
	}

	return nil
}

func unfollowFeed(app *app.App, args []string, user database.User) error {
	if len(args) < 1 {
		return errors.New("usage: unfollow <feed URL>")
	}
	feedURL := args[0]

	ctx := context.Background()
	feed, err := app.DB.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return err
	}

	err = app.DB.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return err
	}

	fmt.Println("User", user.Name, "has unfollowed the feed")

	return nil
}
