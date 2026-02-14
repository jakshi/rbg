package commands

import (
	"context"
	"fmt"

	"github.com/jakshi/rbg/internal/app"
)

func dbURL(app *app.App, args []string) error {
	fmt.Print(app.Config.DBURL)
	return nil
}

func reset(app *app.App, args []string) error {
	ctx := context.Background()

	err := app.DB.DeleteAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}

	fmt.Println("Database reset successfully")
	return nil
}
