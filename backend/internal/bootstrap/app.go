package bootstrap

import (
	"context"

	"github.com/luizf/go-chat/backend/internal/config"
)

// App is a thin wrapper kept for callers that already hold a loaded Config.
type App struct {
	cfg config.Config
}

func New(cfg config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context, getenv func(string) string, args []string) error {
	SetupLogging(a.cfg)
	if len(args) > 0 {
		switch args[0] {
		case "migrate":
			return RunMigrations(a.cfg)
		case "seed":
			return RunSeed(a.cfg)
		}
	}
	return serveHTTP(ctx, a.cfg)
}
