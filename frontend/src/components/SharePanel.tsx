import { useState } from "react";
import type { AccessLevel, ShareLink } from "../types";

interface Props {
  shareLinks: ShareLink[];
  onCreateLink: (level: AccessLevel, expiresAt?: string) => void;
  onSetPermission: (subjectId: string, level: AccessLevel) => void;
}

const levels: AccessLevel[] = ["view", "comment", "edit"];

export function SharePanel({ shareLinks, onCreateLink, onSetPermission }: Props) {
  const [level, setLevel] = useState<AccessLevel>("view");
  const [subject, setSubject] = useState("");
  const [subjectLevel, setSubjectLevel] = useState<AccessLevel>("edit");

  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <div className="eyebrow">Sharing</div>
          <h3>Access controls</h3>
        </div>
      </div>
      <div className="panel-body">
        <div className="stack">
          <div className="row">
            <select value={level} onChange={(e) => setLevel(e.target.value as AccessLevel)}>
              {levels.map((l) => (
                <option key={l}>{l}</option>
              ))}
            </select>
            <button onClick={() => onCreateLink(level)}>Generate link</button>
          </div>
          <div className="share-links">
            {shareLinks.map((link) => (
              <div key={link.id} className="share-chip">
                <span className="pill">{link.level}</span>
                <code>{link.token}</code>
              </div>
            ))}
            {shareLinks.length === 0 && <div className="empty">No share links yet.</div>}
          </div>
        </div>

        <div className="stack">
          <div className="eyebrow">Direct permissions</div>
          <div className="row">
            <input
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="user-or-group-id"
              onKeyDown={(e) => e.key === "Enter" && onSetPermission(subject, subjectLevel)}
            />
            <select value={subjectLevel} onChange={(e) => setSubjectLevel(e.target.value as AccessLevel)}>
              {levels.map((l) => (
                <option key={l}>{l}</option>
              ))}
            </select>
            <button onClick={() => onSetPermission(subject, subjectLevel)}>Save</button>
          </div>
        </div>
      </div>
    </div>
  );
}
