package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/jakshi/rbg/internal/app"
	"github.com/jakshi/rbg/internal/database"
)

type Command struct {
	Description string
	Run         func(app *app.App, args []string) error
}

var AllCommands map[string]Command

// Add a sentinel error
var ErrNoCommand = errors.New("no command provided")

func All() map[string]Command {
	if AllCommands == nil {
		AllCommands = map[string]Command{
			"users":    {Description: "List users", Run: users},
			"login":    {Description: "Login user", Run: login},
			"help":     {Description: "Show help", Run: help},
			"register": {Description: "Register user", Run: register},
			"db-url":   {Description: "Print database URL", Run: dbURL},
		}
	}
	return AllCommands
}

func Run(app *app.App, args []string) error {
	if len(args) < 1 {
		help(app, nil)
		return ErrNoCommand
	}

	cmd, ok := All()[args[0]]
	if !ok {
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return cmd.Run(app, args[1:])
}

func SortedNames() []string {
	names := make([]string, 0, len(AllCommands))
	for name := range AllCommands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

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

// TODO: check that user is already registered
// and make that user check a separate function
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
	// Implement users functionality here
	return nil
}

func help(app *app.App, args []string) error {
	for _, name := range SortedNames() {
		fmt.Printf("  %s - %s\n", name, AllCommands[name].Description)
	}
	return nil
}
func dbURL(app *app.App, args []string) error {
	fmt.Print(app.Config.DBURL)
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
