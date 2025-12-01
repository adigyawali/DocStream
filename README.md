# DocStream

Premium-grade real-time document collaboration: create docs, share securely, collaborate live, manage versions, and enforce permissions across tenants.

## What it does
- Live co-editing with WebSockets (CRDT/OT ready) and broadcast to all viewers.
- Version history with one-click revert; operations persisted for audit/replay.
- Secure sharing: per-subject permissions (view/comment/edit) and expirable share links.
- Multi-tenant isolation with Postgres storage; Redis-ready for fan-out (via Docker).
- Modern React frontend (Vite) with presence, sharing, and history UI.

## Architecture
- **Frontend:** React + Vite, Nginx in Docker for prod; dev proxy to backend.
- **Backend:** Go HTTP + WebSocket server; layered services; Postgres repository.
- **Collaboration:** Hub/Room model per document; operations applied server-side and broadcast.
- **Persistence:** Postgres tables for users, documents, permissions, share links, versions, operations.

## Quick start (Docker)
```bash
docker-compose up --build
```
- Frontend: http://localhost:3000
- Backend: http://localhost:8080 (proxied via Nginx for /api and /ws)

## Quick start (local dev, no Docker)
Backend:
```bash
cd backend
PORT=8081 go run cmd/server/main.go
```
Frontend (proxying to 8081):
```bash
cd frontend
npm install
npm run dev
```
Open the Vite URL (usually http://localhost:5173); /api and /ws are proxied to 8081.

## API overview
- `POST /api/signup` — create account
- `POST /api/login` — returns JWT + userId
- `GET/POST /api/tenants/{tenantId}/docs` — list/create documents
- `GET /api/tenants/{tenantId}/docs/{docId}` — fetch a document
- `POST /api/tenants/{tenantId}/docs/{docId}/permissions` — set subject access
- `POST /api/tenants/{tenantId}/docs/{docId}/share` — create share link
- `GET /api/tenants/{tenantId}/docs/{docId}/versions` — list versions
- `POST /api/tenants/{tenantId}/docs/{docId}/versions/{versionId}/revert` — restore version
- WebSocket: `/ws?tenantId=...&docId=...&userId=...` for live operations/presence

## Configuration
Env vars (backend):
- `PORT` (default 8080)
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (Postgres)
- `SECRET_KEY` (JWT signing)

Frontend:
- `VITE_API_BASE`, `VITE_WS_BASE` (optional; dev proxy covers defaults)

## Development notes
- Go modules are under `backend/`; run `go build ./...` to validate.
- Frontend builds with `npm run build`; production assets served by Nginx in Docker.
- Schema bootstrap via `EnsureSchema` on startup (lightweight migration).

## Status & next steps
- Core flows (auth, create, share, edit, version, revert) are wired.
- CRDT/OT plumbing is ready for a real engine; currently applies whole-document deltas.
- Add auth guards to WebSocket join and fine-grained permission checks per operation.
