import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../state/auth";
import { useToast } from "../components/Toast";
import { ApiError } from "../lib/api";
import { clearDraft, loadDraft, saveDraft } from "../lib/drafts";

const DRAFT_KEY = "register";

type Draft = { email: string };

export default function RegisterPage() {
  const { register } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
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
    if (password.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }
    if (password !== confirm) {
      toast.error("Passwords do not match");
      return;
    }
    setSubmitting(true);
    try {
      await register(email.trim(), password);
      clearDraft(DRAFT_KEY);
      toast.success("Account created");
      navigate("/tasks", { replace: true });
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : "Registration failed";
      toast.error(msg);
    } finally {
      setSubmitting(false);
    }
  };

  const onCancel = () => {
    clearDraft(DRAFT_KEY);
    setEmail("");
    setPassword("");
    setConfirm("");
  };

  return (
    <div className="auth-wrap">
      <form className="card auth-card" onSubmit={onSubmit}>
        <h1>Create account</h1>
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
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
        </label>
        <label>
          Confirm password
          <input
            type="password"
            autoComplete="new-password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            required
            minLength={8}
          />
        </label>
        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel}>
            Clear
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "Creating…" : "Create account"}
          </button>
        </div>
        <p className="muted">
          Already registered? <Link to="/login">Sign in</Link>
        </p>
      </form>
    </div>
  );
}
