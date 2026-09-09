import {
	calculateStatistics,
	type MatrixStatistics,
} from "./service/statistics.service.js";
import { statisticsRequestSchema } from "./validations/statistics-request.schema.js";

interface StatisticsControllerRequest {
	contentType: string | undefined;
	body: unknown;
}

interface StatisticsSuccessResponse {
	status: 200;
	body: MatrixStatistics;
}

interface StatisticsErrorResponse {
	status: 400 | 415;
	body: { error: "invalid_request" };
}

export type StatisticsControllerResponse =
	| StatisticsSuccessResponse
	| StatisticsErrorResponse;

function isJsonContentType(contentType: string | undefined): boolean {
	return (
		contentType === undefined ||
		/^application\/json(?:\s*;|\s*$)/i.test(contentType)
	);
}

export const calculateStatisticsResponse = ({
	contentType,
	body,
}: StatisticsControllerRequest): StatisticsControllerResponse => {
	if (!isJsonContentType(contentType)) {
		return { status: 415, body: { error: "invalid_request" } };
	}

	const parsedRequest = statisticsRequestSchema.safeParse(body);

	if (!parsedRequest.success) {
		return { status: 400, body: { error: "invalid_request" } };
	}

	return {
		status: 200,
		body: calculateStatistics([parsedRequest.data.q, parsedRequest.data.r]),
	};
};
