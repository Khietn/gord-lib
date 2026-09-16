package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Khietn/gord-lib/discord"
	"github.com/Khietn/gord-lib/events"
)

// Gateway manages the WebSocket connection and protocol lifecycle with Discord.
type Gateway struct {
	mu         sync.RWMutex
	token      string
	intents    discord.GatewayIntents
	state      State
	conn       Conn
	dialer     Dialer
	dispatcher *events.Dispatcher

	heartbeat *HeartbeatManager
	sequence  atomic.Int64
	sessionID string
	resumeURL string

	ctx    context.Context
	cancel context.CancelFunc
}

// Config configures a Gateway instance.
type Config struct {
	Token      string
	Intents    discord.GatewayIntents
	Dialer     Dialer
	Dispatcher *events.Dispatcher
}

// New creates a new Gateway instance in StateDisconnected.
func New(cfg Config) *Gateway {
	ctx, cancel := context.WithCancel(context.Background())
	g := &Gateway{
		token:      cfg.Token,
		intents:    cfg.Intents,
		dialer:     cfg.Dialer,
		dispatcher: cfg.Dispatcher,
		state:      StateDisconnected,
		ctx:        ctx,
		cancel:     cancel,
	}
	g.sequence.Store(0)
	return g
}

// State returns the current connection state.
func (g *Gateway) State() State {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}

// setState transitions the FSM state.
func (g *Gateway) setState(s State) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.state = s
}

// Open connects to the Discord gateway WebSocket.
func (g *Gateway) Open(ctx context.Context, endpointURL string) error {
	g.setState(StateConnecting)

	if g.dialer == nil {
		return fmt.Errorf("no websocket dialer configured")
	}

	conn, err := g.dialer.Dial(ctx, endpointURL)
	if err != nil {
		g.setState(StateDisconnected)
		return fmt.Errorf("gateway dial failed: %w", err)
	}

	g.mu.Lock()
	g.conn = conn
	g.mu.Unlock()

	g.setState(StateHandshaking)

	// Start read loop
	go g.readLoop()
	return nil
}

// WriteJSON encodes and writes a JSON payload to the WebSocket.
func (g *Gateway) WriteJSON(ctx context.Context, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal gateway payload: %w", err)
	}

	g.mu.RLock()
	conn := g.conn
	g.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("cannot write: gateway connection is nil")
	}

	return conn.Write(ctx, TextMessage, data)
}

// readLoop reads frames from the WebSocket and routes them.
func (g *Gateway) readLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		default:
		}

		g.mu.RLock()
		conn := g.conn
		g.mu.RUnlock()

		if conn == nil {
			return
		}

		_, data, err := conn.Read(g.ctx)
		if err != nil {
			g.handleDisconnect()
			return
		}

		var p Payload
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}

		if p.Sequence != nil {
			g.sequence.Store(*p.Sequence)
		}

		g.handlePayload(&p)
	}
}

// handlePayload processes incoming Gateway OpCodes.
func (g *Gateway) handlePayload(p *Payload) {
	switch p.Op {
	case OpcodeHello:
		var hello HelloData
		if err := json.Unmarshal(p.Data, &hello); err == nil {
			g.startHeartbeat(hello.HeartbeatInterval)
		}

		// Identify or Resume
		if g.sessionID != "" && g.sequence.Load() > 0 {
			_ = g.resume()
		} else {
			_ = g.identify()
		}

	case OpcodeHeartbeatACK:
		if g.heartbeat != nil {
			g.heartbeat.Acknowledge()
		}

	case OpcodeHeartbeat:
		seq := g.sequence.Load()
		_ = g.sendHeartbeat(&seq)

	case OpcodeReconnect:
		g.handleDisconnect()

	case OpcodeInvalidSession:
		var resumable bool
		_ = json.Unmarshal(p.Data, &resumable)
		if resumable {
			_ = g.resume()
		} else {
			g.sessionID = ""
			g.sequence.Store(0)
			_ = g.identify()
		}

	case OpcodeDispatch:
		g.handleDispatch(p)
	}
}

