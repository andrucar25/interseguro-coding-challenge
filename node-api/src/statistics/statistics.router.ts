import { Router } from "express";

import { calculateStatistics } from "./service/statistics.service.js";
import { statisticsRequestSchema } from "./validations/statistics-request.schema.js";

export const statisticsRouter = Router();

statisticsRouter.post("/api/v1/statistics", (request, response) => {
	if (
		request.get("Content-Type") !== undefined &&
		!request.is("application/json")
	) {
		return response.status(415).json({ error: "invalid_request" });
	}

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
