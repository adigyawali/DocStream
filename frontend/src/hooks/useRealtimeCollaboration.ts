import { useEffect, useRef, useState } from "react";
import type { CollabMessage } from "../types";
import { openCollabSocket } from "../api/client";

type Status = "idle" | "connecting" | "connected" | "disconnected";

interface Params {
  tenantId: string;
  docId?: string;
  userId: string;
  onRemoteContent?: (content: string) => void;
}

export function useRealtimeCollaboration({ tenantId, docId, userId, onRemoteContent }: Params) {
  const [status, setStatus] = useState<Status>("idle");
  const [lastMessage, setLastMessage] = useState<CollabMessage | null>(null);
  const socketRef = useRef<WebSocket | null>(null);
  const lamportRef = useRef<number>(0);

  useEffect(() => {
    if (!docId) {
      return;
    }
    setStatus("connecting");
    const socket = openCollabSocket({
      tenantId,
      docId,
      userId,
      onMessage: (msg) => {
        setLastMessage(msg);
        if (msg.type === "snapshot" || msg.type === "update") {
          onRemoteContent?.(msg.content);
          lamportRef.current = msg.version;
        }
      },
    });
    socketRef.current = socket;
    socket.onopen = () => setStatus("connected");
    socket.onclose = () => setStatus("disconnected");
    return () => socket.close();
  }, [tenantId, docId, userId, onRemoteContent]);

  const sendOperation = (content: string) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN || !docId) {
      return;
    }
    lamportRef.current += 1;
    const payload = {
      type: "operation",
      tenantId,
      documentId: docId,
      userId,
      newContent: content,
      delta: "naive-full-sync",
      lamport: lamportRef.current,
    };
    socketRef.current.send(JSON.stringify(payload));
  };

  return {
    status,
    lastMessage,
    sendOperation,
  };
}
