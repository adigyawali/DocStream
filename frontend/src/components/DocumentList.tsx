import { useState } from "react";
import type { Doc } from "../types";

interface Props {
  documents: Doc[];
  selectedId?: string;
  onSelect: (id: string) => void;
  onCreate: (title: string) => void;
}

export function DocumentList({ documents, selectedId, onSelect, onCreate }: Props) {
  const [title, setTitle] = useState("");

  const handleCreate = () => {
    if (!title.trim()) return;
    onCreate(title.trim());
    setTitle("");
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <div className="eyebrow">Workspace</div>
          <h3>Documents</h3>
        </div>
      </div>
      <div className="panel-body scrollable">
        {documents.map((doc) => (
          <button
            key={doc.id}
            className={`doc-pill ${selectedId === doc.id ? "active" : ""}`}
            onClick={() => onSelect(doc.id)}
          >
            <div className="doc-title">{doc.title || "Untitled"}</div>
          </button>
        ))}
        {documents.length === 0 && <div className="empty">No docs yet — create the first one.</div>}
      </div>
      <div className="panel-footer">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="New doc title"
          onKeyDown={(e) => e.key === "Enter" && handleCreate()}
        />
        <button className="primary" onClick={handleCreate}>
          Create
        </button>
      </div>
    </div>
  );
}
