package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitForReady polls endpoint until it returns 200 or ctx/timeout expires.
func WaitForReady(ctx context.Context, timeout time.Duration, endpoint string) error {
	client := http.Client{Timeout: 2 * time.Second}
	start := time.Now()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("create ready request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(start) >= timeout {
				return fmt.Errorf("timeout waiting for %s", endpoint)
			}
			time.Sleep(250 * time.Millisecond)
		}
	}
}
