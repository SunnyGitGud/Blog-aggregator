package main

import (
	"blog/internal/database"
	"context"
	"fmt"
)

func middlewareLoggedIn(
	handler func(s *State, cmd Command, user database.User) error,
) func(s *State, cmd Command) error {
	return func(s *State, cmd Command) error {
		ctx := context.Background()

		user, err := s.db.GetUserByName(ctx, s.sStruct.CurrentUser)
		if err != nil {
			return fmt.Errorf("User not logged in: %w", err)
		}

		return handler(s, cmd, user)
	}
}
