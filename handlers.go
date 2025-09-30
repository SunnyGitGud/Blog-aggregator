package main

import (
	"blog/internal/database"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

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

func handlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	ctx := context.Background()
	currentUser, err := s.db.GetUserByName(ctx, s.sStruct.CurrentUser)
	if err != nil {
		return fmt.Errorf("failed to find current user: %w", err)
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    uuid.NullUUID{UUID: currentUser.ID, Valid: true},
	}

	feed, err := s.db.CreateFeed(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: currentUser.ID,
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

func handlerFollow(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: follow <feed_url>")
	}
	url := cmd.Args[0]
	currUser, err := s.db.GetUserByName(context.Background(), s.sStruct.CurrentUser)
	if err != nil {
		return fmt.Errorf("Failed to retrive currUser :%w ", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: currUser.ID,
		Url:    url,
	})
	if err != nil {
		return fmt.Errorf("Failed to Create Folowfeed :%w", err)
	}

	fmt.Printf("%s is now following %s\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func handlerFollowing(s *State, cmd Command) error {
	currUser, err := s.db.GetUserByName(context.Background(), s.sStruct.CurrentUser)
	if err != nil {
		return fmt.Errorf("Failed to retrive currUser :%w ", err)
	}

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), currUser.ID)
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
