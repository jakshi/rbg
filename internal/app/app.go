package app

import (
	"database/sql"

	"github.com/jakshi/rbg/internal/config"
	"github.com/jakshi/rbg/internal/database"

	_ "github.com/lib/pq"
)

type App struct {
	DB         *database.Queries
	Config     *config.Config
	ConfigPath string
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.Read(configPath)
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return nil, err
	}

	return &App{
		DB:         database.New(conn),
		Config:     cfg,
		ConfigPath: configPath,
	}, nil
}

func (a *App) SaveConfig() error {
	return config.Write(a.ConfigPath, a.Config)
}
