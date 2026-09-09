export type Matrix = readonly (readonly number[])[];
export type Matrices = readonly Matrix[];

export interface MatrixStatistics {
	maximum: number;
	minimum: number;
	sum: number;
	average: number;
	hasDiagonalMatrix: boolean;
}

const DIAGONAL_RELATIVE_TOLERANCE = 1e-12;

interface AggregateAccumulator {
	maximum: number;
	minimum: number;
	sum: number;
	count: number;
}

function validateMatrixCollection(input: unknown): readonly unknown[] {
	if (!Array.isArray(input) || input.length === 0) {
		throw new TypeError("Input must be a non-empty array of matrices.");
	}

	return input;
}

function createAggregateAccumulator(): AggregateAccumulator {
	return {
		maximum: -Infinity,
		minimum: Infinity,
		sum: 0,
		count: 0,
	};
}

function updateAggregates(
	accumulator: AggregateAccumulator,
	value: number,
): void {
	accumulator.maximum = Math.max(accumulator.maximum, value);
	accumulator.minimum = Math.min(accumulator.minimum, value);
	accumulator.sum += value;
	accumulator.count += 1;
}

function isDiagonalMatrix(
	rowCount: number,
	columnCount: number,
	largestAbsoluteValue: number,
	largestAbsoluteOffDiagonalValue: number,
): boolean {
	const diagonalThreshold =
		DIAGONAL_RELATIVE_TOLERANCE * Math.max(1, largestAbsoluteValue);

	return (
		rowCount === columnCount &&
		largestAbsoluteOffDiagonalValue <= diagonalThreshold
	);
}

function processMatrix(
	matrixInput: unknown,
	accumulator: AggregateAccumulator,
): boolean {
	if (!Array.isArray(matrixInput) || matrixInput.length === 0) {
		throw new TypeError("Each matrix must be a non-empty array of rows.");
	}

	const matrix: readonly unknown[] = matrixInput;
	let columnCount = 0;
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
		if (rowIndex === 0) {
			columnCount = row.length;
		} else if (row.length !== columnCount) {
			throw new TypeError("Each matrix must have rectangular rows.");
		}

		for (let columnIndex = 0; columnIndex < row.length; columnIndex += 1) {
			const value = row[columnIndex];
			if (typeof value !== "number" || !Number.isFinite(value)) {
				throw new TypeError("Each matrix value must be a finite number.");
			}

			updateAggregates(accumulator, value);

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

	return isDiagonalMatrix(
		matrix.length,
		columnCount,
		largestAbsoluteValue,
		largestAbsoluteOffDiagonalValue,
	);
}

export const calculateStatistics = (input: unknown): MatrixStatistics => {
	const matrices = validateMatrixCollection(input);
	const accumulator = createAggregateAccumulator();
	let hasDiagonalMatrix = false;

	for (const matrixInput of matrices) {
		const matrixIsDiagonal = processMatrix(matrixInput, accumulator);
		hasDiagonalMatrix = hasDiagonalMatrix || matrixIsDiagonal;
	}

	return {
		maximum: accumulator.maximum,
		minimum: accumulator.minimum,
		sum: accumulator.sum,
		average: accumulator.sum / accumulator.count,
		hasDiagonalMatrix,
	};
};
