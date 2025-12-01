import { useCallback, useEffect, useMemo, useState } from "react";
import "./App.css";
import { DocumentList } from "./components/DocumentList";
import { DocumentEditor } from "./components/DocumentEditor";
import { SharePanel } from "./components/SharePanel";
import { VersionTimeline } from "./components/VersionTimeline";
import { PresenceBar } from "./components/PresenceBar";
import { Login } from "./components/Login";
import { Signup } from "./components/Signup";
import { useRealtimeCollaboration } from "./hooks/useRealtimeCollaboration";
import {
  createDocument,
  createShareLink,
  getDocument,
  listDocuments,
  listVersions,
  revertVersion,
  setPermission,
  setAuthToken,
} from "./api/client";
import type { AccessLevel, Doc, DocumentVersion } from "./types";

const tenantId = "demo-tenant";

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem("token"));
  const [userId, setUserId] = useState<string>(localStorage.getItem("userId") || "");
  const [authMode, setAuthMode] = useState<"login" | "signup">("login");

  const [view, setView] = useState<"list" | "editor">("list");
  const [documents, setDocuments] = useState<Doc[]>([]);
  const [selectedDocId, setSelectedDocId] = useState<string>();
  const [content, setContent] = useState("");
  const [versions, setVersions] = useState<DocumentVersion[]>([]);
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  useEffect(() => {
    setAuthToken(token);
  }, [token]);

  const handleLogin = (newToken: string, newUserId: string) => {
    localStorage.setItem("token", newToken);
    localStorage.setItem("userId", newUserId);
    setToken(newToken);
    setUserId(newUserId);
  };

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("userId");
    setToken(null);
    setUserId("");
  };

  const selectedDoc = useMemo(
    () => documents.find((doc) => doc.id === selectedDocId),
    [documents, selectedDocId],
  );

  const { status, lastMessage, sendOperation } = useRealtimeCollaboration({
    tenantId,
    docId: selectedDocId,
    userId, // Now real user ID
    onRemoteContent: setContent,
  });

  // Load documents when token changes (and is present)
  useEffect(() => {
    if (!token) return;
    listDocuments(tenantId)
      .then((docs) => setDocuments(docs))
      .catch((err) => {
        console.error(err);
        if (err.message.includes("Unauthorized")) handleLogout();
      });
  }, [token]);

  useEffect(() => {
    if (!selectedDocId || !token) return;
    getDocument(tenantId, selectedDocId)
      .then((doc) => {
        setContent(doc.content);
      })
      .catch(() => setContent(""));

    listVersions(tenantId, selectedDocId).then(setVersions).catch(() => setVersions([]));
  }, [selectedDocId, token]);

  const handleCreateDocument = async (title: string) => {
    try {
      const doc = await createDocument(tenantId, {
        title,
        content: "Welcome to DocStream — start collaborating!",
      });
      setDocuments((prev) => [...prev, doc]);
      setSelectedDocId(doc.id);
      setContent(doc.content);
      setView("editor");
    } catch (err) {
      console.error(err);
    }
  };

  const handleSelectDocument = (id: string) => {
    setSelectedDocId(id);
    setView("editor");
  };

  const handleBack = () => {
    setView("list");
    setSelectedDocId(undefined);
    setIsMenuOpen(false);
  };

  const handleContentChange = (next: string) => {
    setContent(next);
    sendOperation(next);
  };

  const handleShareLink = async (level: AccessLevel) => {
    if (!selectedDocId) return;
    try {
      const link = await createShareLink(tenantId, selectedDocId, { level });
      setDocuments((prev) =>
        prev.map((doc) => (doc.id === selectedDocId ? { ...doc, shareLinks: [...doc.shareLinks, link] } : doc)),
      );
    } catch (err) {
      console.error(err);
    }
  };

  const handlePermission = async (subjectId: string, level: AccessLevel) => {
    if (!selectedDocId || !subjectId) return;
    try {
      const updated = await setPermission(tenantId, selectedDocId, { subjectId, level });
      setDocuments((prev) => prev.map((doc) => (doc.id === updated.id ? updated : doc)));
    } catch (err) {
      console.error(err);
    }
  };

  const handleRevert = useCallback(
    async (versionId: string) => {
      if (!selectedDocId) return;
      try {
        const result = await revertVersion(tenantId, selectedDocId, versionId);
        setDocuments((prev) => prev.map((doc) => (doc.id === result.document.id ? result.document : doc)));
        setContent(result.document.content);
        setVersions((prev) => [...prev, result.version]);
      } catch (err) {
        console.error(err);
      }
    },
    [selectedDocId],
  );

  if (!token) {
    return authMode === "login" ? (
      <Login onLogin={handleLogin} onSwitchToSignup={() => setAuthMode("signup")} />
    ) : (
      <Signup onSignupSuccess={() => setAuthMode("login")} onSwitchToLogin={() => setAuthMode("login")} />
    );
  }

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">DocStream</p>
          <h1>Secure, real-time documents for teams</h1>
          <p className="lede">
            Create docs, share with teammates, collaborate live, and roll back with a click.
          </p>
        </div>
        <div style={{ display: "flex", gap: "12px", alignItems: "center" }}>
          <PresenceBar status={status} lastMessage={lastMessage} />
          <button onClick={handleLogout}>Sign Out</button>
        </div>
      </header>

      {view === "list" ? (
        <div className="grid doc-list-container">
          <DocumentList
            documents={documents}
            selectedId={selectedDocId}
            onSelect={handleSelectDocument}
            onCreate={handleCreateDocument}
          />
        </div>
      ) : (
        <div>
          <button className="back-button" onClick={handleBack}>
            ← Back to documents
          </button>
          <div className="grid" style={{ gridTemplateColumns: "1fr" }}>
            <DocumentEditor
              title={selectedDoc?.title}
              content={content}
              onChange={handleContentChange}
              connectionStatus={status}
              onOpenMenu={() => setIsMenuOpen(true)}
            />
          </div>
        </div>
      )}

      {isMenuOpen && selectedDoc && (
        <div className="modal-overlay" onClick={() => setIsMenuOpen(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Document Settings</h3>
              <button className="close-button" onClick={() => setIsMenuOpen(false)}>
                ×
              </button>
            </div>
            <div className="panel-body">
              <SharePanel
                shareLinks={selectedDoc.shareLinks || []}
                onCreateLink={handleShareLink}
                onSetPermission={handlePermission}
              />
              <VersionTimeline versions={versions} onRevert={handleRevert} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;