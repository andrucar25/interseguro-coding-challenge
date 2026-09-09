import { z } from "zod";

const finiteNumberSchema = z.number().refine(Number.isFinite, {
	message: "Matrix values must be finite numbers.",
});

const matrixSchema = z
	.array(z.array(finiteNumberSchema).min(1))
	.min(1)
	.superRefine((matrix, context) => {
		const columnCount = matrix[0]?.length;

		for (const row of matrix) {
			if (row.length !== columnCount) {
				context.addIssue({
					code: "custom",
					message: "Matrix rows must have the same length.",
				});
				return;
			}
		}
	});

export const statisticsRequestSchema = z
	.object({
		q: matrixSchema,
		r: matrixSchema,
	})
	.strict();

export type StatisticsRequest = z.infer<typeof statisticsRequestSchema>;
