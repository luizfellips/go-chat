package simulator

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	Users             int
	Conversations     int
	MessagesPerSecond float64
	Duration          time.Duration
	Ramp              time.Duration
	APIURL            string
	WSURL             string
	Password          string
}

func LoadConfig() Config {
	users := flag.Int("users", envInt("SIM_USERS", 100), "number of fake users")
	conversations := flag.Int("conversations", envInt("SIM_CONVERSATIONS", 20), "number of 1:1 conversations")
	mps := flag.Float64("messages-per-second", envFloat("SIM_MESSAGES_PER_SECOND", 50), "global message send rate")
	duration := flag.Duration("duration", envDuration("SIM_DURATION", 5*time.Minute), "simulation duration")
	ramp := flag.Duration("ramp", envDuration("SIM_RAMP", 30*time.Second), "ramp-up for websocket connections")
	apiURL := flag.String("api", envString("API_URL", "http://localhost:8080/api/v1"), "REST API base URL")
	wsURL := flag.String("ws", envString("WS_URL", "ws://localhost:8080/ws/connect"), "WebSocket URL")
	password := flag.String("password", envString("SIM_PASSWORD", "loadtest123"), "password for fake users")
	flag.Parse()

	cfg := Config{
		Users:             *users,
		Conversations:     *conversations,
		MessagesPerSecond: *mps,
		Duration:          *duration,
		Ramp:              *ramp,
		APIURL:            *apiURL,
		WSURL:             *wsURL,
		Password:          *password,
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func (c Config) Validate() error {
	if c.Users < 2 {
		return fmt.Errorf("users must be >= 2")
	}
	if c.Conversations < 1 {
		return fmt.Errorf("conversations must be >= 1")
	}
	if c.Conversations*2 > c.Users {
		return fmt.Errorf("conversations*2 (%d) exceeds users (%d)", c.Conversations*2, c.Users)
	}
	if c.MessagesPerSecond <= 0 {
		return fmt.Errorf("messages-per-second must be > 0")
	}
	if c.Duration <= 0 {
		return fmt.Errorf("duration must be > 0")
	}
	return nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		var n float64
		if _, err := fmt.Sscanf(v, "%f", &n); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
