import type { CollabMessage } from "../types";

interface Props {
  status: string;
  lastMessage: CollabMessage | null;
}

export function PresenceBar({ status, lastMessage }: Props) {
  return (
    <div className="panel presence">
      <div className="panel-header">
        <div className="eyebrow">Realtime</div>
        <h3>Presence</h3>
      </div>
      <div className="panel-body">
        <div className="status-dot">
          <span className={`dot ${status}`} />
          <span className="status-label">{status}</span>
        </div>
        {lastMessage ? (
          <div className="mini-log">
            <div className="eyebrow">Last event</div>
            <code>{JSON.stringify(lastMessage)}</code>
          </div>
        ) : (
          <div className="empty">Waiting for activity...</div>
        )}
      </div>
    </div>
  );
}
