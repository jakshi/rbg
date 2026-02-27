package commands

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

// parsePubDate tries several layouts and returns sql.NullTime.
// If s is empty or unparsable, returns Valid=false and nil error.
func parsePubDate(s string) (sql.NullTime, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullTime{}, nil
	}

	// Common RSS date formats handled by net/http
	if t, err := http.ParseTime(s); err == nil {
		return sql.NullTime{Time: t, Valid: true}, nil
	}

	// Try RFC3339 and variants
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RubyDate,
		time.ANSIC,
		"Mon, 02 Jan 2006 15:04:05 MST", // explicit common layout
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return sql.NullTime{Time: t, Valid: true}, nil
		}
	}

	// Try unix timestamp (seconds)
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return sql.NullTime{Time: time.Unix(i, 0), Valid: true}, nil
	}

	// give up — treat as missing rather than failing
	return sql.NullTime{}, nil
}

func scrapeFeeds(ctx context.Context, app *app.App) error {
	feed, err := app.DB.GetNextFeedToFetch(ctx)
	if err != nil {
		// treat sql.ErrNoRows as "nothing to do" rather than a hard error
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	fetchedFeed, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("fetch feed: %w", err)
	}

	if err := app.DB.MarkFeedFetched(ctx, feed.ID); err != nil {
		return fmt.Errorf("mark fetched: %w", err)
	}

	println("Feed Title:", fetchedFeed.Channel.Title)
	println("Feed Description:", fetchedFeed.Channel.Description)

	for _, item := range fetchedFeed.Channel.Item {
		nt, err := parsePubDate(item.PubDate)
		if err != nil {
			fmt.Printf("parse date: %w", err)
		}
		_, err = app.DB.CreatePost(ctx, database.CreatePostParams{
			FeedID: feed.ID,
			Title:  item.Title,
			Url:    item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: nt,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Post already exists, skip it
				continue
			}
			fmt.Printf("create post: %w", err)
		}
	}

	return nil
}

func agg(app *app.App, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: agg <time_between_reqs>")
	}
	durStr := args[0]
	d, err := time.ParseDuration(durStr)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", durStr, err)
	}

	fmt.Printf("Starting aggregator with a delay of %s between requests...\n", d)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Shutting down aggregator...")
			return nil
		case <-ticker.C:
			if err := scrapeFeeds(ctx, app); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				fmt.Printf("scrape error: %v\n", err)
			}
		}
	}
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

func browse(app *app.App, args []string, user database.User) error {
	ctx := context.Background()

	// default values
	limit := 2

	// if first arg provided, parse as limit
	if len(args) >= 1 && args[0] != "" {
		if v, err := strconv.Atoi(args[0]); err == nil && v > 0 {
			limit = v
		} else if err != nil {
			fmt.Printf("invalid limit %q, using default %d\n", args[0], limit)
		}
	}

	posts, err := app.DB.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
		Offset: 0,
	})

	if err != nil {
		return err
	}

	if len(posts) == 0 {
		println("No posts found for your followed feeds.")
		return nil
	}

	println("Posts from your followed feeds:")
	for _, post := range posts {
		println("- " + post.Title)
		println("  Link:", post.Url)
		if post.Description.Valid {
			println("  Description:", post.Description.String)
		}
		if post.PublishedAt.Valid {
			println("  Published At:", post.PublishedAt.Time.Format(time.RFC1123))
		}
		println()
	}

	return nil
}
