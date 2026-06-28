import { useEffect, useState } from "react";
import {
  ArrowUp,
  FileText,
  Folder,
  FolderOpen,
  RefreshCw,
} from "lucide-react";
import { api, FileContent, FSItem, Metadata } from "../api/client";

interface Props {
  activeId: string;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

function parentPath(path: string) {
  if (path === "/") return "/";
  const parts = path.split("/").filter(Boolean);
  parts.pop();
  return parts.length ? `/${parts.join("/")}` : "/";
}

export function FileExplorer({ activeId, onMessage }: Props) {
  const [path, setPath] = useState("/");
  const [items, setItems] = useState<FSItem[]>([]);
  const [selected, setSelected] = useState<Metadata | null>(null);
  const [content, setContent] = useState<FileContent | null>(null);
  const [busy, setBusy] = useState(false);

  async function loadDirectory(nextPath = path) {
    if (!activeId) {
      setItems([]);
      return;
    }
    setBusy(true);
    try {
      const response = await api.listFS(activeId, nextPath);
      setItems(response.data?.items ?? []);
      setPath(nextPath);
      setSelected(null);
      setContent(null);
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    setPath("/");
    void loadDirectory("/");
  }, [activeId]);

  async function openItem(item: FSItem) {
    if (item.type === "directory") {
      await loadDirectory(item.path);
      return;
    }
    setBusy(true);
    try {
      const [readResponse, statResponse] = await Promise.all([
        api.readFS(activeId, item.path),
        api.statFS(activeId, item.path),
      ]);
      setContent(readResponse.data ?? null);
      setSelected(statResponse.data ?? null);
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="workspace-section explorer-section">
      <div className="workspace-header">
        <div>
          <span className="eyebrow">EXT2 · {activeId || "sin montaje"}</span>
          <h1>Explorador de archivos</h1>
        </div>
        <div className="explorer-actions">
          <button
            className="icon-button"
            disabled={!activeId || path === "/"}
            onClick={() => void loadDirectory(parentPath(path))}
            title="Subir al directorio padre"
          >
            <ArrowUp size={18} />
          </button>
          <button
            className="icon-button"
            disabled={!activeId}
            onClick={() => void loadDirectory()}
            title="Actualizar directorio"
          >
            <RefreshCw size={18} />
          </button>
        </div>
      </div>
      <div className="breadcrumb">
        <FolderOpen size={16} />
        <span>{path}</span>
      </div>
      {!activeId ? (
        <div className="large-empty">
          <FolderOpen size={34} />
          <strong>Selecciona o monta una particion</strong>
          <span>El contenido EXT2 aparecera en este panel.</span>
        </div>
      ) : (
        <div className="file-layout">
          <div className="file-table" aria-busy={busy}>
            <div className="file-row file-head">
              <span>Nombre</span><span>Tipo</span><span>Tamano</span><span>Permisos</span>
            </div>
            {items.map((item) => (
              <button
                className={`file-row ${selected?.path === item.path ? "selected" : ""}`}
                key={`${item.inode}-${item.path}`}
                onClick={() => void openItem(item)}
              >
                <span className="file-name">
                  {item.type === "directory" ? <Folder size={17} /> : <FileText size={17} />}
                  {item.name}
                </span>
                <span>{item.type === "directory" ? "Carpeta" : "Archivo"}</span>
                <span>{item.size} B</span>
                <span className="mono">{item.permissions}</span>
              </button>
            ))}
            {!busy && items.length === 0 && (
              <p className="empty-state">Este directorio esta vacio.</p>
            )}
          </div>
          <aside className="file-preview">
            {content && selected ? (
              <>
                <div className="preview-title">
                  <FileText size={18} />
                  <div>
                    <strong>{content.name}</strong>
                    <span>{selected.owner}:{selected.group} · {selected.permissions} · {content.size} B</span>
                  </div>
                </div>
                <pre>{content.content}</pre>
              </>
            ) : (
              <div className="preview-empty">
                <FileText size={26} />
                <span>Selecciona un archivo para inspeccionarlo.</span>
              </div>
            )}
          </aside>
        </div>
      )}
    </section>
  );
}
