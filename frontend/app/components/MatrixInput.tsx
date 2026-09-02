type MatrixInputProps = {
  matrix: string[][];
  onCellChange: (row: number, column: number, value: string) => void;
  onAddRow: () => void;
  onRemoveRow: () => void;
  onAddColumn: () => void;
  onRemoveColumn: () => void;
  disabled: boolean;
};

export function MatrixInput({
  matrix,
  onCellChange,
  onAddRow,
  onRemoveRow,
  onAddColumn,
  onRemoveColumn,
  disabled
}: MatrixInputProps) {
  return (
    <section aria-labelledby="matrix-input-title" className="matrix-section">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Matriz de entrada</p>
          <h2 id="matrix-input-title">Ingresa los valores</h2>
        </div>
        <span className="matrix-size">
          {matrix.length} × {matrix[0].length}
        </span>
      </div>

      <div className="matrix-scroll">
        <div className="matrix-editor" role="grid" aria-label="Matriz editable">
          {matrix.map((row, rowIndex) => (
            <div className="matrix-row" role="row" key={`row-${rowIndex}`}>
              {row.map((value, columnIndex) => (
                <input
                  aria-label={`Fila ${rowIndex + 1}, columna ${columnIndex + 1}`}
                  className="matrix-cell"
                  disabled={disabled}
                  inputMode="decimal"
                  key={`cell-${rowIndex}-${columnIndex}`}
                  onChange={(event) => onCellChange(rowIndex, columnIndex, event.target.value)}
                  step="any"
                  type="number"
                  value={value}
                />
              ))}
            </div>
          ))}
        </div>
      </div>

      <div className="matrix-actions" aria-label="Controles de matriz">
        <button disabled={disabled} onClick={onAddRow} type="button">
          + Fila
        </button>
        <button disabled={disabled || matrix.length === 1} onClick={onRemoveRow} type="button">
          − Fila
        </button>
        <button disabled={disabled} onClick={onAddColumn} type="button">
          + Columna
        </button>
        <button disabled={disabled || matrix[0].length === 1} onClick={onRemoveColumn} type="button">
          − Columna
        </button>
      </div>
    </section>
  );
}
