package document

import "errors"

var (
	// ErrDocumentNotFound is returned when a requested document is missing.
	ErrDocumentNotFound = errors.New("document not found")
)
