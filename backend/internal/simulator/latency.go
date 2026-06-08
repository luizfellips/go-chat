package simulator

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const latencyPrefix = "sim:"

type pendingMessage struct {
	sentAt   time.Time
	senderID string
}

type Metrics struct {
	sent            atomic.Int64
	delivered       atomic.Int64
	roundTrips      atomic.Int64
	errors          atomic.Int64
	connected       atomic.Int64
	pending         sync.Map
	deliverySamples []time.Duration
	roundTripSamples []time.Duration
	mu              sync.Mutex
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) TrackSend(key, senderID string) {
	m.pending.Store(key, pendingMessage{sentAt: time.Now(), senderID: senderID})
}

func (m *Metrics) TrackReceive(content, receiverID string) {
	raw, ok := m.pending.Load(content)
	if !ok {
		return
	}
	pm := raw.(pendingMessage)
	latency := time.Since(pm.sentAt)

	m.mu.Lock()
	defer m.mu.Unlock()

	if receiverID == pm.senderID {
		m.roundTrips.Add(1)
		m.roundTripSamples = append(m.roundTripSamples, latency)
		return
	}

	m.delivered.Add(1)
	m.deliverySamples = append(m.deliverySamples, latency)
	m.pending.Delete(content)
}

func (m *Metrics) RecordSent()     { m.sent.Add(1) }
func (m *Metrics) RecordError()    { m.errors.Add(1) }
func (m *Metrics) ConnUp()         { m.connected.Add(1) }
func (m *Metrics) ConnDown()       { m.connected.Add(-1) }

func (m *Metrics) Snapshot() Report {
	m.mu.Lock()
	defer m.mu.Unlock()

	return Report{
		Connected:       m.connected.Load(),
		Sent:            m.sent.Load(),
		Delivered:       m.delivered.Load(),
		RoundTrips:      m.roundTrips.Load(),
		Errors:          m.errors.Load(),
		DeliveryLatency: summarize(m.deliverySamples),
		RoundTripLatency: summarize(m.roundTripSamples),
	}
}

type LatencySummary struct {
	Count int
	Avg   time.Duration
	Min   time.Duration
	P50   time.Duration
	P95   time.Duration
	P99   time.Duration
	Max   time.Duration
}

type Report struct {
	Connected        int64
	Sent             int64
	Delivered        int64
	RoundTrips       int64
	Errors           int64
	DeliveryLatency  LatencySummary
	RoundTripLatency LatencySummary
}

func (r Report) Print(label string) {
	fmt.Printf("[%s] connected=%d sent=%d delivered=%d round_trips=%d errors=%d\n",
		label, r.Connected, r.Sent, r.Delivered, r.RoundTrips, r.Errors)
	printLatency("delivery", r.DeliveryLatency)
	printLatency("round_trip", r.RoundTripLatency)
}

func printLatency(name string, s LatencySummary) {
	if s.Count == 0 {
		fmt.Printf("  %s_latency: no samples\n", name)
		return
	}
	fmt.Printf("  %s_latency: count=%d avg=%s p50=%s p95=%s p99=%s max=%s\n",
		name, s.Count, s.Avg.Round(time.Millisecond), s.P50.Round(time.Millisecond),
		s.P95.Round(time.Millisecond), s.P99.Round(time.Millisecond), s.Max.Round(time.Millisecond))
}

func summarize(samples []time.Duration) LatencySummary {
	if len(samples) == 0 {
		return LatencySummary{}
	}

	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, d := range sorted {
		total += d
	}

	return LatencySummary{
		Count: len(sorted),
		Avg:   total / time.Duration(len(sorted)),
		Min:   sorted[0],
		P50:   percentile(sorted, 50),
		P95:   percentile(sorted, 95),
		P99:   percentile(sorted, 99),
		Max:   sorted[len(sorted)-1],
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) - 1) * p / 100
	return sorted[idx]
}
