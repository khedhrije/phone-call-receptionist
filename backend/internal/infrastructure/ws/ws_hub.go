// Package ws implements the EventBroadcaster port using WebSocket connections.
package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

)

const (
	// writeWait is the time allowed to write a message to the client.
	writeWait = 10 * time.Second
	// pongWait is the time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// pingPeriod is the interval between ping messages sent to the client.
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize is the maximum allowed message size from a client.
	maxMessageSize = 512
)

// upgrader is the WebSocket connection upgrader with permissive origin check.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a single WebSocket client connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub manages WebSocket client connections and broadcasts events to all clients.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	logger     *zerolog.Logger
}

// NewHub creates a new WebSocket hub and starts its run loop.
func NewHub(logger *zerolog.Logger) *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
	go h.run()
	return h
}

// run processes client registration, unregistration, and broadcast messages.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debug().Int("clients", len(h.clients)).Msg("[WSHub] client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debug().Int("clients", len(h.clients)).Msg("[WSHub] client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends an event to all connected WebSocket clients.
// The event is JSON-encoded before broadcasting.
func (h *Hub) Broadcast(_ context.Context, event interface{}) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error().Err(err).Msg("[WSHub] failed to marshal broadcast event")
		return
	}

	h.logger.Debug().Int("bytes", len(data)).Msg("[WSHub] broadcasting event")
	h.broadcast <- data
}

// HandleWebSocket upgrades an HTTP connection to a WebSocket connection
// and registers the client with the hub.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("[WSHub] failed to upgrade websocket connection")
		return
	}

	h.logger.Debug().Str("remoteAddr", r.RemoteAddr).Msg("[WSHub] websocket connection upgraded")

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection.
// It handles pong messages and detects client disconnection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Warn().Err(err).Msg("[WSHub] unexpected close")
			}
			break
		}
	}
}

// writePump writes messages and ping frames to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
