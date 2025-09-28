package middlewareLogin

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/FazecatGit/Blog-Aggregator/internal/cmd"
	"github.com/FazecatGit/Blog-Aggregator/internal/database"
)

func MiddlewareLoggedIn(
	handler func(s *cmd.State, c cmd.Command, user database.User) error,
) func(*cmd.State, cmd.Command) error {
	return func(s *cmd.State, c cmd.Command) error {
		username := s.Config.CurrentUserName
		if username == "" {
			fmt.Fprintln(os.Stderr, "no user logged in")
			os.Exit(1)
		}

		user, err := s.DataBase.GetUserByName(context.Background(), username)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Fprintln(os.Stderr, "user not found")
				os.Exit(1)
			}
			return fmt.Errorf("could not fetch user: %w", err)
		}

		return handler(s, c, user)
	}
}
