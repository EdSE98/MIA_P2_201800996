import { FormEvent, useState } from "react";
import {
  Copy as CopyIcon,
  FilePenLine,
  FolderInput,
  MoveRight,
  Trash2,
  Type,
  Wrench,
} from "lucide-react";
import { api } from "../api/client";

type Operation = "edit" | "rename" | "remove" | "copy" | "move";

const operations: Array<{
  id: Operation;
  label: string;
  icon: typeof FilePenLine;
}> = [
  { id: "edit", label: "Edit", icon: FilePenLine },
  { id: "rename", label: "Rename", icon: Type },
  { id: "remove", label: "Remove", icon: Trash2 },
  { id: "copy", label: "Copy", icon: CopyIcon },
  { id: "move", label: "Move", icon: MoveRight },
];

interface Props {
  enabled: boolean;
  onChanged: () => void;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function FileOperationsPanel({
  enabled,
  onChanged,
  onMessage,
}: Props) {
  const [operation, setOperation] = useState<Operation>("edit");
  const [path, setPath] = useState("/home/docs/a.txt");
  const [contentPath, setContentPath] = useState("/tmp/nuevo_contenido.txt");
  const [newName, setNewName] = useState("b1.txt");
  const [destination, setDestination] = useState("/home/images");
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!enabled) {
      onMessage("Necesita una sesion activa", "error");
      return;
    }
    if (
      operation === "remove" &&
      !window.confirm(`Eliminar ${path} de forma recursiva?`)
    ) {
      return;
    }

    setBusy(true);
    try {
      let message = "";
      if (operation === "edit") {
        const response = await api.editFile(path, contentPath);
        message = response.message || "Archivo editado";
      } else if (operation === "rename") {
        const response = await api.renamePath(path, newName);
        message = response.message || "Ruta renombrada";
      } else if (operation === "remove") {
        const response = await api.removePath(path);
        message = response.message || "Ruta eliminada";
      } else if (operation === "copy") {
        const response = await api.copyPath(path, destination);
        message = response.message || "Ruta copiada";
        const warnings = response.data?.warnings ?? [];
        if (warnings.length > 0) {
          message += ` · Advertencias: ${warnings.join("; ")}`;
        }
      } else {
        const response = await api.movePath(path, destination);
        message = response.message || "Ruta movida";
      }
      onMessage(message, "success");
      onChanged();
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="tool-section fs-operations">
      <div className="section-title">
        <Wrench size={17} />
        <h2>Operaciones EXT2</h2>
      </div>
      <div className="operation-tabs" role="tablist" aria-label="Operacion EXT2">
        {operations.map((item) => {
          const Icon = item.icon;
          return (
            <button
              key={item.id}
              className={operation === item.id ? "active" : ""}
              onClick={() => setOperation(item.id)}
              type="button"
              role="tab"
              aria-selected={operation === item.id}
              title={item.label}
            >
              <Icon size={14} />
              <span>{item.label}</span>
            </button>
          );
        })}
      </div>
      <form className="compact-form operation-form" onSubmit={submit}>
        <label>
          Ruta EXT2
          <input
            value={path}
            onChange={(event) => setPath(event.target.value)}
            placeholder="/home/docs/a.txt"
            required
          />
        </label>

        {operation === "edit" && (
          <label>
            Archivo de contenido
            <input
              value={contentPath}
              onChange={(event) => setContentPath(event.target.value)}
              placeholder="/tmp/nuevo_contenido.txt"
              required
            />
            <small>Ruta accesible para el proceso Go del backend.</small>
          </label>
        )}

        {operation === "rename" && (
          <label>
            Nuevo nombre
            <input
              value={newName}
              onChange={(event) => setNewName(event.target.value)}
              placeholder="b1.txt"
              required
            />
          </label>
        )}

        {(operation === "copy" || operation === "move") && (
          <label>
            Carpeta destino
            <div className="input-with-icon">
              <FolderInput size={15} />
              <input
                value={destination}
                onChange={(event) => setDestination(event.target.value)}
                placeholder="/home/images"
                required
              />
            </div>
          </label>
        )}

        <button
          className={operation === "remove" ? "danger-button" : "primary-button"}
          disabled={busy || !enabled}
        >
          {operation === "remove" ? <Trash2 size={16} /> : <MoveRight size={16} />}
          {busy ? "Procesando..." : operations.find((item) => item.id === operation)?.label}
        </button>
        {!enabled && (
          <p className="form-note">Inicia sesion para habilitar operaciones.</p>
        )}
      </form>
    </section>
  );
}
