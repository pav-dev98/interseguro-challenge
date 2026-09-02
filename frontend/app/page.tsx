"use client";

import { useState } from "react";
import { MatrixInput } from "./components/MatrixInput";
import { MatrixTable } from "./components/MatrixTable";
import { Statistics } from "./components/Statistics";
import type { ErrorResponse, Matrix, QRResponse, Statistics as StatisticsData } from "./types";

const initialMatrix = [
  ["1", "1"],
  ["1", "0"]
];

type RequestState = "idle" | "loading" | "success" | "error";

function formatNumber(value: number): string {
  return new Intl.NumberFormat("es-PE", { maximumFractionDigits: 6 }).format(value);
}

function isMatrix(value: unknown): value is Matrix {
  return Array.isArray(value) && value.every((row) => Array.isArray(row) && row.every((cell) => typeof cell === "number"));
}

function isStatistics(value: unknown): value is StatisticsData {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return (
    typeof candidate.max === "number" &&
    typeof candidate.min === "number" &&
    typeof candidate.sum === "number" &&
    typeof candidate.average === "number" &&
    typeof candidate.hasDiagonalMatrix === "boolean"
  );
}

function isQRResponse(value: unknown): value is QRResponse {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  return isMatrix(candidate.q) && isMatrix(candidate.r) && isStatistics(candidate.statistics);
}

function errorFromResponse(value: unknown): string | null {
  if (typeof value !== "object" || value === null) {
    return null;
  }

  const candidate = value as Partial<ErrorResponse>;
  return typeof candidate.error === "string" ? candidate.error : null;
}

export default function Home() {
  const [matrix, setMatrix] = useState<string[][]>(initialMatrix);
  const [requestState, setRequestState] = useState<RequestState>("idle");
  const [result, setResult] = useState<QRResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const updateCell = (row: number, column: number, value: string) => {
    setMatrix((current) => current.map((currentRow, rowIndex) => currentRow.map((cell, columnIndex) => (rowIndex === row && columnIndex === column ? value : cell))));
  };

  const addRow = () => {
    setMatrix((current) => [...current, Array.from({ length: current[0].length }, () => "0")]);
  };

  const removeRow = () => {
    setMatrix((current) => current.slice(0, -1));
  };

  const addColumn = () => {
    setMatrix((current) => current.map((row) => [...row, "0"]));
  };

  const removeColumn = () => {
    setMatrix((current) => current.map((row) => row.slice(0, -1)));
  };

  const calculateQR = async () => {
    const apiURL = process.env.NEXT_PUBLIC_GO_API_URL;
    if (!apiURL) {
      setRequestState("error");
      setError("Configura NEXT_PUBLIC_GO_API_URL para conectar con la Go API.");
      return;
    }

    const numericMatrix = matrix.map((row) => row.map((value) => Number(value)));
    if (numericMatrix.some((row) => row.some((value) => !Number.isFinite(value)))) {
      setRequestState("error");
      setError("Cada celda debe contener un número válido.");
      return;
    }

    setRequestState("loading");
    setError(null);

    try {
      const response = await fetch(`${apiURL.replace(/\/$/, "")}/qr`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ matrix: numericMatrix })
      });
      const payload: unknown = await response.json().catch(() => null);

      if (!response.ok) {
        setRequestState("error");
        setError(errorFromResponse(payload) ?? "No fue posible calcular la factorización QR.");
        return;
      }
      if (!isQRResponse(payload)) {
        setRequestState("error");
        setError("La Go API devolvió una respuesta inesperada.");
        return;
      }

      setResult(payload);
      setRequestState("success");
    } catch {
      setRequestState("error");
      setError("No se pudo conectar con la Go API. Verifica la URL y que el servicio esté disponible.");
    }
  };

  const isLoading = requestState === "loading";
  const runtimeStatus = requestState === "loading" ? "CALCULANDO" : requestState === "success" ? "COMPLETADO" : requestState === "error" ? "ERROR" : "LISTO";

  return (
    <main className="application-shell">
      <aside className="sidebar" aria-label="Navegación principal">
        <div className="brand"><span aria-hidden="true" className="brand-mark">⌁</span><span>QR_SOLVER</span></div>
        <nav className="sidebar-nav">
          <p>Compute</p>
          <span className="nav-item nav-item-active"><span aria-hidden="true">Σ</span>Solver</span>
          <span className="nav-item nav-item-muted"><span aria-hidden="true">▤</span>Theory</span>
          <p>Resources</p>
          <span className="nav-item nav-item-muted"><span aria-hidden="true">◷</span>History</span>
          <span className="nav-item nav-item-muted"><span aria-hidden="true">›_</span>API Docs</span>
        </nav>
      </aside>

      <div className="application-content">
        <header className="topbar">
          <div className="topbar-status"><span className={`status-chip status-${requestState}`}>{runtimeStatus}</span><span className="topbar-divider" /><span>QR_SOLVER</span></div>
          <div className="topbar-tools" aria-label="Herramientas visuales"><span aria-hidden="true">⌕</span><span aria-hidden="true">⚙</span><span aria-hidden="true" className="user-glyph">◉</span></div>
        </header>

        <div className="dashboard">
          <section className="workspace input-blade">
            <MatrixInput
              disabled={isLoading}
              matrix={matrix}
              onAddColumn={addColumn}
              onAddRow={addRow}
              onCellChange={updateCell}
              onRemoveColumn={removeColumn}
              onRemoveRow={removeRow}
            />
            <button className="submit-button" disabled={isLoading} onClick={calculateQR} type="button">
              <span>{isLoading ? "Calculando..." : "Ejecutar factorización"}</span><span aria-hidden="true">→</span>
            </button>
            {requestState === "error" && error ? <p className="error-message" role="alert">{error}</p> : null}
          </section>

          {requestState === "success" && result ? (
            <section className="results result-blade" aria-labelledby="results-title">
              <div className="results-heading">
                <h2 id="results-title">Resultado de transformación</h2>
                <p>Descomposición A = QR mediante Modified Gram-Schmidt</p>
              </div>
              <div className="matrices-grid">
                <MatrixTable formatNumber={formatNumber} matrix={result.q} title="Matriz Q" />
                <MatrixTable formatNumber={formatNumber} matrix={result.r} title="Matriz R" />
              </div>
              <Statistics formatNumber={formatNumber} statistics={result.statistics} />
            </section>
          ) : (
            <section className="results result-blade empty-results" aria-labelledby="results-title">
              <div className="results-heading">
                <h2 id="results-title">Resultado de transformación</h2>
                <p>Descomposición A = QR mediante Modified Gram-Schmidt</p>
              </div>
              <div className="empty-state"><span aria-hidden="true">∎</span><p>Ejecuta la factorización para visualizar las matrices Q, R y sus estadísticas.</p></div>
            </section>
          )}
        </div>
        <footer className="statusbar"><div><span className="engine-dot" />ENGINE: READY <span className="statusbar-divider" /> PRECISION: FLOAT64</div><span>UTF-8</span></footer>
      </div>
    </main>
  );
}
