import type { AccessLevel, CollabMessage, Doc, DocumentVersion, ShareLink } from "../types";

// Use relative path so Nginx can proxy to backend
const API_BASE = import.meta.env.VITE_API_BASE ?? "";
// Infer WebSocket URL from current location
const WS_BASE = import.meta.env.VITE_WS_BASE ?? 
  ((window.location.protocol === "https:" ? "wss://" : "ws://") + window.location.host);

let authToken: string | null = null;

export function setAuthToken(token: string | null) {
  authToken = token;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (authToken) {
    // @ts-ignore
    headers["Authorization"] = `Bearer ${authToken}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  // Some endpoints (e.g., signup) return an empty body with 201/204.
  const raw = await res.text();
  if (!raw) {
    return undefined as T;
  }
  return JSON.parse(raw) as T;
}

export async function signup(email: string, password: string): Promise<void> {
  await request("/api/signup", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function login(email: string, password: string): Promise<{ token: string; userId: string }> {
  return request<{ token: string; userId: string }>("/api/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function createDocument(tenantId: string, input: { title: string; content?: string }) {
  return request<Doc>(`/api/tenants/${tenantId}/docs`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listDocuments(tenantId: string) {
  return request<Doc[]>(`/api/tenants/${tenantId}/docs`);
}

export async function getDocument(tenantId: string, docId: string) {
  return request<Doc>(`/api/tenants/${tenantId}/docs/${docId}`);
}

export async function createShareLink(
  tenantId: string,
  docId: string,
  input: { level: AccessLevel },
): Promise<ShareLink> {
  return request<ShareLink>(`/api/tenants/${tenantId}/docs/${docId}/share`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function setPermission(
  tenantId: string,
  docId: string,
  input: { subjectId: string; level: AccessLevel },
): Promise<Doc> {
  return request<Doc>(`/api/tenants/${tenantId}/docs/${docId}/permissions`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listVersions(tenantId: string, docId: string, limit?: number) {
  const query = limit ? `?limit=${limit}` : "";
  return request<DocumentVersion[]>(`/api/tenants/${tenantId}/docs/${docId}/versions${query}`);
}

export async function revertVersion(tenantId: string, docId: string, versionId: string) {
  // userId is now inferred from token
  return request<{ document: Doc; version: DocumentVersion }>(
    `/api/tenants/${tenantId}/docs/${docId}/versions/${versionId}/revert`,
    {
      method: "POST",
      body: JSON.stringify({}),
    },
  );
}

export function openCollabSocket(params: {
  tenantId: string;
  docId: string;
  userId: string;
  onMessage: (msg: CollabMessage) => void;
}): WebSocket {
  const { tenantId, docId, userId, onMessage } = params;
  const socket = new WebSocket(`${WS_BASE}/ws?tenantId=${tenantId}&docId=${docId}&userId=${userId}`);
  socket.onmessage = (event) => {
    try {
      const parsed: CollabMessage = JSON.parse(event.data);
      onMessage(parsed);
    } catch (err) {
      console.warn("failed to parse message", err);
    }
  };
  return socket;
}
