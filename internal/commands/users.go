package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/jakshi/rbg/internal/app"
	"github.com/jakshi/rbg/internal/database"
)

// Returns (user, exists, error)
func getUserByName(ctx context.Context, a *app.App, name string) (database.User, bool, error) {
	user, err := a.DB.GetUserByName(ctx, name)
	if err == nil {
		return user, true, nil // user exists
	}
	if err == sql.ErrNoRows {
		return database.User{}, false, nil // user doesn't exist
	}
	return database.User{}, false, fmt.Errorf("db error: %w", err) // actual error
}

func login(app *app.App, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: login <username>")
	}
	username := args[0]
	app.Config.CurrentUserName = username

	ctx := context.Background()
	_, exists, err := getUserByName(ctx, app, username)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user %q does not exist, please register first", username)
	}

	if err := app.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}
	fmt.Printf("Logged in as %s\n", args[0])

	return nil
}

func users(app *app.App, args []string) error {
	ctx := context.Background()
	users, err := app.DB.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Println("Registered users:")
	for _, user := range users {
		if user.Name == app.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

func register(app *app.App, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: register <username>")
	}
	username := args[0]

	ctx := context.Background()

	// Check if user already exists
	_, exists, err := getUserByName(ctx, app, username)
	if err != nil {
		return err
	}
	if exists {
		fmt.Fprintf(os.Stderr, "user %q already exists\n", username)
		os.Exit(1)
	}

	// Create new user
	_, err = app.DB.CreateUser(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Printf("User %s registered successfully\n", username)
	return login(app, args)
}
