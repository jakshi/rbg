package commands

import (
	"context"
	"errors"

	"github.com/jakshi/rbg/internal/app"
	"github.com/jakshi/rbg/internal/database"
)

// middlewareLoggedIn wraps a handler that expects a logged-in user,
// and returns a regular command handler.
func middlewareLoggedIn(handler func(a *app.App, args []string, user database.User) error) func(a *app.App, args []string) error {
	return func(a *app.App, args []string) error {
		username := a.Config.CurrentUserName
		if username == "" {
			return errors.New("no user logged in, please login first")
		}

		ctx := context.Background()
		user, exists, err := getUserByName(ctx, a, username) // reuse helper in users.go
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("user not found")
		}

		return handler(a, args, user)
	}
}
