package main

import (
	"blog/internal/database"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

func helperHTMLP(input string) string {
	doc, _ := html.Parse(strings.NewReader(input))
	var f func(*html.Node)
	var buf strings.Builder
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return buf.String()
}

func handlerBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Limit: int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedID)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", helperHTMLP(post.Description.String))
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func handlerFeeds(s *State, cmd Command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get all feed : %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("no feed found")
		return nil
	}

	fmt.Println("Feeds in database")
	for _, f := range feeds {
		fmt.Printf("-Name: %s\n, -URL: %s\n, -User: %s\n", f.FeedName, f.FeedUrl, f.UserName)
	}
	return nil
}

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	ctx := context.Background()

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
	}

	feed, err := s.db.CreateFeed(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("Failed to create feedFollow: %w", err)
	}

	fmt.Printf("Feed created:\nID: %s\nName: %s\nURL: %s\nUserID: %s\nCreatedAt: %s\nUpdatedAt: %s\n",
		feed.ID.String(),
		feed.Name,
		feed.Url,
		feed.UserID.UUID,
		feed.CreatedAt.String(),
		feed.UpdatedAt.String(),
	)

	return nil
}

func handlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}
	url := cmd.Args[0]

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("Failed to Create Folowfeed :%w", err)
	}

	fmt.Printf("%s is now following %s\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func handlerFollowing(s *State, cmd Command, user database.User) error {

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Failed to fetch feed Follow: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("This user does not Follow anyone")
		return nil
	}

	fmt.Println("Feeds Followed")
	for _, f := range follows {
		fmt.Printf("- %s\n", f.FeedName)
	}

	return nil
}

func handlerUnfollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: unfollow <feed_id>")
	}
	ctx := context.Background()

	feedID, err := s.db.GetFeedByURL(ctx, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid feed ID: %w", err)
	}

	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feedID.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	fmt.Println("Successfully unfollowed feed:", feedID)
	return nil
}

func handlerFetchRss(s *State, cmd Command) error {
	feedURL := "https://www.wagslane.dev/index.xml"

	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("Failed to fetch RssFeed: %w", err)
	}

	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerGetUsers(s *State, cmd Command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to get Users: %v\n", err)
	}

	for _, u := range users {
		if u == s.sStruct.CurrentUser {
			fmt.Printf("* %s (current)\n", u)
		} else {
			fmt.Printf("* %s\n", u)
		}
	}
	return nil
}

func handlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	err := s.db.DeleteAllUsers(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to reset users: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All users have been successfully removed.")
	os.Exit(0)
	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Register handler expects a single argument: the username")
	}
	username := cmd.Args[0]

	// Check if user already exists
	existingUser, err := s.db.GetUserByName(context.Background(), username)

	if err == nil {
		// User was found successfully, meaning they already exist
		return fmt.Errorf("user %s already exists (ID: %s)", username, existingUser.ID)
	}

	if err != sql.ErrNoRows {
		// Some other database error occurred
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	// err == sql.ErrNoRows means user doesn't exist, so we can create them
	fmt.Printf("User %s doesn't exist, creating...\n", username)

	now := time.Now()
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	fmt.Printf("User created successfully: %s\n", user.Name)
	fmt.Printf("User ID: %s\n", user.ID)
	fmt.Printf("Created at: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	// Set the newly created user as current user (persist to config)
	if err := s.sStruct.SetUser(username); err != nil {
		return fmt.Errorf("user created but failed to set as current user: %w", err)
	}
	fmt.Printf("Current user set to: %s\n", username)

	return nil
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("login handler expects a single argument: the username")
	}
	username := cmd.Args[0]

	// Check if user exists before logging in
	user, err := s.db.GetUserByName(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user %s not found. Please register first", username)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Set the user in config
	if err := s.sStruct.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("Successfully logged in as: %s (ID: %s)\n", user.Name, user.ID)
	return nil
}

func handlerWhoami(s *State, cmd Command) error {
	if s.sStruct.CurrentUser == "" {
		fmt.Println("No user currently logged in")
		return nil
	}
	fmt.Printf("Current user: %s\n", s.sStruct.CurrentUser)
	return nil
}
