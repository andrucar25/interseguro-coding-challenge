import type { Request, RequestHandler, Response } from "express";
import { jwtVerify } from "jose";

export type ServiceTokenConfig = Readonly<{
	secret: Uint8Array;
	issuer: string;
	audience: string;
}>;

export function loadServiceTokenConfig(
	environment: NodeJS.ProcessEnv,
): ServiceTokenConfig {
	return createServiceTokenConfig(
		environment.JWT_SECRET ?? "",
		environment.JWT_ISSUER ?? "",
		environment.JWT_AUDIENCE ?? "",
	);
}

export function createServiceTokenConfig(
	secret: string,
	issuer: string,
	audience: string,
): ServiceTokenConfig {
	if (
		secret.trim().length === 0 ||
		issuer.trim().length === 0 ||
		audience.trim().length === 0
	) {
		throw new Error("invalid JWT configuration");
	}

	return {
		secret: new TextEncoder().encode(secret),
		issuer,
		audience,
	};
}

export function requireServiceToken(
	config: ServiceTokenConfig,
): RequestHandler {
	return async (request, response, next): Promise<void> => {
		const token = extractBearerToken(request);
		if (token === undefined) {
			respondUnauthorized(response);
			return;
		}

		try {
			const { payload } = await jwtVerify(token, config.secret, {
				algorithms: ["HS256"],
				issuer: config.issuer,
				audience: config.audience,
			});
			if (typeof payload.exp !== "number" || !Number.isFinite(payload.exp)) {
				respondUnauthorized(response);
				return;
			}
			next();
		} catch {
			respondUnauthorized(response);
		}
	};
}

function extractBearerToken(request: Request): string | undefined {
	const authorizationHeaders: string[] = [];
	for (let index = 0; index < request.rawHeaders.length; index += 2) {
		if (request.rawHeaders[index]?.toLowerCase() === "authorization") {
			authorizationHeaders.push(request.rawHeaders[index + 1] ?? "");
		}
	}
	if (authorizationHeaders.length !== 1) {
		return undefined;
	}

	const match = /^Bearer ([^\s]+)$/.exec(authorizationHeaders[0]);
	return match?.[1];
}

function respondUnauthorized(response: Response): void {
	response.status(401).json({ error: "unauthorized" });
}
