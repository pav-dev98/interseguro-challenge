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
    <section aria-labelledby="matrix-input-title" className="matrix-section input-section">
      <div className="matrix-heading">
        <h2 id="matrix-input-title">Matriz de entrada</h2>
        <div className="dimension-controls" aria-label="Dimensiones de la matriz">
          <div className="dimension-control">
            <span>Filas</span>
            <div className="dimension-stepper">
              <button aria-label="Eliminar fila" disabled={disabled || matrix.length === 1} onClick={onRemoveRow} type="button">−</button>
              <strong>{matrix.length}</strong>
              <button aria-label="Agregar fila" disabled={disabled} onClick={onAddRow} type="button">+</button>
            </div>
          </div>
          <div className="dimension-control">
            <span>Columnas</span>
            <div className="dimension-stepper">
              <button aria-label="Eliminar columna" disabled={disabled || matrix[0].length === 1} onClick={onRemoveColumn} type="button">−</button>
              <strong>{matrix[0].length}</strong>
              <button aria-label="Agregar columna" disabled={disabled} onClick={onAddColumn} type="button">+</button>
            </div>
          </div>
        </div>
      </div>

      <div className="matrix-scroll input-matrix-scroll">
        <div className="matrix-editor" role="grid" aria-label="Matriz editable" style={{ gridTemplateColumns: `repeat(${matrix[0].length}, minmax(76px, 1fr))` }}>
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

    </section>
  );
}
