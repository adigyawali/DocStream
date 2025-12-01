package realtime

import "docStream/backend/internal/document"

// ClientMessage is the envelope received from the websocket clients.
type ClientMessage struct {
	Type       string `json:"type"` // join | operation | presence | ping
	TenantID   string `json:"tenantId"`
	DocumentID string `json:"documentId"`
	UserID     string `json:"userId"`
	Delta      string `json:"delta,omitempty"`
	NewContent string `json:"newContent,omitempty"`
	Lamport    int64  `json:"lamport,omitempty"`
	Label      string `json:"label,omitempty"`
}

// ServerMessage is broadcast to connected collaborators.
type ServerMessage struct {
	Type       string                    `json:"type"` // update | ack | presence
	TenantID   string                    `json:"tenantId"`
	DocumentID string                    `json:"documentId"`
	UserID     string                    `json:"userId"`
	Version    int64                     `json:"version"`
	Content    string                    `json:"content,omitempty"`
	Operation  *document.Operation       `json:"operation,omitempty"`
	Versioned  *document.DocumentVersion `json:"versioned,omitempty"`
	Message    string                    `json:"message,omitempty"`
}
