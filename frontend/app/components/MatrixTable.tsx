type MatrixTableProps = {
  matrix: number[][];
  title: string;
  formatNumber: (value: number) => string;
};

export function MatrixTable({ matrix, title, formatNumber }: MatrixTableProps) {
  return (
    <section aria-labelledby={`${title.toLowerCase()}-title`} className="result-panel">
      <h3 id={`${title.toLowerCase()}-title`}>{title}</h3>
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