// startHeartbeat starts the heartbeat loop.
func (g *Gateway) startHeartbeat(intervalMs int) {
	if g.heartbeat != nil {
		g.heartbeat.Stop()
	}

	g.heartbeat = NewHeartbeatManager(
		intervalMs,
		g.sendHeartbeat,
		func() {
			// Zombie connection detected
			g.handleDisconnect()
		},
	)

	g.heartbeat.Start(g.ctx, func() *int64 {
		s := g.sequence.Load()
		if s == 0 {
			return nil
		}
		return &s
	})
}

// sendHeartbeat transmits a heartbeat to the server.
func (g *Gateway) sendHeartbeat(seq *int64) error {
	p := Payload{
		Op: OpcodeHeartbeat,
	}
	if seq != nil {
		raw, _ := json.Marshal(*seq)
		p.Data = raw
	}
	return g.WriteJSON(g.ctx, p)
}

// identify sends the OpcodeIdentify payload.
func (g *Gateway) identify() error {
	g.setState(StateIdentifying)

	payload := Payload{
		Op: OpcodeIdentify,
	}
	identifyData := IdentifyData{
		Token:   g.token,
		Intents: g.intents,
		Properties: IdentifyProperties{
			OS:      runtime.GOOS,
			Browser: "gord-lib",
			Device:  "gord-lib",
		},
	}
	raw, _ := json.Marshal(identifyData)
	payload.Data = raw

	return g.WriteJSON(g.ctx, payload)
}

// resume sends the OpcodeResume payload.
func (g *Gateway) resume() error {
	g.setState(StateResuming)

	payload := Payload{
		Op: OpcodeResume,
	}
	resumeData := ResumeData{
		Token:     g.token,
		SessionID: g.sessionID,
		Sequence:  g.sequence.Load(),
	}
	raw, _ := json.Marshal(resumeData)
	payload.Data = raw

	return g.WriteJSON(g.ctx, payload)
}

// handleDispatch unpacks events and sends them to the worker pool.
func (g *Gateway) handleDispatch(p *Payload) {
	if g.dispatcher == nil {
		return
	}

	switch p.EventName {
	case "READY":
		var ready ReadyData
		if err := json.Unmarshal(p.Data, &ready); err == nil {
			g.sessionID = ready.SessionID
			g.resumeURL = ready.ResumeURL
			g.setState(StateReady)

			g.dispatcher.Dispatch(events.Ready{
				User:             ready.User,
				SessionID:        ready.SessionID,
				ResumeGatewayURL: ready.ResumeURL,
				Shard:            ready.Shard,
				ApplicationID:    ready.Application.ID,
			})
		}

	case "MESSAGE_CREATE":
		var msg discord.Message
		if err := json.Unmarshal(p.Data, &msg); err == nil {
			g.dispatcher.Dispatch(events.MessageCreate{Message: msg})
		}

	case "INTERACTION_CREATE":
		var interaction discord.Interaction
		if err := json.Unmarshal(p.Data, &interaction); err == nil {
			g.dispatcher.Dispatch(events.InteractionCreate{Interaction: interaction})
		}

	case "GUILD_CREATE":
		var guild discord.Guild
		if err := json.Unmarshal(p.Data, &guild); err == nil {
			g.dispatcher.Dispatch(events.GuildCreate{Guild: guild})
		}
	}
}

// handleDisconnect coordinates reconnection upon disconnect.
func (g *Gateway) handleDisconnect() {
	g.setState(StateReconnecting)
	if g.heartbeat != nil {
		g.heartbeat.Stop()
	}
	g.mu.Lock()
	if g.conn != nil {
		_ = g.conn.Close(1000, "reconnecting")
		g.conn = nil
	}
	g.mu.Unlock()
}

// Close gracefully closes the Gateway connection.
func (g *Gateway) Close() error {
	g.setState(StateClosed)
	g.cancel()

	if g.heartbeat != nil {
		g.heartbeat.Stop()
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn != nil {
		err := g.conn.Close(1000, "closing")
		g.conn = nil
		return err
	}
	return nil
}
