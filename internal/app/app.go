package app

import "github.com/jakshi/rbg/internal/config"

type App struct {
	Config     *config.Config
	ConfigPath string
}

func NewApp(configPath string) (*App, error) {
	cfg, err := config.Read(configPath)
	if err != nil {
		return nil, err
	}
	return &App{
		Config:     cfg,
		ConfigPath: configPath,
	}, nil
}

func (a *App) SaveConfig() error {
	return config.Write(a.ConfigPath, a.Config)
}
