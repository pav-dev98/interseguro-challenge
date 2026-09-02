export type Matrix = number[][];

export type Statistics = {
  max: number;
  min: number;
  sum: number;
  average: number;
  hasDiagonalMatrix: boolean;
};

export type QRResponse = {
  q: Matrix;
  r: Matrix;
  statistics: Statistics;
};

export type ErrorResponse = {
  error: string;
};
