type MatrixTableProps = {
  matrix: number[][];
  title: string;
  formatNumber: (value: number) => string;
};

export function MatrixTable({ matrix, title, formatNumber }: MatrixTableProps) {
  const kind = title === "Matriz Q" ? "q" : "r";

  return (
    <section aria-labelledby={`${kind}-matrix-title`} className={`result-panel result-panel-${kind}`}>
      <h3 id={`${kind}-matrix-title`}><span className="matrix-indicator" />{title}{kind === "q" ? " (Ortogonal)" : " (Triangular sup.)"}</h3>
      <div className="matrix-scroll">
        <table className="result-matrix">
          <tbody>
            {matrix.map((row, rowIndex) => (
              <tr key={`${title}-row-${rowIndex}`}>
                {row.map((value, columnIndex) => (
                  <td key={`${title}-cell-${rowIndex}-${columnIndex}`}>{formatNumber(value)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
