package main

import (
	"blog/internal/config"
	"fmt"
	"log"
	"os"
)

type State struct {
	sStruct *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmap map[string]func(*State, Command) error
}

func (c *Commands) run(s *State, cmd Command) error {
	hander, ok := c.Cmap[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return hander(s, cmd)
}

func (c *Commands) register(name string, f func(*State, Command) error) {
	if c.Cmap == nil {
		c.Cmap = make(map[string]func(*State, Command) error)
	}
	c.Cmap[name] = f
}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("login handler expects a single argument: the username")
	}

	username := cmd.Args[0]
	if err := s.sStruct.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("user has been set")

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	state := &State{sStruct: cfg}

	cmds := &Commands{}
	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("usage: gator <command> [args]")
		os.Exit(1)
	}

	cmd := Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if err := cmds.run(state, cmd); err != nil {
		fmt.Println("Error :", err)
		os.Exit(1)
	}
}
