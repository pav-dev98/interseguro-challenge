package qr

import (
	"errors"
	"math"
)

const Tolerance = 1e-12

var (
	ErrEmptyMatrix      = errors.New("matrix must not be empty")
	ErrEmptyRow         = errors.New("matrix rows must not be empty")
	ErrIrregularMatrix  = errors.New("all matrix rows must have the same length")
	ErrWideMatrix       = errors.New("matrix must have at least as many rows as columns (m >= n)")
	ErrInvalidValue     = errors.New("matrix values must be finite numbers")
	ErrDependentColumns = errors.New("matrix columns are linearly dependent or numerically near dependent")
)

// Factorize calculates the reduced QR factorization using Modified Gram-Schmidt.
func Factorize(matrix [][]float64) ([][]float64, [][]float64, error) {
	m, n, err := validate(matrix)
	if err != nil {
		return nil, nil, err
	}

	q := makeMatrix(m, n)
	r := makeMatrix(n, n)

	for column := 0; column < n; column++ {
		residual := make([]float64, m)
		for row := 0; row < m; row++ {
			residual[row] = matrix[row][column]
		}

		originalNorm := norm(residual)
		for previous := 0; previous < column; previous++ {
			r[previous][column] = dotColumn(q, previous, residual)
			for row := 0; row < m; row++ {
				residual[row] -= r[previous][column] * q[row][previous]
			}
		}

		r[column][column] = norm(residual)
		if r[column][column] <= Tolerance*math.Max(1, originalNorm) {
			return nil, nil, ErrDependentColumns
		}

		for row := 0; row < m; row++ {
			q[row][column] = residual[row] / r[column][column]
		}
	}

	return q, r, nil
}

func validate(matrix [][]float64) (int, int, error) {
	if len(matrix) == 0 {
		return 0, 0, ErrEmptyMatrix
	}

	columns := len(matrix[0])
	if columns == 0 {
		return 0, 0, ErrEmptyRow
	}
	if len(matrix) < columns {
		return 0, 0, ErrWideMatrix
	}

	for _, row := range matrix {
		if len(row) == 0 {
			return 0, 0, ErrEmptyRow
		}
		if len(row) != columns {
			return 0, 0, ErrIrregularMatrix
		}
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return 0, 0, ErrInvalidValue
			}
		}
	}

	return len(matrix), columns, nil
}

func makeMatrix(rows, columns int) [][]float64 {
	matrix := make([][]float64, rows)
	for row := range matrix {
		matrix[row] = make([]float64, columns)
	}
	return matrix
}

func dotColumn(matrix [][]float64, column int, vector []float64) float64 {
	result := 0.0
	for row := range matrix {
		result += matrix[row][column] * vector[row]
	}
	return result
}

func norm(vector []float64) float64 {
	sum := 0.0
	for _, value := range vector {
		sum += value * value
	}
	return math.Sqrt(sum)
}
