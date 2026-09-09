package qr

import (
	"errors"
	"math"
	"testing"
)

const tolerance = 1e-10

func TestFactorizeProperties(t *testing.T) {
	testCases := []struct {
		name  string
		input Matrix
	}{
		{
			name: "square",
			input: Matrix{
				{12, -51, 4},
				{6, 167, -68},
				{-4, 24, -41},
			},
		},
		{
			name: "tall with negative decimal values",
			input: Matrix{
				{1.5, -2, 3},
				{4, -5.25, 6.5},
				{-7, 8, 9.75},
				{2.25, -3.5, 4.125},
			},
		},
		{
			name: "single column",
			input: Matrix{
				{-3},
				{4},
				{12},
			},
		},
		{
			name: "rank deficient",
			input: Matrix{
				{1, 2},
				{2, 4},
				{3, 6},
			},
		},
		{
			name: "all zero",
			input: Matrix{
				{0, 0},
				{0, 0},
				{0, 0},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			q, r, err := Factorize(testCase.input)
			if err != nil {
				t.Fatalf("Factorize() error = %v", err)
			}

			assertShape(t, q, len(testCase.input), len(testCase.input[0]))
			assertShape(t, r, len(testCase.input[0]), len(testCase.input[0]))
			assertMatrixApproxEqual(t, multiply(transpose(q), q), identity(len(testCase.input[0])))
			assertMatrixApproxEqual(t, multiply(q, r), testCase.input)
			assertUpperTriangular(t, r)
		})
	}
}

func TestFactorizeDoesNotMutateInput(t *testing.T) {
	input := Matrix{
		{1, 2},
		{3, 4},
		{5, 6},
	}
	want := copyMatrix(input)

	if _, _, err := Factorize(input); err != nil {
		t.Fatalf("Factorize() error = %v", err)
	}

	if !matricesExactlyEqual(input, want) {
		t.Fatalf("Factorize() mutated input: got %v, want %v", input, want)
	}
}

func TestFactorizeInvalidInput(t *testing.T) {
	testCases := []struct {
		name  string
		input Matrix
		want  error
	}{
		{name: "nil", input: nil, want: ErrEmptyMatrix},
		{name: "empty", input: Matrix{}, want: ErrEmptyMatrix},
		{name: "empty first row", input: Matrix{{}}, want: ErrEmptyRow},
		{name: "empty later row", input: Matrix{{1}, {}}, want: ErrEmptyRow},
		{name: "ragged rows", input: Matrix{{1, 2}, {3}}, want: ErrNonRectangular},
		{name: "wide matrix", input: Matrix{{1, 2, 3}, {4, 5, 6}}, want: ErrWideMatrix},
		{name: "not a number", input: Matrix{{math.NaN()}}, want: ErrNonFinite},
		{name: "positive infinity", input: Matrix{{math.Inf(1)}}, want: ErrNonFinite},
		{name: "negative infinity", input: Matrix{{math.Inf(-1)}}, want: ErrNonFinite},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			q, r, err := Factorize(testCase.input)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("Factorize() error = %v, want errors.Is(..., %v)", err, testCase.want)
			}
			if q != nil || r != nil {
				t.Fatalf("Factorize() returned matrices on error: q = %v, r = %v", q, r)
			}
		})
	}
}

func assertShape(t *testing.T, matrix Matrix, rows, columns int) {
	t.Helper()
	if len(matrix) != rows {
		t.Fatalf("matrix rows = %d, want %d", len(matrix), rows)
	}
	for _, row := range matrix {
		if len(row) != columns {
			t.Fatalf("matrix columns = %d, want %d", len(row), columns)
		}
	}
}

func assertMatrixApproxEqual(t *testing.T, got, want Matrix) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("matrix row count = %d, want %d", len(got), len(want))
	}
	for row := range got {
		if len(got[row]) != len(want[row]) {
			t.Fatalf("matrix column count at row %d = %d, want %d", row, len(got[row]), len(want[row]))
		}
		for column := range got[row] {
			if !approximatelyEqual(got[row][column], want[row][column]) {
				t.Fatalf("matrix[%d][%d] = %.16g, want %.16g", row, column, got[row][column], want[row][column])
			}
		}
	}
}

func assertUpperTriangular(t *testing.T, matrix Matrix) {
	t.Helper()
	for row := range matrix {
		for column := 0; column < row; column++ {
			if !approximatelyEqual(matrix[row][column], 0) {
				t.Fatalf("matrix[%d][%d] = %.16g, want approximately zero", row, column, matrix[row][column])
			}
		}
	}
}

func approximatelyEqual(got, want float64) bool {
	return math.Abs(got-want) <= tolerance*math.Max(1, math.Max(math.Abs(got), math.Abs(want)))
}

func transpose(matrix Matrix) Matrix {
	result := make(Matrix, len(matrix[0]))
	for column := range result {
		result[column] = make([]float64, len(matrix))
		for row := range matrix {
			result[column][row] = matrix[row][column]
		}
	}
	return result
}

func multiply(left, right Matrix) Matrix {
	result := make(Matrix, len(left))
	for row := range result {
		result[row] = make([]float64, len(right[0]))
		for column := range result[row] {
			for index := range right {
				result[row][column] += left[row][index] * right[index][column]
			}
		}
	}
	return result
}

func identity(size int) Matrix {
	result := make(Matrix, size)
	for index := range result {
		result[index] = make([]float64, size)
		result[index][index] = 1
	}
	return result
}

func matricesExactlyEqual(left, right Matrix) bool {
	if len(left) != len(right) {
		return false
	}
	for row := range left {
		if len(left[row]) != len(right[row]) {
			return false
		}
		for column := range left[row] {
			if left[row][column] != right[row][column] {
				return false
			}
		}
	}
	return true
}
