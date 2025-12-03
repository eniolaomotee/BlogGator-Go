package config

import (
	"context"
	"fmt"

	"github.com/eniolaomotee/BlogGator-Go/internal/database"
)

func MiddlewareLoggedIn(
	handler func(s *State, cmd Command, user database.User) error,
) func(*State, Command) error {
	return func(s *State, cmd Command) error {

		name := s.Conf.UserName

		currentUser, err := s.Db.GetUser(context.Background(), name)
		if err != nil {
			return fmt.Errorf("error getting user from database :%s", err)
		}

		return handler(s, cmd, currentUser)
	}

}

func ArgumentValidationMiddleware(handler func(s *State, cmd Command) error, expectedArgs int,
) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		if len(cmd.Args) != expectedArgs {
			return fmt.Errorf("usage: %s <expected %d args>", cmd.Name, expectedArgs)
		}
		return handler(s, cmd)
	}
}
