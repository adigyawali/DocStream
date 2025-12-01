package document

import (
	"context"
)

// Repository abstracts persistence for documents, operations, and versions.
type Repository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByEmail(ctx context.Context, email string) (User, error)

	CreateDocument(ctx context.Context, doc Document) (Document, error)
	GetDocument(ctx context.Context, tenantID, documentID string) (Document, error)
	UpdateDocument(ctx context.Context, doc Document) error
	ListDocuments(ctx context.Context, tenantID string) ([]Document, error)

	SaveOperation(ctx context.Context, op Operation) error

	SaveVersion(ctx context.Context, version DocumentVersion) error
	ListVersions(ctx context.Context, tenantID, documentID string, limit int) ([]DocumentVersion, error)
}
