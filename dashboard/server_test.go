package dashboard

import (
	"context"
	"testing"
	"time"
)

func TestGracefulShutdown(t *testing.T) {
	// port 0 lets the OS pick a free port; we only care that Shutdown returns.
	s := NewServer(0, "config.json", "logs", "test", nil)
	s.Start()
	time.Sleep(50 * time.Millisecond) // let ListenAndServe bind

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- s.Shutdown(ctx) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Shutdown returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown did not return within 3s")
	}

	// A second call must be safe (sync.Once guards the channel close).
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatalf("second Shutdown error: %v", err)
	}
}
