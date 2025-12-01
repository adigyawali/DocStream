import { useState } from "react";

interface Props {
  onSignupSuccess: () => void;
  onSwitchToLogin: () => void;
}

export function Signup({ onSignupSuccess, onSwitchToLogin }: Props) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const { signup } = await import("../api/client");
      await signup(email, password);
      onSignupSuccess();
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
          <h3>Create Account</h3>
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
            {loading ? "Creating Account..." : "Sign Up"}
          </button>
          <div className="auth-footer">
            Already have an account?{" "}
            <button type="button" className="link-button" onClick={onSwitchToLogin}>
              Sign In
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
