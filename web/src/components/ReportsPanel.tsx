import { useEffect, useState } from "react";
import {
  BarChart3,
  ExternalLink,
  FileImage,
  FileText,
  LoaderCircle,
} from "lucide-react";
import { API_BASE_URL, api, ReportResult } from "../api/client";

const reportNames = ["disk", "tree", "inode", "block", "bm_inode", "bm_block"];

interface Props {
  activeId: string;
  currentPath: string;
  refreshKey: number;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function ReportsPanel({
  activeId,
  currentPath,
  refreshKey,
  onMessage,
}: Props) {
  const [format, setFormat] = useState("svg");
  const [result, setResult] = useState<ReportResult | null>(null);
  const [busyName, setBusyName] = useState("");
  const reportURL = result
    ? result.url.startsWith("http")
      ? result.url
      : `${API_BASE_URL}${result.url}`
    : "";
  const isImage = result?.contentType.startsWith("image/") ?? false;
  const isPDF = result?.contentType === "application/pdf";
  const isText = result?.contentType.startsWith("text/") ?? false;

  useEffect(() => {
    if (refreshKey > 0) setResult(null);
  }, [refreshKey]);

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
          <option value="pdf">PDF</option>
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
          <div className="report-meta">
            <strong>{result.name}</strong>
            <span>{result.contentType}</span>
          </div>
          <a
            className="icon-button"
            href={reportURL}
            target="_blank"
            rel="noreferrer"
            title="Abrir reporte"
          >
            <ExternalLink size={16} />
          </a>
          {isImage && (
            <a
              className="report-preview"
              href={reportURL}
              target="_blank"
              rel="noreferrer"
              title="Abrir reporte en una pestana nueva"
            >
              <img src={reportURL} alt={`Reporte ${result.name}`} />
            </a>
          )}
          {(isPDF || isText) && (
            <a
              className="report-file-link"
              href={reportURL}
              target="_blank"
              rel="noreferrer"
            >
              {isText ? <FileText size={16} /> : <FileImage size={16} />}
              Abrir {isPDF ? "PDF" : "reporte de texto"}
            </a>
          )}
          <span className="technical-label">Ruta local</span>
          <code>{result.path}</code>
        </div>
      )}
    </section>
  );
}
