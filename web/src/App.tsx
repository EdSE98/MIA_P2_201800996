import { useEffect, useState } from "react";
import { Activity, Boxes, Server, X } from "lucide-react";
import { api, Session } from "./api/client";
import { DiskPanel } from "./components/DiskPanel";
import { FileExplorer } from "./components/FileExplorer";
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

  function showMessage(text: string, kind: "success" | "error" = "success") {
    setNotice({ text, kind });
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
            onSessionChange={setSession}
            onMessage={showMessage}
          />
          <DiskPanel
            selectedPath={selectedDisk}
            onSelect={setSelectedDisk}
            onMessage={showMessage}
          />
          <PartitionPanel
            diskPath={selectedDisk}
            activeId={activeId}
            onActiveId={setActiveId}
            onMessage={showMessage}
          />
        </aside>

        <FileExplorer activeId={activeId} onMessage={showMessage} />

        <aside className="right-panel">
          <ReportsPanel
            activeId={activeId}
            currentPath="/"
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
