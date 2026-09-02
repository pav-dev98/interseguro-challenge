export const DIAGONAL_TOLERANCE = 1e-12;

export type Matrix = number[][];

export type Statistics = {
  max: number;
  min: number;
  average: number;
  sum: number;
  hasDiagonalMatrix: boolean;
};

export function calculateStatistics(q: unknown, r: unknown): Statistics {
  validateMatrix(q, "q");
  validateMatrix(r, "r");

  let max = Number.NEGATIVE_INFINITY;
  let min = Number.POSITIVE_INFINITY;
  let sum = 0;
  let count = 0;

  for (const matrix of [q, r]) {
    for (const row of matrix) {
      for (const value of row) {
        max = Math.max(max, value);
        min = Math.min(min, value);
        sum += value;
        count += 1;
      }
    }
  }

  return {
    max,
    min,
    sum,
    average: sum / count,
    hasDiagonalMatrix: isDiagonal(q) || isDiagonal(r)
  };
}

export function isDiagonal(matrix: Matrix): boolean {
  if (matrix.length !== matrix[0].length) {
    return false;
  }

  for (let row = 0; row < matrix.length; row += 1) {
    for (let column = 0; column < matrix.length; column += 1) {
      if (row !== column && Math.abs(matrix[row][column]) > DIAGONAL_TOLERANCE) {
        return false;
      }
    }
  }

  return true;
}

function validateMatrix(value: unknown, name: string): asserts value is Matrix {
  if (!Array.isArray(value) || value.length === 0) {
    throw new Error(`${name} must be a non-empty matrix`);
  }

  let columns: number | undefined;
  for (const row of value) {
    if (!Array.isArray(row) || row.length === 0) {
      throw new Error(`${name} must not contain empty rows`);
    }
    if (columns === undefined) {
      columns = row.length;
    } else if (row.length !== columns) {
      throw new Error(`${name} rows must have the same length`);
    }
    for (const cell of row) {
      if (typeof cell !== "number" || !Number.isFinite(cell)) {
        throw new Error(`${name} must contain only finite numbers`);
      }
    }
  }
}
