package proxy

import (
	"sync"
	"testing"
	"time"

	"nvidialimiter/internal/config"
)

func TestReserveAndWaitFIFO(t *testing.T) {
	st := NewState(config.Config{Enabled: true, MatchMode: "mixed", Models: []string{"nvidia/"}})
	st.lastSentAt = time.Now().Add(-3 * time.Second)

	var mu sync.Mutex
	order := make([]int, 0, 2)

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if _, ok := st.ReserveAndWait(); !ok {
				t.Errorf("request %d was rejected", id)
				return
			}
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	if time.Since(start) < 2*time.Second {
		t.Fatalf("expected cooldown wait")
	}

	if len(order) != 2 {
		t.Fatalf("expected 2 completions, got %d", len(order))
	}

	if order[0] != 0 || order[1] != 1 {
		t.Fatalf("expected FIFO order, got %v", order)
	}
}
