import { FormEvent, KeyboardEvent, useState } from "react";
import {
  ChevronRight,
  CircleAlert,
  CircleCheck,
  Play,
  TerminalSquare,
  Trash2,
} from "lucide-react";
import { api, CommandExecution } from "../api/client";

const examples = [
  "mkdir -p -path=/home/docs",
  "mkfile -path=/home/docs/a.txt -size=20",
  "cat -file=/home/docs/a.txt",
  "rename -path=/home/docs/a.txt -name=b1.txt",
  "copy -path=/home/docs/b1.txt -destino=/home",
  "remove -path=/home/docs/b1.txt",
];

interface HistoryEntry {
  id: number;
  command: string;
  output: string;
  ok: boolean;
}

interface Props {
  onExecuted: (result: CommandExecution) => void;
  onMessage: (message: string, kind?: "success" | "error") => void;
}

export function CommandConsole({ onExecuted, onMessage }: Props) {
  const [command, setCommand] = useState("");
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [busy, setBusy] = useState(false);

  async function execute(event?: FormEvent) {
    event?.preventDefault();
    const input = command.trim();
    if (!input) {
      onMessage("Escribe un comando", "error");
      return;
    }
    setBusy(true);
    try {
      const response = await api.executeCommand(input);
      const result = response.data;
      if (!result) throw new Error("La API no devolvio el resultado del comando");
      setHistory((current) =>
        [
          {
            id: Date.now(),
            command: result.command,
            output: result.output || response.message || "Comando ejecutado",
            ok: true,
          },
          ...current,
        ].slice(0, 30),
      );
      setCommand("");
      onExecuted(result);
      onMessage(response.message || "Comando ejecutado", "success");
    } catch (error) {
      const message = (error as Error).message;
      setHistory((current) =>
        [
          { id: Date.now(), command: input, output: message, ok: false },
          ...current,
        ].slice(0, 30),
      );
      onMessage(message, "error");
    } finally {
      setBusy(false);
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === "Enter") {
      event.preventDefault();
      if (!busy) void execute();
    }
  }

  return (
    <section className="tool-section command-console">
      <div className="section-title">
        <TerminalSquare size={17} />
        <h2>Consola MIA</h2>
        <button
          className="icon-button"
          type="button"
          onClick={() => setHistory([])}
          disabled={history.length === 0}
          title="Limpiar historial"
        >
          <Trash2 size={15} />
        </button>
      </div>
      <form className="console-input" onSubmit={execute}>
        <ChevronRight size={16} />
        <textarea
          rows={2}
          value={command}
          onChange={(event) => setCommand(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="mkdir -p -path=/home/docs"
          spellCheck={false}
          aria-label="Comando MIA"
        />
        <button
          className="icon-button console-run"
          disabled={busy || !command.trim()}
          title="Ejecutar comando"
        >
          <Play size={16} />
        </button>
      </form>
      <div className="quick-commands" aria-label="Ejemplos de comandos">
        {examples.map((example) => (
          <button
            key={example}
            type="button"
            onClick={() => setCommand(example)}
            title={example}
          >
            {example.split(" ")[0]}
          </button>
        ))}
      </div>
      <div className="console-history" aria-live="polite">
        {history.map((entry) => (
          <article
            className={`console-entry ${entry.ok ? "success" : "error"}`}
            key={entry.id}
          >
            <div>
              {entry.ok ? <CircleCheck size={13} /> : <CircleAlert size={13} />}
              <code>$ {entry.command}</code>
            </div>
            <pre>{entry.output}</pre>
          </article>
        ))}
        {history.length === 0 && (
          <p className="empty-state">La salida de comandos aparecera aqui.</p>
        )}
      </div>
    </section>
  );
}
