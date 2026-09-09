import { SignJWT } from "jose";
import request from "supertest";
import { describe, expect, it } from "vitest";

import { createApp } from "./app.js";
import { createServiceTokenConfig } from "./auth/service-token.js";

const testSecret = "test-service-token-secret";
const testIssuer = "interseguro-go-api";
const testAudience = "interseguro-node-api";
const app = createApp(
	createServiceTokenConfig(testSecret, testIssuer, testAudience),
);

describe("POST /api/v1/statistics", () => {
	it("returns 401 when Authorization is absent", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it.each([
		["an empty Bearer token", "Bearer"],
		["a malformed Bearer token", "Bearer malformed.token"],
		["an unsupported authorization scheme", "Basic credentials"],
	])("returns 401 for %s", async (_description, authorization) => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", authorization)
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it("returns 401 for a token with an invalid signature", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set(
				"Authorization",
				`Bearer ${await createToken({ secret: "different-secret" })}`,
			)
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it("returns 401 for an expired token", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set(
				"Authorization",
				`Bearer ${await createToken({
					expiresAt: Math.floor(Date.now() / 1000) - 60,
				})}`,
			)
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it("returns 401 for a token with the wrong issuer", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set(
				"Authorization",
				`Bearer ${await createToken({ issuer: "other-issuer" })}`,
			)
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it("returns 401 for a token with the wrong audience", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set(
				"Authorization",
				`Bearer ${await createToken({ audience: "other-audience" })}`,
			)
			.send({ q: [[1]], r: [[1]] });

		expectUnauthorized(response);
	});

	it("returns statistics from valid Q and R matrices", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", await authorizedHeader())
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
			.set("Authorization", await authorizedHeader())
			.set("Content-Type", "application/json")
			.send('{"q":');

		expect(response.status).toBe(400);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({ error: "invalid_request" });
	});

	it("returns 415 for a non-JSON request media type", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", await authorizedHeader())
			.set("Content-Type", "text/plain")
			.send('{"q":[[1]],"r":[[1]]}');

		expect(response.status).toBe(415);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({ error: "invalid_request" });
	});

	it("returns 413 for JSON larger than the parser limit", async () => {
		const body = JSON.stringify({
			q: [[1]],
			r: [[1]],
			padding: "x".repeat(100 * 1024),
		});

		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", await authorizedHeader())
			.set("Content-Type", "application/json")
			.send(body);

		expect(response.status).toBe(413);
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
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", await authorizedHeader())
			.send(body);

		expect(response.status).toBe(400);
		expect(response.headers["content-type"]).toMatch(/^application\/json/);
		expect(response.body).toEqual({ error: "invalid_request" });
	});

	it("returns 400 for non-finite matrix values", async () => {
		const response = await request(app)
			.post("/api/v1/statistics")
			.set("Authorization", await authorizedHeader())
			.set("Content-Type", "application/json")
			.send('{"q":[[1e999]],"r":[[1]]}');

		expect(response.status).toBe(400);
		expect(response.body).toEqual({ error: "invalid_request" });
	});
});

async function authorizedHeader(): Promise<string> {
	return `Bearer ${await createToken()}`;
}

async function createToken({
	secret = testSecret,
	issuer = testIssuer,
	audience = testAudience,
	expiresAt = Math.floor(Date.now() / 1000) + 60,
}: {
	secret?: string;
	issuer?: string;
	audience?: string;
	expiresAt?: number;
} = {}): Promise<string> {
	return new SignJWT()
		.setProtectedHeader({ alg: "HS256" })
		.setIssuedAt()
		.setExpirationTime(expiresAt)
		.setIssuer(issuer)
		.setAudience(audience)
		.sign(new TextEncoder().encode(secret));
}

function expectUnauthorized(response: request.Response): void {
	expect(response.status).toBe(401);
	expect(response.headers["content-type"]).toMatch(/^application\/json/);
	expect(response.body).toEqual({ error: "unauthorized" });
}
