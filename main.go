package main

import (
	"blog/internal/config"
	"blog/internal/database"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
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
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollowFeed))
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))

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
