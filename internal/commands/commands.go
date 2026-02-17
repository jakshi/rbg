package commands

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jakshi/rbg/internal/app"
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
			"users":     {Description: "List users", Run: users},
			"login":     {Description: "Login user", Run: login},
			"help":      {Description: "Show help", Run: help},
			"register":  {Description: "Register user", Run: register},
			"db-url":    {Description: "Print database URL", Run: dbURL},
			"reset":     {Description: "Reset the database", Run: reset},
			"agg":       {Description: "Aggregator service", Run: agg},
			"addfeed":   {Description: "Add a new feed", Run: addFeed},
			"feeds":     {Description: "List all feeds", Run: listFeeds},
			"follow":    {Description: "Follow a feed", Run: followFeed},
			"following": {Description: "List followed feeds", Run: listFollowing},
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

func help(app *app.App, args []string) error {
	for _, name := range SortedNames() {
		fmt.Printf("  %s - %s\n", name, AllCommands[name].Description)
	}
	return nil
}
