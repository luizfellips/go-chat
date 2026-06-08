package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

type Simulator struct {
	cfg     Config
	api     *APIClient
	users   []*User
	conns   map[int]*websocket.Conn
	connMu  sync.Mutex
	metrics *Metrics
}

func New(cfg Config) *Simulator {
	return &Simulator{
		cfg:     cfg,
		api:     NewAPIClient(cfg.APIURL),
		metrics: NewMetrics(),
		conns:   make(map[int]*websocket.Conn),
	}
}

func (s *Simulator) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	if s.cfg.Duration > 0 {
		timed, timedCancel := context.WithTimeout(ctx, s.cfg.Duration)
		defer timedCancel()
		ctx = timed
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	fmt.Printf("simulator users=%d conversations=%d mps=%.0f duration=%s ramp=%s\n",
		s.cfg.Users, s.cfg.Conversations, s.cfg.MessagesPerSecond, s.cfg.Duration, s.cfg.Ramp)

	if err := s.setupUsers(); err != nil {
		return fmt.Errorf("setup users: %w", err)
	}
	if err := s.setupConversations(); err != nil {
		return fmt.Errorf("setup conversations: %w", err)
	}

	go s.reportProgress(ctx)

	if err := s.connectAll(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	s.runTraffic(ctx)
	s.closeAll()

	s.metrics.Snapshot().Print("done")
	return nil
}

func (s *Simulator) setupUsers() error {
	fmt.Printf("creating %d users...\n", s.cfg.Users)
	s.users = make([]*User, s.cfg.Users)

	for i := 0; i < s.cfg.Users; i++ {
		u := &User{
			Index:    i,
			Email:    fmt.Sprintf("simuser-%04d@loadtest.local", i),
			Username: fmt.Sprintf("simuser%04d", i),
			Password: s.cfg.Password,
		}
		if err := s.api.Register(u.Email, u.Username, u.Password); err != nil {
			return err
		}
		token, userID, err := s.api.Login(u.Email, u.Password)
		if err != nil {
			return err
		}
		u.Token = token
		u.UserID = userID
		s.users[i] = u
	}
	return nil
}

func (s *Simulator) setupConversations() error {
	fmt.Printf("creating %d conversations...\n", s.cfg.Conversations)

	for c := 0; c < s.cfg.Conversations; c++ {
		a := s.users[c*2]
		b := s.users[c*2+1]

		convID, err := s.api.CreateConversation(a.Token, b.UserID)
		if err != nil {
			return fmt.Errorf("conversation %d: %w", c, err)
		}

		a.ConvID = convID
		a.PeerID = b.UserID
		b.ConvID = convID
		b.PeerID = a.UserID
	}
	return nil
}

func (s *Simulator) connectAll(ctx context.Context) error {
	fmt.Printf("opening %d websocket connections...\n", len(s.users))

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i, u := range s.users {
		delay := time.Duration(0)
		if len(s.users) > 1 && s.cfg.Ramp > 0 {
			delay = time.Duration(i) * (s.cfg.Ramp / time.Duration(len(s.users)-1))
		}

		wg.Add(1)
		go func(user *User, startDelay time.Duration) {
			defer wg.Done()

			timer := time.NewTimer(startDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}

			conn, err := user.Connect(ctx, s.cfg.WSURL, s.api, s.metrics)
			if err != nil {
				s.metrics.RecordError()
				select {
				case errCh <- fmt.Errorf("user %d ws: %w", user.Index, err):
				default:
				}
				return
			}

			s.connMu.Lock()
			s.conns[user.Index] = conn
			s.connMu.Unlock()
		}(u, delay)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
	}

	// Allow presence snapshots to settle.
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (s *Simulator) runTraffic(ctx context.Context) {
	senders := make([]*User, 0, s.cfg.Conversations*2)
	for _, u := range s.users {
		if u.ConvID != "" {
			senders = append(senders, u)
		}
	}
	if len(senders) == 0 {
		return
	}

	fmt.Printf("traffic: %d active senders, %.0f msg/s\n", len(senders), s.cfg.MessagesPerSecond)

	limiter := rate.NewLimiter(rate.Limit(s.cfg.MessagesPerSecond), max(1, int(s.cfg.MessagesPerSecond)))

	for {
		if err := limiter.Wait(ctx); err != nil {
			return
		}

		u := senders[rand.Intn(len(senders))]
		s.connMu.Lock()
		conn := s.conns[u.Index]
		s.connMu.Unlock()
		if conn == nil {
			s.metrics.RecordError()
			continue
		}
		if err := u.SendMessage(conn, s.metrics); err != nil {
			continue
		}
	}
}

func (s *Simulator) closeAll() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	for _, conn := range s.conns {
		if conn != nil {
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			conn.Close()
		}
	}
}

func (s *Simulator) reportProgress(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.metrics.Snapshot().Print("progress")
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
