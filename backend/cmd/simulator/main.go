package main

import (
	"context"
	"fmt"
	"os"

	"github.com/luizf/go-chat/backend/internal/simulator"
)

func main() {
	cfg := simulator.LoadConfig()
	sim := simulator.New(cfg)

	if err := sim.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "simulator failed: %v\n", err)
		os.Exit(1)
	}
}
