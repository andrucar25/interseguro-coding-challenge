export type Matrix = readonly (readonly number[])[];
export type Matrices = readonly Matrix[];

export interface MatrixStatistics {
	maximum: number;
	minimum: number;
	sum: number;
	average: number;
	hasDiagonalMatrix: boolean;
}

// QR output can retain floating-point round-off in values that are mathematically zero.
const DIAGONAL_RELATIVE_TOLERANCE = 1e-12;

export function calculateStatistics(input: unknown): MatrixStatistics {
	if (!Array.isArray(input) || input.length === 0) {
		throw new TypeError("Input must be a non-empty array of matrices.");
	}

	const matrices: readonly unknown[] = input;
	let maximum = -Infinity;
	let minimum = Infinity;
	let sum = 0;
	let count = 0;
	let hasDiagonalMatrix = false;

	for (const matrixInput of matrices) {
		if (!Array.isArray(matrixInput) || matrixInput.length === 0) {
			throw new TypeError("Each matrix must be a non-empty array of rows.");
		}

		const matrix: readonly unknown[] = matrixInput;
		let columnCount: number | undefined;
		let largestAbsoluteValue = 0;
		let largestAbsoluteOffDiagonalValue = 0;

		for (let rowIndex = 0; rowIndex < matrix.length; rowIndex += 1) {
			const rowInput = matrix[rowIndex];
			if (!Array.isArray(rowInput) || rowInput.length === 0) {
				throw new TypeError(
					"Each matrix row must be a non-empty array of numbers.",
				);
			}

			const row: readonly unknown[] = rowInput;
			if (columnCount === undefined) {
				columnCount = row.length;
			} else if (row.length !== columnCount) {
				throw new TypeError("Each matrix must have rectangular rows.");
			}

			for (let columnIndex = 0; columnIndex < row.length; columnIndex += 1) {
				const value = row[columnIndex];
				if (typeof value !== "number" || !Number.isFinite(value)) {
					throw new TypeError("Each matrix value must be a finite number.");
				}

				maximum = Math.max(maximum, value);
				minimum = Math.min(minimum, value);
				sum += value;
				count += 1;

				const absoluteValue = Math.abs(value);
				largestAbsoluteValue = Math.max(largestAbsoluteValue, absoluteValue);
				if (rowIndex !== columnIndex) {
					largestAbsoluteOffDiagonalValue = Math.max(
						largestAbsoluteOffDiagonalValue,
						absoluteValue,
					);
				}
			}
		}

		const isSquare = matrix.length === columnCount;
		const diagonalThreshold =
			DIAGONAL_RELATIVE_TOLERANCE * Math.max(1, largestAbsoluteValue);
		if (isSquare && largestAbsoluteOffDiagonalValue <= diagonalThreshold) {
			hasDiagonalMatrix = true;
		}
	}

	return {
		maximum,
		minimum,
		sum,
		average: sum / count,
		hasDiagonalMatrix,
	};
}
