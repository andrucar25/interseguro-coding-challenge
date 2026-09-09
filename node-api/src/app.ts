import express, {
	type NextFunction,
	type Request,
	type Response,
} from "express";
import helmet from "helmet";

import { statisticsRouter } from "./statistics/statistics.router.js";

function isMalformedJsonError(error: unknown): boolean {
	return (
		error instanceof SyntaxError && "status" in error && error.status === 400
	);
}

export const app = express();

app.use(helmet());
app.use(express.json());
app.use(statisticsRouter);

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

		response.status(500).json({ error: "internal_error" });
	},
);
