import { useState } from "react";
import { BarChart3, ExternalLink, FileImage, LoaderCircle } from "lucide-react";
import { api, ReportResult } from "../api/client";

const reportNames = ["disk", "tree", "inode", "block", "bm_inode", "bm_block"];

interface Props {
  activeId: string;
  currentPath: string;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function ReportsPanel({ activeId, currentPath, onMessage }: Props) {
  const [format, setFormat] = useState("svg");
  const [result, setResult] = useState<ReportResult | null>(null);
  const [busyName, setBusyName] = useState("");

  async function generate(name: string) {
    if (!activeId) {
      onMessage("Selecciona un ID montado", "error");
      return;
    }
    setBusyName(name);
    try {
      const response = await api.report({
        id: activeId,
        name,
        pathFileLs: name === "ls" || name === "file" ? currentPath : undefined,
        format,
      });
      setResult(response.data ?? null);
      onMessage(response.message || "Reporte generado", "success");
    } catch (error) {
      onMessage((error as Error).message, "error");
    } finally {
      setBusyName("");
    }
  }

  return (
    <section className="tool-section report-section">
      <div className="section-title">
        <BarChart3 size={17} />
        <h2>Reportes</h2>
        <select
          className="format-select"
          value={format}
          onChange={(event) => setFormat(event.target.value)}
          aria-label="Formato del reporte"
        >
          <option value="svg">SVG</option>
          <option value="png">PNG</option>
        </select>
      </div>
      <div className="report-grid">
        {reportNames.map((name) => (
          <button
            key={name}
            className="report-button"
            disabled={!activeId || Boolean(busyName)}
            onClick={() => void generate(name)}
          >
            {busyName === name ? (
              <LoaderCircle className="spin" size={16} />
            ) : (
              <FileImage size={16} />
            )}
            {name}
          </button>
        ))}
      </div>
      {result && (
        <div className="report-result">
          <div>
            <strong>{result.name}</strong>
            <span>{result.contentType}</span>
          </div>
          <ExternalLink size={16} />
          <code>{result.path}</code>
          <p>
            La API devuelve una ruta local. La vista embebida se habilitara cuando
            el backend exponga reportes como archivos estaticos.
          </p>
        </div>
      )}
    </section>
  );
}
