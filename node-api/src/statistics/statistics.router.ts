import { Router } from "express";

import { calculateStatisticsResponse } from "./statistics.controller.js";

export const statisticsRouter = Router();

statisticsRouter.post("/api/v1/statistics", (request, response) => {
	const result = calculateStatisticsResponse({
		contentType: request.get("Content-Type"),
		body: request.body,
	});

	return response.status(result.status).json(result.body);
});
