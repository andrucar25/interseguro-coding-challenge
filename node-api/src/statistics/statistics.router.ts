import { type RequestHandler, Router } from "express";

import { calculateStatisticsResponse } from "./statistics.controller.js";

export function statisticsRouter(
	requireServiceToken: RequestHandler,
	parseJson: RequestHandler,
) {
	const router = Router();

	router.post(
		"/api/v1/statistics",
		requireServiceToken,
		parseJson,
		(request, response) => {
			const result = calculateStatisticsResponse({
				contentType: request.get("Content-Type"),
				body: request.body,
			});

			return response.status(result.status).json(result.body);
		},
	);

	return router;
}
