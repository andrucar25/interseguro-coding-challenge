import { describe, expect, it } from "vitest";

import { calculateStatisticsResponse } from "./statistics.controller.js";

describe("calculateStatisticsResponse", () => {
	it("returns statistics without requiring Express request or response objects", () => {
		const result = calculateStatisticsResponse({
			contentType: "application/json; charset=utf-8",
			body: {
				q: [
					[1, 0],
					[0, 1],
				],
				r: [
					[2, 3],
					[0, 4],
				],
			},
		});

		expect(result).toEqual({
			status: 200,
			body: {
				maximum: 4,
				minimum: 0,
				sum: 11,
				average: 1.375,
				hasDiagonalMatrix: true,
			},
		});
	});

	it("maps invalid input to a stable client error", () => {
		const result = calculateStatisticsResponse({
			contentType: "application/json",
			body: { q: [[1]] },
		});

		expect(result).toEqual({
			status: 400,
			body: { error: "invalid_request" },
		});
	});

	it("rejects requests with a non-JSON media type", () => {
		const result = calculateStatisticsResponse({
			contentType: "text/plain",
			body: { q: [[1]], r: [[1]] },
		});

		expect(result).toEqual({
			status: 415,
			body: { error: "invalid_request" },
		});
	});
});
