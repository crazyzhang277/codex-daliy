package proxy

import (
	"sync"
	"time"

	"nvidialimiter/internal/config"
)

type State struct {
	cfg        config.Config
	mu         sync.Mutex
	cond       *sync.Cond
	lastSentAt time.Time
	total      int
	nextTicket uint64
	serving    uint64
}

func NewState(cfg config.Config) *State {
	s := &State{cfg: cfg}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *State) ShouldLimit(model string) bool {
	return s.cfg.Match(model)
}

func (s *State) ReserveAndWait() (time.Duration, bool) {
	if !s.cfg.Enabled {
		return 0, true
	}

	s.mu.Lock()
	if s.total >= 10 {
		wait := s.remainingLocked()
		s.mu.Unlock()
		return wait, false
	}

	myTicket := s.nextTicket
	s.nextTicket++
	s.total++

	for {
		if myTicket != s.serving {
			s.cond.Wait()
			continue
		}

		wait := s.remainingLocked()
		if wait > 0 {
			s.mu.Unlock()
			time.Sleep(wait)
			s.mu.Lock()
			continue
		}

		s.lastSentAt = time.Now()
		s.serving++
		s.total--
		s.cond.Broadcast()
		s.mu.Unlock()
		return 0, true
	}
}

func (s *State) remainingLocked() time.Duration {
	const cooldown = 2 * time.Second
	if s.lastSentAt.IsZero() {
		return 0
	}
	remaining := cooldown - time.Since(s.lastSentAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}
