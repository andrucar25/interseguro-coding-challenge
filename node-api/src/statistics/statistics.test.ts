import { describe, expect, it } from "vitest";

import { calculateStatistics } from "./statistics.js";

describe("calculateStatistics", () => {
	it("calculates every aggregate across positive-valued matrices", () => {
		const result = calculateStatistics([
			[
				[1, 2],
				[3, 4],
			],
			[[5]],
		]);

		expect(result.maximum).toBe(5);
		expect(result.minimum).toBe(1);
		expect(result.sum).toBe(15);
		expect(result.average).toBe(3);
	});

	it("calculates every aggregate across negative-valued matrices", () => {
		const result = calculateStatistics([[[-1, -2]], [[-3]]]);

		expect(result.maximum).toBe(-1);
		expect(result.minimum).toBe(-3);
		expect(result.sum).toBe(-6);
		expect(result.average).toBe(-2);
	});

	it("calculates decimal-valued aggregates using the actual scalar count", () => {
		const result = calculateStatistics([[[0.1, 0.2]], [[0.3]]]);

		expect(result.maximum).toBeCloseTo(0.3);
		expect(result.minimum).toBeCloseTo(0.1);
		expect(result.sum).toBeCloseTo(0.6);
		expect(result.average).toBeCloseTo(0.2);
	});

	it("includes all Q and R values while accepting a rectangular matrix", () => {
		const result = calculateStatistics([
			[
				[0.6, 0.8],
				[-0.8, 0.6],
				[0, 0],
			],
			[
				[5, 6],
				[0, 7],
			],
		]);

		expect(result.maximum).toBe(7);
		expect(result.minimum).toBe(-0.8);
		expect(result.sum).toBeCloseTo(19.2);
		expect(result.average).toBeCloseTo(1.92);
	});

	it("recognizes square diagonal matrices, including one-by-one matrices", () => {
		expect(
			calculateStatistics([
				[
					[4, 0],
					[0, -3],
				],
			]).hasDiagonalMatrix,
		).toBe(true);
		expect(calculateStatistics([[[42]]]).hasDiagonalMatrix).toBe(true);
	});

	it("rejects a square matrix with a non-zero off-diagonal value", () => {
		const result = calculateStatistics([
			[
				[1, 0.01],
				[0, 2],
			],
		]);

		expect(result.hasDiagonalMatrix).toBe(false);
	});

	it("reports a diagonal matrix when another matrix in the collection is not diagonal", () => {
		const result = calculateStatistics([
			[
				[1, 3],
				[0, 2],
			],
			[
				[4, 0],
				[0, 5],
			],
		]);

		expect(result.hasDiagonalMatrix).toBe(true);
	});

	it("uses a relative tolerance for floating-point off-diagonal round-off", () => {
		const atThreshold = calculateStatistics([
			[
				[1_000_000, 0.000001],
				[0, 2],
			],
		]);
		const aboveThreshold = calculateStatistics([
			[
				[1_000_000, 0.000002],
				[0, 2],
			],
		]);

		expect(atThreshold.hasDiagonalMatrix).toBe(true);
		expect(aboveThreshold.hasDiagonalMatrix).toBe(false);
	});

	it("does not mutate valid inputs", () => {
		const matrices = [
			[1, 0],
			[0, 2],
		];
		const original = structuredClone(matrices);

		calculateStatistics([matrices]);

		expect(matrices).toEqual(original);
	});

	it.each([
		["an empty collection", []],
		["an empty matrix", [[]]],
		["an empty row", [[[]]]],
		["ragged rows", [[[1, 2], [3]]]],
		["a non-array input", "not a matrix collection"],
		["a non-numeric scalar", [[[1, "2"]]]],
		["NaN", [[[Number.NaN]]]],
		["positive infinity", [[[Infinity]]]],
		["negative infinity", [[[-Infinity]]]],
	])("rejects %s", (_description, input) => {
		expect(() => calculateStatistics(input)).toThrow(TypeError);
	});
});
