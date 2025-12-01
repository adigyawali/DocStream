interface Props {
  title?: string;
  content: string;
  onChange: (next: string) => void;
  connectionStatus: string;
  onOpenMenu: () => void;
}

export function DocumentEditor({ title, content, onChange, connectionStatus, onOpenMenu }: Props) {
  return (
    <div className="panel editor">
      <div className="panel-header">
        <div>
          <div className="eyebrow">Editing</div>
          <h3>{title || "Untitled document"}</h3>
        </div>
        <div className="editor-header-controls">
          <div className="status-dot">
            <span className={`dot ${connectionStatus}`} />
            <span className="status-label">{connectionStatus}</span>
          </div>
          <button onClick={onOpenMenu}>Menu</button>
        </div>
      </div>
      <div className="panel-body">
        <textarea
          value={content}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Start typing... changes stream to collaborators in real time."
        />
      </div>
    </div>
  );
}
