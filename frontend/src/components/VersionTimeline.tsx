import type { DocumentVersion } from "../types";

interface Props {
  versions: DocumentVersion[];
  onRevert: (versionId: string) => void;
}

export function VersionTimeline({ versions, onRevert }: Props) {
  return (
    <div className="panel">
      <div className="panel-header">
        <div>
          <div className="eyebrow">Versions</div>
          <h3>History</h3>
        </div>
      </div>
      <div className="panel-body scrollable">
        {versions
          .slice()
          .sort((a, b) => b.sequence - a.sequence)
          .map((v) => (
            <div key={v.id} className="version-row">
              <div>
                <div className="doc-title">
                  #{v.sequence} {v.label || ""}
                </div>
                <div className="doc-meta">
                  by {v.authorId} • {new Date(v.createdAt).toLocaleTimeString()}
                </div>
              </div>
              <button onClick={() => onRevert(v.id)}>Revert</button>
            </div>
          ))}
        {versions.length === 0 && <div className="empty">No versions yet.</div>}
      </div>
    </div>
  );
}
