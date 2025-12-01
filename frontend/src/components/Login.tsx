import { useState } from "react";

interface Props {
  onLogin: (token: string, userId: string) => void;
  onSwitchToSignup: () => void;
}

export function Login({ onLogin, onSwitchToSignup }: Props) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      // Dynamic import to avoid circular dependency issues if client imports App
      const { login } = await import("../api/client");
      const { token, userId } = await login(email, password);
      onLogin(token, userId);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="panel auth-panel">
        <div className="panel-header">
          <h3>Sign In</h3>
        </div>
        <form className="panel-body" onSubmit={handleSubmit}>
          {error && <div className="error-banner">{error}</div>}
          <div className="stack">
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button className="primary" type="submit" disabled={loading}>
            {loading ? "Signing in..." : "Sign In"}
          </button>
          <div className="auth-footer">
            Don't have an account?{" "}
            <button type="button" className="link-button" onClick={onSwitchToSignup}>
              Sign Up
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
