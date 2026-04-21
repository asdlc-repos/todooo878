import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "../state/auth";
import { useToast } from "./Toast";
import type { ReactNode } from "react";

export function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  const onLogout = async () => {
    try {
      await logout();
      toast.success("Signed out");
      navigate("/login", { replace: true });
    } catch {
      toast.error("Failed to sign out");
    }
  };

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="brand">Todooo</div>
        <nav className="nav">
          <NavLink to="/tasks" className={({ isActive }) => (isActive ? "active" : "")}>
            Tasks
          </NavLink>
          <NavLink
            to="/categories"
            className={({ isActive }) => (isActive ? "active" : "")}
          >
            Categories
          </NavLink>
        </nav>
        <div className="user-area">
          {user && <span className="user-email">{user.email}</span>}
          <button className="btn btn-ghost" onClick={onLogout} type="button">
            Sign out
          </button>
        </div>
      </header>
      <main className="app-main">{children}</main>
    </div>
  );
}
