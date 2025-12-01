package document

import (
	"context"
	"sync"
)

// InMemoryRepository is a thread-safe store for development and tests.
type InMemoryRepository struct {
	mu         sync.RWMutex
	documents  map[string]map[string]Document
	versions   map[string][]DocumentVersion
	operations map[string][]Operation
	users      map[string]User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		documents:  make(map[string]map[string]Document),
		versions:   make(map[string][]DocumentVersion),
		operations: make(map[string][]Operation),
		users:      make(map[string]User),
	}
}

func (r *InMemoryRepository) CreateUser(_ context.Context, user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.Email] = user
	return nil
}

func (r *InMemoryRepository) GetUserByEmail(_ context.Context, email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[email]
	if !ok {
		return User{}, ErrDocumentNotFound // Reusing error or need ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryRepository) CreateDocument(_ context.Context, doc Document) (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.documents[doc.TenantID]; !ok {
		r.documents[doc.TenantID] = make(map[string]Document)
	}
	r.documents[doc.TenantID][doc.ID] = doc
	return doc, nil
}

func (r *InMemoryRepository) GetDocument(_ context.Context, tenantID, documentID string) (Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantDocs, ok := r.documents[tenantID]
	if !ok {
		return Document{}, ErrDocumentNotFound
	}
	doc, ok := tenantDocs[documentID]
	if !ok {
		return Document{}, ErrDocumentNotFound
	}
	return doc, nil
}

func (r *InMemoryRepository) UpdateDocument(_ context.Context, doc Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.documents[doc.TenantID]; !ok {
		return ErrDocumentNotFound
	}
	if _, ok := r.documents[doc.TenantID][doc.ID]; !ok {
		return ErrDocumentNotFound
	}
	r.documents[doc.TenantID][doc.ID] = doc
	return nil
}

func (r *InMemoryRepository) ListDocuments(_ context.Context, tenantID string) ([]Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenantDocs, ok := r.documents[tenantID]
	if !ok {
		return []Document{}, nil
	}

	out := make([]Document, 0, len(tenantDocs))
	for _, doc := range tenantDocs {
		out = append(out, doc)
	}
	return out, nil
}

func (r *InMemoryRepository) SaveOperation(_ context.Context, op Operation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.operations[op.DocumentID] = append(r.operations[op.DocumentID], op)
	return nil
}

func (r *InMemoryRepository) SaveVersion(_ context.Context, version DocumentVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.versions[version.DocumentID] = append(r.versions[version.DocumentID], version)
	return nil
}

func (r *InMemoryRepository) ListVersions(_ context.Context, tenantID, documentID string, limit int) ([]DocumentVersion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Tenant ID is included for future persistence implementations.
	_ = tenantID

	versions := r.versions[documentID]
	if limit > 0 && len(versions) > limit {
		return versions[len(versions)-limit:], nil
	}
	return append([]DocumentVersion(nil), versions...), nil
}
