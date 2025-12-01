package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"docStream/backend/internal/document"
	"github.com/gorilla/websocket"
)

type inboundEvent struct {
	client  *Client
	message ClientMessage
}

// Room represents collaborators on a single document.
type Room struct {
	tenantID   string
	documentID string
	service    *document.Service

	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
	inbound    chan inboundEvent
}

func newRoom(service *document.Service, tenantID, documentID string) *Room {
	return &Room{
		tenantID:   tenantID,
		documentID: documentID,
		service:    service,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		inbound:    make(chan inboundEvent, 64),
	}
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case evt := <-r.inbound:
			r.handleEvent(evt)
		}
	}
}

func (r *Room) handleEvent(evt inboundEvent) {
	switch evt.message.Type {
	case "operation":
		r.handleOperation(evt)
	case "presence":
		r.broadcast(ServerMessage{
			Type:       "presence",
			TenantID:   r.tenantID,
			DocumentID: r.documentID,
			UserID:     evt.message.UserID,
			Message:    "presence",
		})
	default:
		r.broadcast(ServerMessage{
			Type:       "ack",
			TenantID:   r.tenantID,
			DocumentID: r.documentID,
			UserID:     evt.message.UserID,
			Message:    "noop",
		})
	}
}

func (r *Room) handleOperation(evt inboundEvent) {
	doc, op, version, err := r.service.ApplyOperation(evt.client.ctx, document.ApplyOperationInput{
		TenantID:   r.tenantID,
		DocumentID: r.documentID,
		UserID:     evt.message.UserID,
		Delta:      evt.message.Delta,
		NewContent: evt.message.NewContent,
		Lamport:    evt.message.Lamport,
		Label:      evt.message.Label,
	})
	if err != nil {
		log.Printf("apply operation failed: %v", err)
		evt.client.send <- marshal(ServerMessage{
			Type:       "error",
			TenantID:   r.tenantID,
			DocumentID: r.documentID,
			UserID:     evt.message.UserID,
			Message:    err.Error(),
		})
		return
	}

	r.broadcast(ServerMessage{
		Type:       "update",
		TenantID:   r.tenantID,
		DocumentID: r.documentID,
		UserID:     evt.message.UserID,
		Version:    doc.Version,
		Content:    doc.Content,
		Operation:  &op,
		Versioned:  &version,
	})
}

func (r *Room) broadcast(msg ServerMessage) {
	payload := marshal(msg)
	for client := range r.clients {
		select {
		case client.send <- payload:
		default:
			close(client.send)
			delete(r.clients, client)
		}
	}
}

// Hub keeps track of rooms per tenant/document.
type Hub struct {
	service *document.Service
	rooms   map[string]*Room
	mu      sync.Mutex
}

func NewHub(service *document.Service) *Hub {
	return &Hub{
		service: service,
		rooms:   make(map[string]*Room),
	}
}

func (h *Hub) roomKey(tenantID, documentID string) string {
	return tenantID + ":" + documentID
}

func (h *Hub) getRoom(tenantID, documentID string) *Room {
	key := h.roomKey(tenantID, documentID)
	h.mu.Lock()
	defer h.mu.Unlock()
	room, ok := h.rooms[key]
	if !ok {
		room = newRoom(h.service, tenantID, documentID)
		h.rooms[key] = room
		go room.run()
	}
	return room
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// ServeWS upgrades and attaches a client to a room.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenantId")
	docID := r.URL.Query().Get("docId")
	userID := r.URL.Query().Get("userId")
	if tenantID == "" || docID == "" || userID == "" {
		http.Error(w, "tenantId, docId, and userId are required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	room := h.getRoom(tenantID, docID)
	client := &Client{
		room:   room,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		ctx:    r.Context(),
	}
	room.register <- client

	go client.writePump()
	go client.readPump()

	// Send initial document snapshot to the new client.
	doc, err := h.service.GetDocument(r.Context(), tenantID, docID)
	if err == nil {
		client.send <- marshal(ServerMessage{
			Type:       "snapshot",
			TenantID:   tenantID,
			DocumentID: docID,
			UserID:     userID,
			Version:    doc.Version,
			Content:    doc.Content,
			Message:    "initial",
		})
	}
}

// marshal keeps websocket writes lightweight.
func marshal(msg ServerMessage) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return []byte("{}")
	}
	return b
}
