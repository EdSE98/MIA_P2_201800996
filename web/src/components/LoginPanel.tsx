import { FormEvent, useEffect, useState } from "react";
import { KeyRound, LogIn, LogOut, UserRound } from "lucide-react";
import { api, Session } from "../api/client";

interface Props {
  activeId: string;
  session: Session | null;
  onSessionChange: (session: Session | null) => void;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function LoginPanel({
  activeId,
  session,
  onSessionChange,
  onMessage,
}: Props) {
  const [id, setId] = useState(activeId);
  const [user, setUser] = useState("root");
  const [password, setPassword] = useState("123");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (activeId) setId(activeId);
  }, [activeId]);

  async function handleLogin(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    try {
      const response = await api.login({ id, user, password });
      onSessionChange(response.data ?? null);
      onMessage(response.message || "Sesion iniciada", "success");
      setPassword("");
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  async function handleLogout() {
    setBusy(true);
    try {
      const response = await api.logout();
      onSessionChange(null);
      onMessage(response.message || "Sesion cerrada", "success");
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="tool-section">
      <div className="section-title">
        <UserRound size={17} />
        <h2>Sesion</h2>
      </div>
      {session ? (
        <div className="session-row">
          <div>
            <strong>{session.user}</strong>
            <span>{session.mountedId} · {session.group}</span>
          </div>
          <button
            className="icon-button danger"
            onClick={handleLogout}
            disabled={busy}
            title="Cerrar sesion"
          >
            <LogOut size={17} />
          </button>
        </div>
      ) : (
        <form className="compact-form" onSubmit={handleLogin}>
          <label>
            ID de particion
            <input
              value={id}
              onChange={(event) => setId(event.target.value)}
              placeholder="961A"
              required
            />
          </label>
          <label>
            Usuario
            <input
              value={user}
              onChange={(event) => setUser(event.target.value)}
              autoComplete="username"
              required
            />
          </label>
          <label>
            Password
            <div className="input-with-icon">
              <KeyRound size={15} />
              <input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                autoComplete="current-password"
                required
              />
            </div>
          </label>
          <button className="primary-button" disabled={busy}>
            <LogIn size={16} />
            Iniciar sesion
          </button>
        </form>
      )}
    </section>
  );
}
