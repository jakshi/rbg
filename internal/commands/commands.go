package commands

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
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

func login(app *app.App, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: login <username>")
	}
	app.Config.CurrentUserName = args[0]
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
func register(app *app.App, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: register <username>")
	}
	username := args[0]

	ctx := context.Background()

	// Check if user already exists
	_, err := app.DB.GetUser(ctx, username)
	if err == nil {
		return fmt.Errorf("user %s already exists", username)
	}

	// Create new user
	_, err = app.DB.CreateUser(app.DB.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Printf("User %s registered successfully\n", username)
	return nil
}
