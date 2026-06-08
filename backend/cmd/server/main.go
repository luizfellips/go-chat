package main

import (
	"context"
	"fmt"
	"os"

	"github.com/luizf/go-chat/backend/internal/bootstrap"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string, args []string) error {
	return bootstrap.Run(ctx, getenv, args)
}
