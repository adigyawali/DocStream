package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"docStream/backend/internal/auth"
	"docStream/backend/internal/document"
	"docStream/backend/internal/httpapi"
	"docStream/backend/internal/realtime"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "docstream")
	
	// Handle Docker internal hostname vs localhost for dev
	dbPort := "5432"
	
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Initialize Repository
	repo := document.NewPostgresRepository(pool)
	
	// Ensure Schema (Simple migration for now)
	if err := repo.EnsureSchema(ctx); err != nil {
		log.Fatalf("Failed to ensure schema: %v\n", err)
	}

	// Services
	docService := document.NewService(repo)
	authService := auth.NewService(repo)

	hub := realtime.NewHub(docService)
	api := httpapi.New(docService, authService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "DocStream Backend is Running with Postgres!")
	})
	mux.Handle("/api/", api.Routes())
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add token validation for WS
		hub.ServeWS(w, r)
	})

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server started on port %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// withCORS is a small helper to allow the local Vite frontend to talk to the Go API.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Let the API handler manage its own CORS for specific routes if needed, 
		// but global fallback is good for development.
		// However, since we moved CORS logic inside ServeHTTP for API routes, 
		// we should be careful not to double-write headers. 
		// The API's ServeHTTP handles its own OPTIONS, so we can skip this for /api
		// or just set them here generally.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}