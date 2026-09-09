import { Router } from "express";

import { calculateStatistics } from "./statistics.js";
import { statisticsRequestSchema } from "./statistics-request.schema.js";

export const statisticsRouter = Router();

statisticsRouter.post("/api/v1/statistics", (request, response) => {
	const parsedRequest = statisticsRequestSchema.safeParse(request.body);

	if (!parsedRequest.success) {
		return response.status(400).json({ error: "invalid_request" });
	}

	const statistics = calculateStatistics([
		parsedRequest.data.q,
		parsedRequest.data.r,
	]);

	return response.status(200).json(statistics);
});
