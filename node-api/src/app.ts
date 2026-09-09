import express, {
	type NextFunction,
	type Request,
	type Response,
} from "express";
import helmet from "helmet";

import {
	requireServiceToken,
	type ServiceTokenConfig,
} from "./auth/service-token.js";
import { statisticsRouter } from "./statistics/statistics.router.js";

function hasBodyParserErrorDetails(
	error: unknown,
): error is Error & { status: number; type: string } {
	return (
		error instanceof Error &&
		"status" in error &&
		"type" in error &&
		typeof error.status === "number" &&
		typeof error.type === "string"
	);
}

function isMalformedJsonError(error: unknown): boolean {
	return (
		error instanceof SyntaxError &&
		hasBodyParserErrorDetails(error) &&
		error.status === 400 &&
		error.type === "entity.parse.failed"
	);
}

function isOversizedJsonError(error: unknown): boolean {
	return (
		hasBodyParserErrorDetails(error) &&
		error.status === 413 &&
		error.type === "entity.too.large"
	);
}

export function createApp(serviceTokenConfig: ServiceTokenConfig) {
	const app = express();

	app.use(helmet());
	app.use(
		statisticsRouter(requireServiceToken(serviceTokenConfig), express.json()),
	);

	app.use(
		(
			error: unknown,
			_request: Request,
			response: Response,
			_next: NextFunction,
		): void => {
			if (isMalformedJsonError(error)) {
				response.status(400).json({ error: "invalid_request" });
				return;
			}

			if (isOversizedJsonError(error)) {
				response.status(413).json({ error: "invalid_request" });
				return;
			}

			response.status(500).json({ error: "internal_error" });
		},
	);

	return app;
}
