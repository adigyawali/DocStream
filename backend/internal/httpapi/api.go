package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"docStream/backend/internal/auth"
	"docStream/backend/internal/document"

	"github.com/golang-jwt/jwt/v5"
)

// API wires HTTP handlers to the document service.
type API struct {
	docs *document.Service
	auth *auth.Service
}

func New(docs *document.Service, auth *auth.Service) *API {
	return &API{docs: docs, auth: auth}
}

// Routes returns an http.Handler with all API endpoints mounted.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", a)
	return mux
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers (simple version)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Auth routes
	if len(parts) == 2 && parts[1] == "signup" && r.Method == http.MethodPost {
		a.handleSignup(w, r)
		return
	}
	if len(parts) == 2 && parts[1] == "login" && r.Method == http.MethodPost {
		a.handleLogin(w, r)
		return
	}

	// Protected routes middleware
	tokenString := extractToken(r)
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return auth.SecretKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, _ := claims["sub"].(string)
	ctx := context.WithValue(r.Context(), "userID", userID)
	r = r.WithContext(ctx)

	// Expected: /api/tenants/{tenantId}/docs...
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "tenants" {
		http.NotFound(w, r)
		return
	}
	tenantID := parts[2]

	if len(parts) == 4 && parts[3] == "docs" {
		switch r.Method {
		case http.MethodGet:
			a.listDocuments(w, r, tenantID)
		case http.MethodPost:
			a.createDocument(w, r, tenantID)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) >= 5 && parts[3] == "docs" {
		docID := parts[4]
		if len(parts) == 5 {
			switch r.Method {
			case http.MethodGet:
				a.getDocument(w, r, tenantID, docID)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if len(parts) == 6 && parts[5] == "share" && r.Method == http.MethodPost {
			a.createShareLink(w, r, tenantID, docID)
			return
		}
		if len(parts) == 6 && parts[5] == "permissions" && r.Method == http.MethodPost {
			a.setPermission(w, r, tenantID, docID)
			return
		}
		if len(parts) == 6 && parts[5] == "versions" && r.Method == http.MethodGet {
			a.listVersions(w, r, tenantID, docID)
			return
		}
		if len(parts) == 8 && parts[5] == "versions" && parts[7] == "revert" && r.Method == http.MethodPost {
			versionID := parts[6]
			a.revertVersion(w, r, tenantID, docID, versionID)
			return
		}
	}

	http.NotFound(w, r)
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func (a *API) handleSignup(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if _, err := a.auth.Register(r.Context(), req.Email, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	token, userID, err := a.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token, "userId": userID})
}

func (a *API) createDocument(w http.ResponseWriter, r *http.Request, tenantID string) {
	type request struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(string)

	doc, err := a.docs.CreateDocument(r.Context(), tenantID, userID, req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (a *API) getDocument(w http.ResponseWriter, r *http.Request, tenantID, docID string) {
	doc, err := a.docs.GetDocument(r.Context(), tenantID, docID)
	if err != nil {
		if err == document.ErrDocumentNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: Check permission
	writeJSON(w, http.StatusOK, doc)
}

func (a *API) listDocuments(w http.ResponseWriter, r *http.Request, tenantID string) {
	docs, err := a.docs.ListDocuments(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (a *API) createShareLink(w http.ResponseWriter, r *http.Request, tenantID, docID string) {
	type request struct {
		Level string `json:"level"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(string)

	link, err := a.docs.CreateShareLink(r.Context(), tenantID, docID, userID, document.AccessLevel(req.Level), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (a *API) setPermission(w http.ResponseWriter, r *http.Request, tenantID, docID string) {
	type request struct {
		SubjectID string `json:"subjectId"`
		Level     string `json:"level"`
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doc, err := a.docs.SetPermission(r.Context(), tenantID, docID, req.SubjectID, document.AccessLevel(req.Level))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (a *API) listVersions(w http.ResponseWriter, r *http.Request, tenantID, docID string) {
	limit := 0
	if value := r.URL.Query().Get("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}
	versions, err := a.docs.ListVersions(r.Context(), tenantID, docID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, versions)
}

func (a *API) revertVersion(w http.ResponseWriter, r *http.Request, tenantID, docID, versionID string) {
	userID := r.Context().Value("userID").(string)

	doc, version, err := a.docs.RevertToVersion(r.Context(), tenantID, docID, versionID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"document": doc,
		"version":  version,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}