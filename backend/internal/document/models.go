package document

import "time"

// AccessLevel defines coarse permission tiers for a document.
type AccessLevel string

const (
	AccessView    AccessLevel = "view"
	AccessComment AccessLevel = "comment"
	AccessEdit    AccessLevel = "edit"
)

// Permission models an explicit subject -> access mapping.
type Permission struct {
	SubjectID   string      `json:"subjectId"`
	SubjectType string      `json:"subjectType"` // user | group | link
	Level       AccessLevel `json:"level"`
}

// ShareLink represents an expirable link with a scoped permission.
type ShareLink struct {
	ID         string      `json:"id"`
	Token      string      `json:"token"`
	Level      AccessLevel `json:"level"`
	ExpiresAt  *time.Time  `json:"expiresAt,omitempty"`
	CreatedAt  time.Time   `json:"createdAt"`
	CreatedBy  string      `json:"createdBy"`
	DocumentID string      `json:"documentId"`
	TenantID   string      `json:"tenantId"`
}

// Operation captures a CRDT/OT operation envelope for auditing and replay.
type Operation struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	TenantID   string    `json:"tenantId"`
	UserID     string    `json:"userId"`
	Lamport    int64     `json:"lamport"`
	Delta      string    `json:"delta"` // transport-safe serialized operation payload
	CreatedAt  time.Time `json:"createdAt"`
}

// DocumentVersion stores a fully materialized version snapshot.
type DocumentVersion struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	TenantID   string    `json:"tenantId"`
	AuthorID   string    `json:"authorId"`
	Sequence   int64     `json:"sequence"`
	Content    string    `json:"content"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"createdAt"`
}

// User represents a registered user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Document is the aggregate root for collaboration.
type Document struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	OwnerID     string                 `json:"ownerId"`
	Permissions map[string]AccessLevel `json:"permissions"`
	ShareLinks  []ShareLink            `json:"shareLinks"`
	Version     int64                  `json:"version"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}
