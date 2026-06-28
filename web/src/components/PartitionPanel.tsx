import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  HardDrive,
  Link,
  Link2Off,
  Plus,
  RefreshCw,
  Scaling,
  Trash2,
  WandSparkles,
} from "lucide-react";
import { api, MountedPartition, Partition } from "../api/client";

interface Props {
  diskPath: string;
  refreshKey: number;
  activeId: string;
  sessionId: string;
  onActiveId: (id: string) => void;
  onSessionInvalidated: () => void;
  onDataChanged: () => void;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function PartitionPanel({
  diskPath,
  refreshKey,
  activeId,
  sessionId,
  onActiveId,
  onSessionInvalidated,
  onDataChanged,
  onMessage,
}: Props) {
  const [partitions, setPartitions] = useState<Partition[]>([]);
  const [mounts, setMounts] = useState<MountedPartition[]>([]);
  const [name, setName] = useState("Part1");
  const [size, setSize] = useState(15);
  const [unit, setUnit] = useState("M");
  const [type, setType] = useState("P");
  const [fit, setFit] = useState("FF");
  const [managedName, setManagedName] = useState("");
  const [resizeAmount, setResizeAmount] = useState(1);
  const [resizeUnit, setResizeUnit] = useState("M");
  const [deleteMode, setDeleteMode] = useState("fast");
  const [busy, setBusy] = useState(false);

  const mountedByName = useMemo(
    () =>
      new Map(
        mounts
          .filter((item) => item.path === diskPath)
          .map((item) => [item.name, item]),
      ),
    [diskPath, mounts],
  );

  async function load() {
    if (!diskPath) {
      setPartitions([]);
      return;
    }
    setBusy(true);
    try {
      const [partitionResponse, mountResponse] = await Promise.all([
        api.partitions(diskPath),
        api.mounts(),
      ]);
      const loaded = partitionResponse.data ?? [];
      setPartitions(loaded);
      setMounts(mountResponse.data ?? []);
      setManagedName((current) =>
        loaded.some((partition) => partition.name === current)
          ? current
          : loaded[0]?.name || "",
      );
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  useEffect(() => {
    void load();
  }, [diskPath, refreshKey]);

  async function createPartition(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    try {
      const response = await api.createPartition({
        path: diskPath,
        name,
        size,
        unit,
        type,
        fit,
      });
      onMessage(response.message || "Particion creada", "success");
      await load();
      onDataChanged();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  async function mountPartition(partition: Partition) {
    setBusy(true);
    try {
      const response = await api.mount(diskPath, partition.name);
      const mounted = response.data;
      if (mounted) onActiveId(mounted.id);
      onMessage(
        mounted ? `Particion montada: ${mounted.id}` : "Particion montada",
        "success",
      );
      await load();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  async function unmountPartition(mounted: MountedPartition) {
    setBusy(true);
    try {
      const response = await api.unmount(mounted.id);
      if (activeId.toLowerCase() === mounted.id.toLowerCase()) onActiveId("");
      if (sessionId.toLowerCase() === mounted.id.toLowerCase()) {
        onSessionInvalidated();
      }
      onMessage(response.message || "Particion desmontada", "success");
      await load();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  async function formatPartition(mounted: MountedPartition) {
    if (!window.confirm(`Formatear ${mounted.name} (${mounted.id})?`)) return;
    setBusy(true);
    try {
      const response = await api.mkfs(mounted.id);
      onActiveId(mounted.id);
      onMessage(response.message || "Particion formateada", "success");
      onDataChanged();
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  async function resizePartition(event: FormEvent) {
    event.preventDefault();
    if (!managedName) {
      onMessage("Selecciona una particion", "error");
      return;
    }
    setBusy(true);
    try {
      const response = await api.resizePartition({
        path: diskPath,
        name: managedName,
        add: resizeAmount,
        unit: resizeUnit,
      });
      onMessage(response.message || "Particion redimensionada", "success");
      await load();
      onDataChanged();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  async function deletePartition() {
    if (!managedName) {
      onMessage("Selecciona una particion", "error");
      return;
    }
    if (!window.confirm(`Eliminar ${managedName} en modo ${deleteMode}?`)) return;
    setBusy(true);
    try {
      const response = await api.deletePartition({
        path: diskPath,
        name: managedName,
        delete: deleteMode,
      });
      onMessage(response.message || "Particion eliminada", "success");
      await load();
      onDataChanged();
    } catch (error) {
      onMessage((error as Error).message, "error");
      setBusy(false);
    }
  }

  if (!diskPath) {
    return (
      <section className="tool-section">
        <div className="section-title"><HardDrive size={17} /><h2>Particiones</h2></div>
        <p className="empty-state">Selecciona un disco.</p>
      </section>
    );
  }

  return (
    <section className="tool-section">
      <div className="section-title">
        <HardDrive size={17} />
        <h2>Particiones</h2>
        <button className="icon-button" onClick={() => void load()} title="Actualizar">
          <RefreshCw size={16} />
        </button>
      </div>
      <p className="path-caption" title={diskPath}>{diskPath}</p>
      <div className="item-list">
        {partitions.map((partition) => {
          const mounted = mountedByName.get(partition.name);
          return (
            <div className="partition-row" key={`${partition.name}-${partition.start}`}>
              <div className="partition-main">
                <span className={`type-badge type-${partition.type.toLowerCase()}`}>
                  {partition.type}
                </span>
                <span>
                  <strong>{partition.name}</strong>
                  <small>{(partition.size / (1024 * 1024)).toFixed(2)} MB · {partition.fit}</small>
                </span>
              </div>
              <div className="row-actions">
                {mounted ? (
                  <>
                    <button
                      className="id-badge"
                      onClick={() => onActiveId(mounted.id)}
                      title={`Usar ${mounted.id} como ID activo`}
                    >
                      {mounted.id}
                    </button>
                    <button
                      className="icon-button"
                      onClick={() => void formatPartition(mounted)}
                      title="Formatear EXT2"
                    >
                      <WandSparkles size={15} />
                    </button>
                    <button
                      className="icon-button danger"
                      onClick={() => void unmountPartition(mounted)}
                      title="Desmontar"
                    >
                      <Link2Off size={15} />
                    </button>
                  </>
                ) : (
                  <button
                    className="icon-button"
                    onClick={() => void mountPartition(partition)}
                    title="Montar"
                  >
                    <Link size={15} />
                  </button>
                )}
              </div>
            </div>
          );
        })}
        {!busy && partitions.length === 0 && (
          <p className="empty-state">El disco no tiene particiones.</p>
        )}
      </div>
      <details>
        <summary><Plus size={15} /> Crear particion</summary>
        <form className="compact-form" onSubmit={createPartition}>
          <label>
            Nombre
            <input value={name} onChange={(event) => setName(event.target.value)} required />
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
                <option value="B">Bytes</option>
              </select>
            </label>
            <label>
              Tipo
              <select value={type} onChange={(event) => setType(event.target.value)}>
                <option value="P">Primaria</option>
                <option value="E">Extendida</option>
                <option value="L">Logica</option>
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
      <details>
        <summary><Scaling size={15} /> Administrar particion</summary>
        <div className="compact-form">
          <label>
            Particion
            <select
              value={managedName}
              onChange={(event) => setManagedName(event.target.value)}
              disabled={partitions.length === 0}
            >
              {partitions.map((partition) => (
                <option key={`manage-${partition.name}`} value={partition.name}>
                  {partition.name} ({partition.type})
                </option>
              ))}
            </select>
          </label>
          <form className="partition-operation" onSubmit={resizePartition}>
            <label>
              Add
              <input
                type="number"
                value={resizeAmount}
                onChange={(event) => setResizeAmount(Number(event.target.value))}
                required
              />
            </label>
            <label>
              Unidad
              <select
                value={resizeUnit}
                onChange={(event) => setResizeUnit(event.target.value)}
              >
                <option value="B">Bytes</option>
                <option value="K">KB</option>
                <option value="M">MB</option>
              </select>
            </label>
            <button
              className="primary-button"
              disabled={busy || !managedName || mountedByName.has(managedName)}
            >
              <Scaling size={15} /> Aplicar resize
            </button>
          </form>
          <div className="partition-operation delete-operation">
            <label>
              Modo delete
              <select
                value={deleteMode}
                onChange={(event) => setDeleteMode(event.target.value)}
              >
                <option value="fast">Fast</option>
                <option value="full">Full</option>
              </select>
            </label>
            <button
              className="danger-button"
              type="button"
              disabled={busy || !managedName || mountedByName.has(managedName)}
              onClick={() => void deletePartition()}
            >
              <Trash2 size={15} /> Eliminar particion
            </button>
          </div>
          {mountedByName.has(managedName) && (
            <p className="form-note">Desmonta la particion antes de modificarla.</p>
          )}
        </div>
      </details>
    </section>
  );
}
