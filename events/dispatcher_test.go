package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Khietn/gord-lib/discord"
)

func TestDispatcher_TypedHandler(t *testing.T) {
	d := NewDispatcher(WithWorkerCount(4), WithQueueBuffer(64))
	defer d.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMsg string
	Register(d, func(ctx context.Context, e MessageCreate) {
		receivedMsg = e.Content
		wg.Done()
	})

	d.Dispatch(MessageCreate{
		Message: discord.Message{
			Content: "Hello from Worker Pool",
		},
	})

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		if receivedMsg != "Hello from Worker Pool" {
			t.Fatalf("expected message 'Hello from Worker Pool', got %s", receivedMsg)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for event dispatch")
	}
}

func TestDispatcher_Middleware(t *testing.T) {
	d := NewDispatcher(WithWorkerCount(2))
	defer d.Close()

	var middlewareRan bool
	var mu sync.Mutex

	d.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, e Event) {
			mu.Lock()
			middlewareRan = true
			mu.Unlock()
			next(ctx, e)
		}
	})

	var wg sync.WaitGroup
	wg.Add(1)

	Register(d, func(ctx context.Context, e MessageCreate) {
		wg.Done()
	})

	d.Dispatch(MessageCreate{})
	wg.Wait()

	mu.Lock()
	ran := middlewareRan
	mu.Unlock()

	if !ran {
		t.Fatalf("expected custom middleware to execute")
	}
}

func TestDispatcher_PanicRecovery(t *testing.T) {
	d := NewDispatcher(WithWorkerCount(1))
	defer d.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	firstRan := false
	secondRan := false

	// First handler panics
	Register(d, func(ctx context.Context, e MessageCreate) {
		firstRan = true
		wg.Done()
		panic("deliberate test panic in handler")
	})

	// Dispatch first event
	d.Dispatch(MessageCreate{})

	// Register second handler to verify worker pool didn't die
	Register(d, func(ctx context.Context, e Ready) {
		secondRan = true
		wg.Done()
	})

	// Dispatch second event
	d.Dispatch(Ready{})

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		if !firstRan || !secondRan {
			t.Errorf("expected both handlers to run despite panic")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out: worker pool died or blocked on panic")
	}
}
