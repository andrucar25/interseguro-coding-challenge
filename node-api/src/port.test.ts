import { describe, expect, it } from "vitest";

import { resolvePort } from "./port.js";

const invalidPortMessage = "Invalid PORT: expected an integer from 1 to 65535.";

describe("resolvePort", () => {
	it.each([
		[undefined, 8080],
		["8080", 8080],
		["3000", 3000],
	])("returns %i for %j", (value, expectedPort) => {
		expect(resolvePort(value)).toBe(expectedPort);
	});

	it.each(["not-a-port", "3000abc", "0", "-1", "65536"])(
		"rejects invalid port %j",
		(value) => {
			expect(() => resolvePort(value)).toThrow(invalidPortMessage);
		},
	);
});
