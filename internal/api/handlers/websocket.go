package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    int64       `json:"userId,omitempty"`
}

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[string]*WebSocketClient
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	broadcast  chan WebSocketMessage
	mutex      sync.RWMutex
	logger     *logger.Logger
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID        string
	UserID    int64
	MessageCh chan []byte
	Done      chan struct{}
	Logger    *logger.Logger
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub *WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(log *logger.Logger) *WebSocketHandler {
	hub := &WebSocketHub{
		clients:    make(map[string]*WebSocketClient),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		broadcast:  make(chan WebSocketMessage),
		logger:     log,
	}

	// Start the hub
	go hub.run()

	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleConnection handles new WebSocket connections
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// For now, implement a simple SSE (Server-Sent Events) approach
	// In a full implementation, you'd use gorilla/websocket or nhooyr/websocket

	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client
	client := &WebSocketClient{
		ID:        generateClientID(),
		UserID:    userID,
		MessageCh: make(chan []byte, 256),
		Done:      make(chan struct{}),
		Logger:    h.hub.logger,
	}

	// Register client
	h.hub.register <- client
	defer func() {
		h.hub.unregister <- client
	}()

	// Keep connection alive
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial connection message
	initMsg := WebSocketMessage{
		Type:      "connected",
		Data:      map[string]interface{}{"clientID": client.ID},
		Timestamp: time.Now(),
		UserID:    userID,
	}
	if data, err := json.Marshal(initMsg); err == nil {
		w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
	}

	// Listen for messages
	for {
		select {
		case message := <-client.MessageCh:
			w.Write([]byte("data: " + string(message) + "\n\n"))
			flusher.Flush()
		case <-client.Done:
			return
		case <-r.Context().Done():
			return
		}
	}
}

// run starts the WebSocket hub
func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.ID] = client
			h.mutex.Unlock()
			h.logger.WithFields(map[string]interface{}{
				"client_id": client.ID,
				"user_id":   client.UserID,
			}).Info("WebSocket client connected")

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.MessageCh)
				close(client.Done)
				h.logger.WithFields(map[string]interface{}{
					"client_id": client.ID,
					"user_id":   client.UserID,
				}).Info("WebSocket client disconnected")
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for _, client := range h.clients {
				// Send only to matching user or broadcast to all
				if message.UserID == 0 || message.UserID == client.UserID {
					if data, err := json.Marshal(message); err == nil {
						select {
						case client.MessageCh <- data:
						default:
							// Client channel is full, skip
						}
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastDriftDetected sends drift detection notification to specific user
func (h *WebSocketHub) BroadcastDriftDetected(userID int64, drift interface{}) {
	message := WebSocketMessage{
		Type:      "drift_detected",
		Data:      drift,
		Timestamp: time.Now(),
		UserID:    userID,
	}
	h.broadcast <- message
}

// BroadcastDriftResolved sends drift resolution notification
func (h *WebSocketHub) BroadcastDriftResolved(userID int64, driftID int64) {
	message := WebSocketMessage{
		Type:      "drift_resolved",
		Data:      map[string]interface{}{"id": driftID},
		Timestamp: time.Now(),
		UserID:    userID,
	}
	h.broadcast <- message
}

// BroadcastScanProgress sends scan progress updates
func (h *WebSocketHub) BroadcastScanProgress(userID int64, progress int, message string) {
	msg := WebSocketMessage{
		Type: "scan_progress",
		Data: map[string]interface{}{
			"progress": progress,
			"message":  message,
		},
		Timestamp: time.Now(),
		UserID:    userID,
	}
	h.broadcast <- msg
}

// BroadcastToAll sends a message to all connected clients
func (h *WebSocketHub) BroadcastToAll(messageType string, data interface{}) {
	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now(),
		UserID:    0, // 0 means broadcast to all
	}
	h.broadcast <- message
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
