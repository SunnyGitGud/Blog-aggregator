package main

import (
	"blog/internal/config"
	"blog/internal/database"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type State struct {
	sStruct *config.Config
	db      *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmap map[string]func(*State, Command) error
}

func (c *Commands) run(s *State, cmd Command) error {
	handler, ok := c.Cmap[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	if c.Cmap == nil {
		c.Cmap = make(map[string]func(*State, Command) error)
	}
	c.Cmap[name] = f
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

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create the database queries instance
	dbQueries := database.New(db)

	// Create state with both config and database queries
	state := &State{
		sStruct: cfg,
		db:      dbQueries,
	}

	cmds := &Commands{}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("whoami", handlerWhoami)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	if len(os.Args) < 2 {
		fmt.Println("usage: gator <command> [args]")
		os.Exit(1)
	}

	cmd := Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmds.run(state, cmd); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
