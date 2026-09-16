package gateway

// State represents the lifecycle state of the Gateway connection FSM.
type State int

const (
	StateDisconnected State = iota
	StateConnecting
	StateHandshaking
	StateIdentifying
	StateResuming
	StateReady
	StateReconnecting
	StateClosed
)

// String returns the string representation of the Gateway State.
func (s State) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateHandshaking:
		return "Handshaking"
	case StateIdentifying:
		return "Identifying"
	case StateResuming:
		return "Resuming"
	case StateReady:
		return "Ready"
	case StateReconnecting:
		return "Reconnecting"
	case StateClosed:
		return "Closed"
	default:
		return "Unknown"
	}
}
