package main

import (
	"fmt"
	"sync/atomic"
	"time"
)

type Stats struct {
	connected atomic.Int64
	sent      atomic.Int64
	received  atomic.Int64
	errors    atomic.Int64
}

func (s *Stats) ReportEvery(ctx <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx:
			s.print("final")
			return
		case <-ticker.C:
			s.print("progress")
		}
	}
}

func (s *Stats) print(label string) {
	fmt.Printf(
		"[%s] connected=%d sent=%d received=%d errors=%d\n",
		label,
		s.connected.Load(),
		s.sent.Load(),
		s.received.Load(),
		s.errors.Load(),
	)
}
