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

  return (
    <main className="page-shell">
      <header className="page-header">
        <p className="eyebrow">Interseguro Coding Challenge</p>
        <h1>Factorización QR</h1>
        <p className="intro">Ingresa una matriz rectangular para calcular su factorización QR y las estadísticas de las matrices resultantes.</p>
      </header>

      <section className="workspace">
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
          {isLoading ? "Calculando..." : "Calcular QR"}
        </button>
        {requestState === "error" && error ? <p className="error-message" role="alert">{error}</p> : null}
      </section>

      {requestState === "success" && result ? (
        <section className="results" aria-labelledby="results-title">
          <div className="results-heading">
            <p className="eyebrow">Resultado</p>
            <h2 id="results-title">Factorización y estadísticas</h2>
          </div>
          <div className="matrices-grid">
            <MatrixTable formatNumber={formatNumber} matrix={result.q} title="Matriz Q" />
            <MatrixTable formatNumber={formatNumber} matrix={result.r} title="Matriz R" />
          </div>
          <Statistics formatNumber={formatNumber} statistics={result.statistics} />
        </section>
      ) : null}
    </main>
  );
}
