package qr

import (
	"errors"
	"math"
	"testing"
)

const comparisonTolerance = 1e-10

func TestFactorizeKnownTwoByTwoMatrix(t *testing.T) {
	matrix := [][]float64{{1, 1}, {1, 0}}
	q, r, err := Factorize(matrix)
	if err != nil {
		t.Fatalf("Factorize returned an error: %v", err)
	}

	expectedQ := [][]float64{{math.Sqrt(0.5), math.Sqrt(0.5)}, {math.Sqrt(0.5), -math.Sqrt(0.5)}}
	expectedR := [][]float64{{math.Sqrt(2), math.Sqrt(0.5)}, {0, math.Sqrt(0.5)}}
	assertMatrixApproxEqual(t, q, expectedQ)
	assertMatrixApproxEqual(t, r, expectedR)
	assertQRProperties(t, matrix, q, r)
}

func TestFactorizeRectangularMatrix(t *testing.T) {
	matrix := [][]float64{{1, 1}, {1, 0}, {0, 1}}
	q, r, err := Factorize(matrix)
	if err != nil {
		t.Fatalf("Factorize returned an error: %v", err)
	}

	if len(q) != 3 || len(q[0]) != 2 || len(r) != 2 || len(r[0]) != 2 {
		t.Fatalf("unexpected dimensions: Q is %dx%d, R is %dx%d", len(q), len(q[0]), len(r), len(r[0]))
	}
	assertQRProperties(t, matrix, q, r)
}

func TestFactorizeRejectsInvalidMatrices(t *testing.T) {
	tests := []struct {
		name   string
		matrix [][]float64
		want   error
	}{
		{name: "empty matrix", matrix: [][]float64{}, want: ErrEmptyMatrix},
		{name: "empty row", matrix: [][]float64{{}}, want: ErrEmptyRow},
		{name: "irregular rows", matrix: [][]float64{{1, 2}, {3}}, want: ErrIrregularMatrix},
		{name: "more columns than rows", matrix: [][]float64{{1, 2, 3}, {4, 5, 6}}, want: ErrWideMatrix},
		{name: "non-finite value", matrix: [][]float64{{math.NaN()}}, want: ErrInvalidValue},
		{name: "dependent columns", matrix: [][]float64{{1, 2}, {2, 4}, {3, 6}}, want: ErrDependentColumns},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := Factorize(test.matrix)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func assertQRProperties(t *testing.T, matrix, q, r [][]float64) {
	t.Helper()
	assertMatrixApproxEqual(t, multiply(q, r), matrix)

	for column := range q[0] {
		for other := range q[0] {
			dotProduct := 0.0
			for row := range q {
				dotProduct += q[row][column] * q[row][other]
			}
			expected := 0.0
			if column == other {
				expected = 1
			}
			if math.Abs(dotProduct-expected) > comparisonTolerance {
				t.Fatalf("QᵀQ[%d][%d] = %v, expected %v", column, other, dotProduct, expected)
			}
		}
	}

	for row := range r {
		for column := 0; column < row; column++ {
			if math.Abs(r[row][column]) > comparisonTolerance {
				t.Fatalf("R[%d][%d] = %v, expected 0", row, column, r[row][column])
			}
		}
	}
}

func multiply(left, right [][]float64) [][]float64 {
	result := make([][]float64, len(left))
	for row := range left {
		result[row] = make([]float64, len(right[0]))
		for column := range right[0] {
			for index := range right {
				result[row][column] += left[row][index] * right[index][column]
			}
		}
	}
	return result
}

func assertMatrixApproxEqual(t *testing.T, actual, expected [][]float64) {
	t.Helper()
	for row := range expected {
		for column := range expected[row] {
			if math.Abs(actual[row][column]-expected[row][column]) > comparisonTolerance {
				t.Fatalf("matrix[%d][%d] = %v, expected %v", row, column, actual[row][column], expected[row][column])
			}
		}
	}
}
