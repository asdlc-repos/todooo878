import { useEffect, useState, type FormEvent } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../state/auth";
import { useToast } from "../components/Toast";
import { ApiError } from "../lib/api";
import { clearDraft, loadDraft, saveDraft } from "../lib/drafts";

const DRAFT_KEY = "login";

type Draft = { email: string };

export default function LoginPage() {
  const { login } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const redirectTo =
    (location.state as { from?: { pathname?: string } } | null)?.from?.pathname ??
    "/tasks";

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const draft = loadDraft<Draft>(DRAFT_KEY);
    if (draft?.email) setEmail(draft.email);
  }, []);

  useEffect(() => {
    saveDraft<Draft>(DRAFT_KEY, { email });
  }, [email]);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    try {
      await login(email.trim(), password);
      clearDraft(DRAFT_KEY);
      toast.success("Welcome back");
      navigate(redirectTo, { replace: true });
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Login failed";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const onCancel = () => {
    clearDraft(DRAFT_KEY);
    setEmail("");
    setPassword("");
  };

  return (
    <div className="auth-wrap">
      <form className="card auth-card" onSubmit={onSubmit}>
        <h1>Sign in</h1>
        <label>
          Email
          <input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Password
          <input
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </label>
        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel}>
            Clear
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "Signing in…" : "Sign in"}
          </button>
        </div>
        <p className="muted">
          No account? <Link to="/register">Create one</Link>
        </p>
      </form>
    </div>
  );
}
