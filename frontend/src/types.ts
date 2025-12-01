export type AccessLevel = "view" | "comment" | "edit";

export interface Permission {
  subjectId: string;
  subjectType: "user" | "group" | "link";
  level: AccessLevel;
}

export interface ShareLink {
  id: string;
  documentId: string;
  token: string;
  level: AccessLevel;
  createdAt: string;
  createdBy: string;
  expiresAt?: string;
}

export interface DocumentVersion {
  id: string;
  documentId: string;
  tenantId: string;
  authorId: string;
  sequence: number;
  content: string;
  label: string;
  createdAt: string;
}

export interface Doc {
  id: string;
  tenantId: string;
  title: string;
  content: string;
  ownerId: string;
  permissions: Record<string, AccessLevel>;
  shareLinks: ShareLink[];
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface Operation {
  id: string;
  documentId: string;
  tenantId: string;
  userId: string;
  lamport: number;
  delta: string;
  createdAt: string;
}

export type CollabMessage =
  | { type: "snapshot"; tenantId: string; documentId: string; version: number; content: string }
  | { type: "update"; tenantId: string; documentId: string; userId: string; version: number; content: string }
  | { type: "presence"; tenantId: string; documentId: string; userId: string; message: string }
  | { type: "ack"; tenantId: string; documentId: string; userId: string; message: string }
  | { type: "error"; tenantId: string; documentId: string; userId: string; message: string };
