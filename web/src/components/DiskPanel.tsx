import { FormEvent, useEffect, useState } from "react";
import { Database, Plus, RefreshCw, Trash2 } from "lucide-react";
import { api, Disk } from "../api/client";

interface Props {
  selectedPath: string;
  onSelect: (path: string) => void;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

function formatBytes(bytes: number) {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function DiskPanel({ selectedPath, onSelect, onMessage }: Props) {
  const [disks, setDisks] = useState<Disk[]>([]);
  const [path, setPath] = useState("/home/eduardo/mia/cali/disco.dsk");
  const [size, setSize] = useState(20);
  const [unit, setUnit] = useState("M");
  const [fit, setFit] = useState("FF");
  const [busy, setBusy] = useState(false);

  async function loadDisks() {
    setBusy(true);
    try {
      const response = await api.disks();
      setDisks(response.data ?? []);
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    void loadDisks();
  }, []);

  async function createDisk(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    try {
      const response = await api.createDisk({ path, size, unit, fit });
      onMessage(response.message || "Disco creado", "success");
      await loadDisks();
      onSelect(path);
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  async function deleteDisk(disk: Disk) {
    if (!window.confirm(`Eliminar ${disk.name}?`)) return;
    setBusy(true);
    try {
      const response = await api.deleteDisk(disk.path);
      if (selectedPath === disk.path) onSelect("");
      onMessage(response.message || "Disco eliminado", "success");
      await loadDisks();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  return (
    <section className="tool-section">
      <div className="section-title">
        <Database size={17} />
        <h2>Discos</h2>
        <button
          className="icon-button"
          onClick={() => void loadDisks()}
          disabled={busy}
          title="Actualizar discos"
        >
          <RefreshCw size={16} />
        </button>
      </div>
      <div className="item-list">
        {disks.map((disk) => (
          <div
            className={`disk-row ${selectedPath === disk.path ? "selected" : ""}`}
            key={disk.path}
          >
            <button
              className="disk-select"
              onClick={() => onSelect(disk.path)}
              title={`Seleccionar ${disk.name}`}
            >
              <Database size={16} />
              <span>
                <strong>{disk.name}</strong>
                <small>{formatBytes(disk.size)}</small>
              </span>
            </button>
            <button
              className="icon-button danger"
              onClick={() => void deleteDisk(disk)}
              title={`Eliminar ${disk.name}`}
            >
              <Trash2 size={15} />
            </button>
          </div>
        ))}
        {!busy && disks.length === 0 && (
          <p className="empty-state">No hay discos disponibles.</p>
        )}
      </div>
      <details>
        <summary>
          <Plus size={15} /> Crear disco
        </summary>
        <form className="compact-form" onSubmit={createDisk}>
          <label>
            Ruta
            <input value={path} onChange={(event) => setPath(event.target.value)} required />
          </label>
          <div className="form-grid">
            <label>
              Tamano
              <input
                type="number"
                min="1"
                value={size}
                onChange={(event) => setSize(Number(event.target.value))}
                required
              />
            </label>
            <label>
              Unidad
              <select value={unit} onChange={(event) => setUnit(event.target.value)}>
                <option value="M">MB</option>
                <option value="K">KB</option>
              </select>
            </label>
            <label>
              Fit
              <select value={fit} onChange={(event) => setFit(event.target.value)}>
                <option value="FF">First Fit</option>
                <option value="BF">Best Fit</option>
                <option value="WF">Worst Fit</option>
              </select>
            </label>
          </div>
          <button className="primary-button" disabled={busy}>
            <Plus size={16} /> Crear
          </button>
        </form>
      </details>
    </section>
  );
}
