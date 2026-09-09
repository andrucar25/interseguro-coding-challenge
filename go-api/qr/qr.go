// Package qr provides a pure, economy QR factorization for finite matrices.
package qr

import (
	"errors"
	"math"
)

// Matrix is a row-major matrix of float64 values.
type Matrix [][]float64

var (
	// ErrEmptyMatrix indicates that a matrix has no rows.
	ErrEmptyMatrix = errors.New("matrix must contain at least one row")
	// ErrEmptyRow indicates that a matrix contains a row with no columns.
	ErrEmptyRow = errors.New("matrix rows must contain at least one column")
	// ErrNonRectangular indicates that matrix rows do not all have the same length.
	ErrNonRectangular = errors.New("matrix rows must have equal lengths")
	// ErrWideMatrix indicates that a matrix has more columns than rows.
	ErrWideMatrix = errors.New("matrix must have at least as many rows as columns")
	// ErrNonFinite indicates that a matrix contains NaN or an infinite value.
	ErrNonFinite = errors.New("matrix values must be finite")
	// ErrNumericalFailure indicates that QR calculation produced a non-finite value.
	ErrNumericalFailure = errors.New("qr factorization encountered a numerical failure")
)

type reflector struct {
	vector []float64
	beta   float64
}

// Factorize returns the economy QR factorization of input using Householder
// reflections. For an m by n input, m >= n, Q is m by n with orthonormal
// columns and R is n by n upper triangular such that input is approximately
// Q multiplied by R.
func Factorize(input Matrix) (q Matrix, r Matrix, err error) {
	rows, columns, err := validate(input)
	if err != nil {
		return nil, nil, err
	}

	r = copyMatrix(input)
	reflectors := make([]reflector, columns)

	for column := range columns {
		norm := 0.0
		for row := column; row < rows; row++ {
			norm = math.Hypot(norm, r[row][column])
			if !isFinite(norm) {
				return nil, nil, ErrNumericalFailure
			}
		}

		if norm == 0 {
			continue
		}

		alpha := -norm
		if r[column][column] < 0 {
			alpha = norm
		}

		vector := make([]float64, rows-column)
		for row := column; row < rows; row++ {
			vector[row-column] = r[row][column] / norm
			if !isFinite(vector[row-column]) {
				return nil, nil, ErrNumericalFailure
			}
		}
		vector[0] -= alpha / norm
		if !isFinite(vector[0]) {
			return nil, nil, ErrNumericalFailure
		}

		vectorNormSquared := 0.0
		for _, value := range vector {
			vectorNormSquared += value * value
			if !isFinite(vectorNormSquared) {
				return nil, nil, ErrNumericalFailure
			}
		}
		if vectorNormSquared == 0 {
			return nil, nil, ErrNumericalFailure
		}
		beta := 2 / vectorNormSquared
		if !isFinite(beta) {
			return nil, nil, ErrNumericalFailure
		}

		if !applyReflector(r, column, column+1, columns, vector, beta) {
			return nil, nil, ErrNumericalFailure
		}
		r[column][column] = alpha
		for row := column + 1; row < rows; row++ {
			r[row][column] = 0
		}
		reflectors[column] = reflector{vector: vector, beta: beta}
	}

	q = identityColumns(rows, columns)
	for column := columns - 1; column >= 0; column-- {
		reflection := reflectors[column]
		if reflection.vector == nil {
			continue
		}
		if !applyReflector(q, column, 0, columns, reflection.vector, reflection.beta) {
			return nil, nil, ErrNumericalFailure
		}
	}

	resultR := make(Matrix, columns)
	for row := range resultR {
		resultR[row] = append([]float64(nil), r[row]...)
	}

	if !matrixIsFinite(q) || !matrixIsFinite(resultR) {
		return nil, nil, ErrNumericalFailure
	}
	return q, resultR, nil
}

func validate(input Matrix) (rows, columns int, err error) {
	if len(input) == 0 {
		return 0, 0, ErrEmptyMatrix
	}

	columns = len(input[0])
	if columns == 0 {
		return 0, 0, ErrEmptyRow
	}

	for _, row := range input {
		if len(row) == 0 {
			return 0, 0, ErrEmptyRow
		}
		if len(row) != columns {
			return 0, 0, ErrNonRectangular
		}
		for _, value := range row {
			if !isFinite(value) {
				return 0, 0, ErrNonFinite
			}
		}
	}

	rows = len(input)
	if rows < columns {
		return 0, 0, ErrWideMatrix
	}
	return rows, columns, nil
}

func copyMatrix(input Matrix) Matrix {
	copy := make(Matrix, len(input))
	for row := range input {
		copy[row] = append([]float64(nil), input[row]...)
	}
	return copy
}

func identityColumns(rows, columns int) Matrix {
	matrix := make(Matrix, rows)
	for row := range matrix {
		matrix[row] = make([]float64, columns)
		if row < columns {
			matrix[row][row] = 1
		}
	}
	return matrix
}

func applyReflector(matrix Matrix, rowStart, columnStart, columnEnd int, vector []float64, beta float64) bool {
	for column := columnStart; column < columnEnd; column++ {
		dot := 0.0
		for row, value := range vector {
			dot += value * matrix[rowStart+row][column]
			if !isFinite(dot) {
				return false
			}
		}

		scale := beta * dot
		if !isFinite(scale) {
			return false
		}
		for row, value := range vector {
			matrix[rowStart+row][column] -= scale * value
			if !isFinite(matrix[rowStart+row][column]) {
				return false
			}
		}
	}
	return true
}

func matrixIsFinite(matrix Matrix) bool {
	for _, row := range matrix {
		for _, value := range row {
			if !isFinite(value) {
				return false
			}
		}
	}
	return true
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
