import request from "supertest";
import { describe, expect, it } from "vitest";

import { app } from "./app.js";

describe("POST /api/v1/statistics", () => {
	it("returns statistics from valid Q and R matrices", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.send({
				q: [
					[1, 0],
					[0, 1],
				],
				r: [
					[2, 3],
					[0, 4],
				],
			});

		expect(response.status).toBe(200);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({
			maximum: 4,
			minimum: 0,
			sum: 11,
			average: 1.375,
			hasDiagonalMatrix: true,
		});
	});

	it("returns a stable JSON error for malformed JSON", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Content-Type", "application/json")
			.send('{"q":');

		expect(response.status).toBe(400);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({ error: "invalid_request" });
	});

	it.each([
		["an absent body", undefined],
		["a missing q matrix", { r: [[1]] }],
		["a missing r matrix", { q: [[1]] }],
		["an empty matrix", { q: [], r: [[1]] }],
		["an empty row", { q: [[]], r: [[1]] }],
		["a non-array matrix", { q: "not a matrix", r: [[1]] }],
		["an invalid nested structure", { q: [[1], "invalid"], r: [[1]] }],
		["ragged rows", { q: [[1, 2], [3]], r: [[1]] }],
		["a non-numeric value", { q: [["1"]], r: [[1]] }],
		["an unexpected top-level field", { q: [[1]], r: [[1]], extra: true }],
	])("returns 400 for %s", async (_description, body) => {
		const response = await request(app).post("/api/v1/statistics").send(body);

		expect(response.status).toBe(400);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({ error: "invalid_request" });
	});

	it("returns 400 for non-finite matrix values", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Content-Type", "application/json")
			.send('{"q":[[1e999]],"r":[[1]]}');

		expect(response.status).toBe(400);
		expect(response.body).toEqual({ error: "invalid_request" });
	});
});
