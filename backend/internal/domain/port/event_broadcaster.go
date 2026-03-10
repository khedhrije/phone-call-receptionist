package port

// EventBroadcaster defines the interface for broadcasting real-time events.
// Implementations push events to connected clients (e.g., via WebSocket).
type EventBroadcaster interface {
	// Broadcast sends an event to all connected clients.
	Broadcast(event interface{})
}
