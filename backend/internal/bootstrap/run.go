package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/luizf/go-chat/backend/internal/config"
)

// Run is the application entry used by main and integration tests.
// Pass os.Getenv for production; tests can supply a custom getenv to control config in parallel.
func Run(ctx context.Context, getenv func(string) string, args []string) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}
	SetupLogging(cfg)

	if len(args) > 0 {
		switch args[0] {
		case "migrate":
			return RunMigrations(cfg)
		case "seed":
			return RunSeed(cfg)
		}
	}

	return serveHTTP(ctx, cfg)
}

// RunningServer is a started HTTP server for integration tests.
type RunningServer struct {
	BaseURL string
}

// StartHTTP wires dependencies, listens on cfg.Port (use "0" for an ephemeral port),
// and blocks until the server is ready or ctx is cancelled.
func StartHTTP(ctx context.Context, getenv func(string) string) (*RunningServer, error) {
	cfg, err := config.Load(getenv)
	if err != nil {
		return nil, err
	}
	SetupLogging(cfg)

	deps, listener, err := wireAndListen(ctx, cfg)
	if err != nil {
		return nil, err
	}

	baseURL := "http://" + listener.Addr().String()
	if err := WaitForReady(ctx, 15*time.Second, baseURL+"/ready"); err != nil {
		_ = listener.Close()
		deps.Pool.Close()
		deps.Tickets.Stop()
		return nil, err
	}

	go func() {
		if err := runUntilDone(ctx, cfg, deps, listener); err != nil {
			log.Error().Err(err).Msg("test server stopped with error")
		}
	}()

	return &RunningServer{BaseURL: baseURL}, nil
}

func serveHTTP(ctx context.Context, cfg config.Config) error {
	deps, listener, err := wireAndListen(ctx, cfg)
	if err != nil {
		return err
	}
	return runUntilDone(ctx, cfg, deps, listener)
}

func wireAndListen(ctx context.Context, cfg config.Config) (*Dependencies, net.Listener, error) {
	deps, err := Wire(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("", cfg.Port))
	if err != nil {
		deps.Pool.Close()
		deps.Tickets.Stop()
		return nil, nil, fmt.Errorf("listen: %w", err)
	}

	return deps, listener, nil
}

func runUntilDone(ctx context.Context, cfg config.Config, deps *Dependencies, listener net.Listener) error {
	defer deps.Pool.Close()
	defer deps.Tickets.Stop()

	go deps.Hub.Run()

	srv := &http.Server{
		Handler:      deps.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", listener.Addr().String()).Msg("server started")
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	shutdown := func() error {
		log.Info().Msg("shutting down")
		deps.Hub.Shutdown()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-ctx.Done():
		return shutdown()
	case <-quit:
		return shutdown()
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}
}
