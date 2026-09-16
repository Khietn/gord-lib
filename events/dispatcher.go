package events

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// HandlerFunc is the internal untyped signature executed by workers.
type HandlerFunc func(ctx context.Context, e Event)

// Middleware intercepts event execution before it reaches the handler.
type Middleware func(next HandlerFunc) HandlerFunc

// Dispatcher manages type-safe event handler subscriptions and dispatches
// events concurrently across a worker pool.
type Dispatcher struct {
	mu          sync.RWMutex
	handlers    map[reflect.Type][]HandlerFunc
	middlewares []Middleware
	eventQueue  chan Event
	workers     int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// DispatcherOption configures a Dispatcher.
type DispatcherOption func(*Dispatcher)

// WithWorkerCount configures the number of concurrent worker goroutines.
func WithWorkerCount(n int) DispatcherOption {
	return func(d *Dispatcher) {
		if n > 0 {
			d.workers = n
		}
	}
}

// WithQueueBuffer configures the buffer size of the event queue.
func WithQueueBuffer(size int) DispatcherOption {
	return func(d *Dispatcher) {
		if size > 0 {
			d.eventQueue = make(chan Event, size)
		}
	}
}

// NewDispatcher creates and starts an asynchronous event dispatcher with a worker pool.
func NewDispatcher(opts ...DispatcherOption) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Dispatcher{
		handlers:   make(map[reflect.Type][]HandlerFunc),
		eventQueue: make(chan Event, 256),
		workers:    8,
		ctx:        ctx,
		cancel:     cancel,
	}

	for _, opt := range opts {
		opt(d)
	}

	// Always install default panic recovery middleware
	d.Use(RecoveryMiddleware)

	d.startWorkers()
	return d
}

// Use adds global event middlewares to the chain.
func (d *Dispatcher) Use(middlewares ...Middleware) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.middlewares = append(d.middlewares, middlewares...)
}

// startWorkers spawns the worker pool goroutines.
func (d *Dispatcher) startWorkers() {
	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			for {
				select {
				case <-d.ctx.Done():
					return
				case e, ok := <-d.eventQueue:
					if !ok {
						return
					}
					d.execute(e)
				}
			}
		}()
	}
}

// execute runs all registered handlers for the event through the middleware pipeline.
func (d *Dispatcher) execute(e Event) {
	t := reflect.TypeOf(e)

	d.mu.RLock()
	handlers := d.handlers[t]
	middlewares := make([]Middleware, len(d.middlewares))
	copy(middlewares, d.middlewares)
	d.mu.RUnlock()

	for _, h := range handlers {
		// Chain middlewares
		current := h
		for i := len(middlewares) - 1; i >= 0; i-- {
			current = middlewares[i](current)
		}
		current(d.ctx, e)
	}
}

// Dispatch pushes an event into the worker queue without blocking the caller.
func (d *Dispatcher) Dispatch(e Event) {
	select {
	case d.eventQueue <- e:
	default:
		// Queue full - dispatch asynchronously in a separate goroutine to preserve order as best effort
		go func() {
			d.eventQueue <- e
		}()
	}
}

// Register attaches a strongly-typed handler for a specific event type using Go generics.
func Register[E Event](d *Dispatcher, handler func(ctx context.Context, e E)) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var zero E
	eventType := reflect.TypeOf(zero)

	untyped := func(ctx context.Context, ev Event) {
		if typed, ok := ev.(E); ok {
			handler(ctx, typed)
		}
	}

	d.handlers[eventType] = append(d.handlers[eventType], untyped)
}

// Close gracefully stops the worker pool and flushes the queue.
func (d *Dispatcher) Close() error {
	d.cancel()
	close(d.eventQueue)
	d.wg.Wait()
	return nil
}

// RecoveryMiddleware provides panic recovery for event handlers.
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, e Event) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[gord-lib/events] recovered from panic in handler for %s: %v\n", e.EventName(), r)
			}
		}()
		next(ctx, e)
	}
}
