package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	bots := flag.Int("bots", envInt("BOTS", 100), "number of concurrent bots")
	interval := flag.Duration("interval", envDuration("INTERVAL", time.Second), "delay between messages per bot")
	duration := flag.Duration("duration", envDuration("DURATION", 3*time.Minute), "test duration (0 = until Ctrl+C)")
	ramp := flag.Duration("ramp", envDuration("RAMP", 10*time.Second), "ramp-up period for bot connections")
	apiURL := flag.String("api", envString("API_URL", "http://localhost:8080/api/v1"), "REST API base URL")
	wsURL := flag.String("ws", envString("WS_URL", "ws://localhost:8080/ws/connect"), "WebSocket URL")
	peerUsername := flag.String("peer", envString("PEER_USERNAME", "bob"), "username of the chat peer")
	setupEmail := flag.String("setup-email", envString("SETUP_EMAIL", "alice@example.com"), "user for peer lookup")
	setupPassword := flag.String("setup-password", envString("SETUP_PASSWORD", "password123"), "password for peer lookup")
	botPassword := flag.String("bot-password", envString("BOT_PASSWORD", "loadtest123"), "password for generated bots")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	api := NewAPIClient(*apiURL)
	stats := &Stats{}

	peerID, err := resolvePeerID(api, *setupEmail, *setupPassword, *peerUsername)
	if err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("peer %s id=%s bots=%d interval=%s ramp=%s duration=%s\n",
		*peerUsername, peerID, *bots, *interval, *ramp, *duration)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *duration > 0 {
		timerCtx, timerCancel := context.WithTimeout(ctx, *duration)
		defer timerCancel()
		ctx = timerCtx
	}

	go stats.ReportEvery(ctx.Done(), 5*time.Second)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < *bots; i++ {
		delay := time.Duration(0)
		if *bots > 1 && *ramp > 0 {
			delay = time.Duration(i) * (*ramp / time.Duration(*bots-1))
		}

		wg.Add(1)
		go func(id int, startDelay time.Duration) {
			defer wg.Done()
			time.Sleep(startDelay)

			bot := &Bot{
				id:       id,
				email:    fmt.Sprintf("wsbot-%04d@loadtest.local", id),
				username: fmt.Sprintf("wsbot%04d", id),
				password: *botPassword,
				api:      api,
				wsURL:    *wsURL,
				peerID:   peerID,
				interval: *interval,
				stats:    stats,
			}
			bot.Run(ctx)
		}(i, delay)
	}

	wg.Wait()
	stats.print("done")
}

func resolvePeerID(api *APIClient, email, password, username string) (string, error) {
	token, err := api.Login(email, password)
	if err != nil {
		return "", err
	}
	return api.SearchUser(token, username)
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

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if v == "0" {
			return 0
		}
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}
