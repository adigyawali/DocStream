package document

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Service orchestrates document workflows (creation, permissions, versioning).
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type ApplyOperationInput struct {
	TenantID   string
	DocumentID string
	UserID     string
	Delta      string
	NewContent string
	Lamport    int64
	Label      string
}

func (s *Service) CreateDocument(ctx context.Context, tenantID, ownerID, title, initialContent string) (Document, error) {
	now := time.Now().UTC()
	doc := Document{
		ID:          NewID(),
		TenantID:    tenantID,
		Title:       title,
		Content:     initialContent,
		OwnerID:     ownerID,
		Permissions: map[string]AccessLevel{ownerID: AccessEdit},
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if _, err := s.repo.CreateDocument(ctx, doc); err != nil {
		return Document{}, fmt.Errorf("create document: %w", err)
	}

	version := DocumentVersion{
		ID:         NewID(),
		DocumentID: doc.ID,
		TenantID:   tenantID,
		AuthorID:   ownerID,
		Sequence:   doc.Version,
		Content:    doc.Content,
		Label:      "initial",
		CreatedAt:  now,
	}

	_ = s.repo.SaveVersion(ctx, version)
	return doc, nil
}

func (s *Service) GetDocument(ctx context.Context, tenantID, documentID string) (Document, error) {
	return s.repo.GetDocument(ctx, tenantID, documentID)
}

func (s *Service) ListDocuments(ctx context.Context, tenantID string) ([]Document, error) {
	return s.repo.ListDocuments(ctx, tenantID)
}

func (s *Service) ApplyOperation(ctx context.Context, in ApplyOperationInput) (Document, Operation, DocumentVersion, error) {
	doc, err := s.repo.GetDocument(ctx, in.TenantID, in.DocumentID)
	if err != nil {
		return Document{}, Operation{}, DocumentVersion{}, err
	}

	now := time.Now().UTC()
	doc.Content = in.NewContent
	doc.Version++
	doc.UpdatedAt = now

	op := Operation{
		ID:         NewID(),
		DocumentID: doc.ID,
		TenantID:   doc.TenantID,
		UserID:     in.UserID,
		Lamport:    in.Lamport,
		Delta:      in.Delta,
		CreatedAt:  now,
	}

	version := DocumentVersion{
		ID:         NewID(),
		DocumentID: doc.ID,
		TenantID:   doc.TenantID,
		AuthorID:   in.UserID,
		Sequence:   doc.Version,
		Content:    doc.Content,
		Label:      in.Label,
		CreatedAt:  now,
	}

	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return Document{}, Operation{}, DocumentVersion{}, fmt.Errorf("apply operation: %w", err)
	}
	_ = s.repo.SaveOperation(ctx, op)
	_ = s.repo.SaveVersion(ctx, version)

	return doc, op, version, nil
}

func (s *Service) SetPermission(ctx context.Context, tenantID, documentID, subjectID string, level AccessLevel) (Document, error) {
	doc, err := s.repo.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return Document{}, err
	}
	if doc.Permissions == nil {
		doc.Permissions = make(map[string]AccessLevel)
	}
	doc.Permissions[subjectID] = level
	doc.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return Document{}, fmt.Errorf("update permissions: %w", err)
	}
	return doc, nil
}

func (s *Service) CreateShareLink(ctx context.Context, tenantID, documentID, creatorID string, level AccessLevel, expiresAt *time.Time) (ShareLink, error) {
	doc, err := s.repo.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return ShareLink{}, err
	}

	link := ShareLink{
		ID:         NewID(),
		Token:      NewID(),
		Level:      level,
		ExpiresAt:  expiresAt,
		DocumentID: doc.ID,
		TenantID:   doc.TenantID,
		CreatedAt:  time.Now().UTC(),
		CreatedBy:  creatorID,
	}

	doc.ShareLinks = append(doc.ShareLinks, link)
	doc.UpdatedAt = link.CreatedAt
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return ShareLink{}, fmt.Errorf("create share link: %w", err)
	}
	return link, nil
}

func (s *Service) ListVersions(ctx context.Context, tenantID, documentID string, limit int) ([]DocumentVersion, error) {
	return s.repo.ListVersions(ctx, tenantID, documentID, limit)
}

func (s *Service) RevertToVersion(ctx context.Context, tenantID, documentID, versionID, userID string) (Document, DocumentVersion, error) {
	versions, err := s.repo.ListVersions(ctx, tenantID, documentID, 0)
	if err != nil {
		return Document{}, DocumentVersion{}, err
	}

	var target DocumentVersion
	for _, v := range versions {
		if v.ID == versionID {
			target = v
			break
		}
	}
	if target.ID == "" {
		return Document{}, DocumentVersion{}, fmt.Errorf("version not found: %s", versionID)
	}

	doc, err := s.repo.GetDocument(ctx, tenantID, documentID)
	if err != nil {
		return Document{}, DocumentVersion{}, err
	}

	now := time.Now().UTC()
	doc.Content = target.Content
	doc.Version++
	doc.UpdatedAt = now
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return Document{}, DocumentVersion{}, fmt.Errorf("revert version: %w", err)
	}

	restoreVersion := DocumentVersion{
		ID:         NewID(),
		DocumentID: doc.ID,
		TenantID:   tenantID,
		AuthorID:   userID,
		Sequence:   doc.Version,
		Content:    target.Content,
		Label:      "revert",
		CreatedAt:  now,
	}
	_ = s.repo.SaveVersion(ctx, restoreVersion)
	return doc, restoreVersion, nil
}

func NewID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
