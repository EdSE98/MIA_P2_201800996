import { useEffect, useState } from "react";
import { Activity, Boxes, Server, X } from "lucide-react";
import { api, CommandExecution, Session } from "./api/client";
import { CommandConsole } from "./components/CommandConsole";
import { DiskPanel } from "./components/DiskPanel";
import { FileExplorer } from "./components/FileExplorer";
import { FileOperationsPanel } from "./components/FileOperationsPanel";
import { LoginPanel } from "./components/LoginPanel";
import { PartitionPanel } from "./components/PartitionPanel";
import { ReportsPanel } from "./components/ReportsPanel";

interface Notice {
  text: string;
  kind: "success" | "error";
}

export default function App() {
  const [selectedDisk, setSelectedDisk] = useState("");
  const [activeId, setActiveId] = useState("");
  const [session, setSession] = useState<Session | null>(null);
  const [apiOnline, setApiOnline] = useState(false);
  const [notice, setNotice] = useState<Notice | null>(null);
  const [dataRevision, setDataRevision] = useState(0);
  const [currentFSPath, setCurrentFSPath] = useState("/");

  function showMessage(text: string, kind: "success" | "error" = "success") {
    setNotice({ text, kind });
  }

  function dataChanged() {
    setDataRevision((current) => current + 1);
  }

  function sessionChanged(nextSession: Session | null) {
    setSession(nextSession);
    if (nextSession?.mountedId) setActiveId(nextSession.mountedId);
  }

  async function commandExecuted(result: CommandExecution) {
    setSession(result.session ?? null);
    if (result.session?.mountedId) {
      setActiveId(result.session.mountedId);
    } else {
      const mountedMatch = result.output.match(/ID:\s*([A-Za-z0-9]+)/i);
      if (mountedMatch) setActiveId(mountedMatch[1]);
    }
    dataChanged();
    try {
      const response = await api.mounts();
      const mounts = response.data ?? [];
      setActiveId((current) =>
        current && mounts.some((mounted) => mounted.id.toLowerCase() === current.toLowerCase())
          ? current
          : result.session?.mountedId || "",
      );
    } catch {
      // The command output remains useful even if mount reconciliation fails.
    }
  }

  useEffect(() => {
    api.health().then(
      () => setApiOnline(true),
      () => setApiOnline(false),
    );
  }, []);

  useEffect(() => {
    if (!notice) return;
    const timeout = window.setTimeout(() => setNotice(null), 5000);
    return () => window.clearTimeout(timeout);
  }, [notice]);

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark"><Boxes size={21} /></span>
          <div>
            <strong>MIA Explorer</strong>
            <span>Proyecto Fase 2 · 201800996</span>
          </div>
        </div>
        <div className="topbar-status">
          {activeId && <span className="active-id">ID {activeId}</span>}
          <span className={`api-status ${apiOnline ? "online" : "offline"}`}>
            <Activity size={14} />
            API {apiOnline ? "conectada" : "sin conexion"}
          </span>
          <Server size={18} />
        </div>
      </header>

      <main className="dashboard">
        <aside className="left-panel">
          <LoginPanel
            activeId={activeId}
            session={session}
            onSessionChange={sessionChanged}
            onMessage={showMessage}
          />
          <DiskPanel
            selectedPath={selectedDisk}
            refreshKey={dataRevision}
            onSelect={setSelectedDisk}
            onMessage={showMessage}
          />
          <PartitionPanel
            diskPath={selectedDisk}
            refreshKey={dataRevision}
            activeId={activeId}
            sessionId={session?.mountedId || ""}
            onActiveId={setActiveId}
            onSessionInvalidated={() => setSession(null)}
            onDataChanged={dataChanged}
            onMessage={showMessage}
          />
        </aside>

        <FileExplorer
          activeId={activeId}
          refreshKey={dataRevision}
          onPathChange={setCurrentFSPath}
          onMessage={showMessage}
        />

        <aside className="right-panel">
          <CommandConsole
            onExecuted={(result) => void commandExecuted(result)}
            onMessage={showMessage}
          />
          <FileOperationsPanel
            enabled={Boolean(session)}
            onChanged={dataChanged}
            onMessage={showMessage}
          />
          <ReportsPanel
            activeId={activeId}
            currentPath={currentFSPath}
            refreshKey={dataRevision}
            onMessage={showMessage}
          />
          <section className="tool-section system-summary">
            <div className="section-title"><Server size={17} /><h2>Contexto</h2></div>
            <dl>
              <div><dt>Disco</dt><dd title={selectedDisk}>{selectedDisk || "Sin seleccionar"}</dd></div>
              <div><dt>Montaje</dt><dd>{activeId || "Ninguno"}</dd></div>
              <div><dt>Usuario</dt><dd>{session?.user || "Sin sesion"}</dd></div>
            </dl>
          </section>
        </aside>
      </main>

      {notice && (
        <div className={`toast ${notice.kind}`} role="status">
          <span>{notice.text}</span>
          <button onClick={() => setNotice(null)} title="Cerrar mensaje">
            <X size={16} />
          </button>
        </div>
      )}
    </div>
  );
}
