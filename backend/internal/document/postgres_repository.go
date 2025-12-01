package document

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) EnsureSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			owner_id TEXT NOT NULL, -- No FK to allow loose coupling if needed, or add REFERENCES users(id)
			version BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS permissions (
			document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			subject_id TEXT NOT NULL,
			subject_type TEXT NOT NULL,
			level TEXT NOT NULL,
			PRIMARY KEY (document_id, subject_id)
		);`,
		`CREATE TABLE IF NOT EXISTS share_links (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			tenant_id TEXT NOT NULL,
			token TEXT NOT NULL,
			level TEXT NOT NULL,
			expires_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			created_by TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS document_versions (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			tenant_id TEXT NOT NULL,
			author_id TEXT NOT NULL,
			sequence BIGINT NOT NULL,
			content TEXT NOT NULL,
			label TEXT,
			created_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS operations (
			id TEXT PRIMARY KEY,
			document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
			tenant_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			lamport BIGINT NOT NULL,
			delta TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("failed to exec schema query: %w", err)
		}
	}
	return nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
	return err
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (r *PostgresRepository) CreateDocument(ctx context.Context, doc Document) (Document, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO documents (id, tenant_id, title, content, owner_id, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, doc.ID, doc.TenantID, doc.Title, doc.Content, doc.OwnerID, doc.Version, doc.CreatedAt, doc.UpdatedAt)
	if err != nil {
		return Document{}, err
	}

	// Save Permissions
	for sub, lvl := range doc.Permissions {
		_, err = tx.Exec(ctx, `INSERT INTO permissions (document_id, subject_id, subject_type, level) VALUES ($1, $2, $3, $4)`,
			doc.ID, sub, "user", lvl) // Defaulting to 'user' for now as per map structure
		if err != nil {
			return Document{}, err
		}
	}

	// Save ShareLinks
	for _, link := range doc.ShareLinks {
		_, err = tx.Exec(ctx, `INSERT INTO share_links (id, document_id, tenant_id, token, level, expires_at, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			link.ID, doc.ID, link.TenantID, link.Token, link.Level, link.ExpiresAt, link.CreatedAt, link.CreatedBy)
		if err != nil {
			return Document{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Document{}, err
	}

	return doc, nil
}

func (r *PostgresRepository) GetDocument(ctx context.Context, tenantID, documentID string) (Document, error) {
	var doc Document
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, content, owner_id, version, created_at, updated_at
		FROM documents WHERE tenant_id = $1 AND id = $2
	`, tenantID, documentID).Scan(&doc.ID, &doc.TenantID, &doc.Title, &doc.Content, &doc.OwnerID, &doc.Version, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		return Document{}, err // Could check for pgx.ErrNoRows and return ErrDocumentNotFound
	}

	// Load Permissions
	rows, err := r.db.Query(ctx, `SELECT subject_id, level FROM permissions WHERE document_id = $1`, documentID)
	if err != nil {
		return Document{}, err
	}
	doc.Permissions = make(map[string]AccessLevel)
	for rows.Next() {
		var sub string
		var lvl AccessLevel
		if err := rows.Scan(&sub, &lvl); err != nil {
			rows.Close()
			return Document{}, err
		}
		doc.Permissions[sub] = lvl
	}
	rows.Close()

	// Load ShareLinks
	rows, err = r.db.Query(ctx, `SELECT id, token, level, expires_at, created_at, created_by, tenant_id FROM share_links WHERE document_id = $1`, documentID)
	if err != nil {
		return Document{}, err
	}
	doc.ShareLinks = []ShareLink{}
	for rows.Next() {
		var sl ShareLink
		sl.DocumentID = documentID
		if err := rows.Scan(&sl.ID, &sl.Token, &sl.Level, &sl.ExpiresAt, &sl.CreatedAt, &sl.CreatedBy, &sl.TenantID); err != nil {
			rows.Close()
			return Document{}, err
		}
		doc.ShareLinks = append(doc.ShareLinks, sl)
	}
	rows.Close()

	return doc, nil
}

func (r *PostgresRepository) UpdateDocument(ctx context.Context, doc Document) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ct, err := tx.Exec(ctx, `
		UPDATE documents SET title=$1, content=$2, version=$3, updated_at=$4
		WHERE id=$5 AND tenant_id=$6
	`, doc.Title, doc.Content, doc.Version, doc.UpdatedAt, doc.ID, doc.TenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("document not found or no change")
	}

	// Replace Permissions
	_, err = tx.Exec(ctx, `DELETE FROM permissions WHERE document_id = $1`, doc.ID)
	if err != nil {
		return err
	}
	for sub, lvl := range doc.Permissions {
		_, err = tx.Exec(ctx, `INSERT INTO permissions (document_id, subject_id, subject_type, level) VALUES ($1, $2, $3, $4)`,
			doc.ID, sub, "user", lvl)
		if err != nil {
			return err
		}
	}

	// Replace ShareLinks (simpler than diffing)
	_, err = tx.Exec(ctx, `DELETE FROM share_links WHERE document_id = $1`, doc.ID)
	if err != nil {
		return err
	}
	for _, link := range doc.ShareLinks {
		_, err = tx.Exec(ctx, `INSERT INTO share_links (id, document_id, tenant_id, token, level, expires_at, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			link.ID, doc.ID, link.TenantID, link.Token, link.Level, link.ExpiresAt, link.CreatedAt, link.CreatedBy)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) ListDocuments(ctx context.Context, tenantID string) ([]Document, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tenant_id, title, content, owner_id, version, created_at, updated_at
		FROM documents WHERE tenant_id = $1 ORDER BY updated_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := []Document{}
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.TenantID, &doc.Title, &doc.Content, &doc.OwnerID, &doc.Version, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		// Optimization: Don't load permissions/links for list view if not needed,
		// or do N+1 (bad) or JOIN (complex). For now, returning basic info is usually enough for list.
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *PostgresRepository) SaveOperation(ctx context.Context, op Operation) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO operations (id, document_id, tenant_id, user_id, lamport, delta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, op.ID, op.DocumentID, op.TenantID, op.UserID, op.Lamport, op.Delta, op.CreatedAt)
	return err
}

func (r *PostgresRepository) SaveVersion(ctx context.Context, version DocumentVersion) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO document_versions (id, document_id, tenant_id, author_id, sequence, content, label, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, version.ID, version.DocumentID, version.TenantID, version.AuthorID, version.Sequence, version.Content, version.Label, version.CreatedAt)
	return err
}

func (r *PostgresRepository) ListVersions(ctx context.Context, tenantID, documentID string, limit int) ([]DocumentVersion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, document_id, tenant_id, author_id, sequence, content, label, created_at
		FROM document_versions WHERE tenant_id = $1 AND document_id = $2
		ORDER BY sequence DESC LIMIT $3
	`, tenantID, documentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := []DocumentVersion{}
	for rows.Next() {
		var v DocumentVersion
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.TenantID, &v.AuthorID, &v.Sequence, &v.Content, &v.Label, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}
