const DEFAULT_PORT = 8080;
const INVALID_PORT_MESSAGE =
	"Invalid PORT: expected an integer from 1 to 65535.";

export function resolvePort(value: string | undefined): number {
	if (value === undefined) {
		return DEFAULT_PORT;
	}

	if (!/^\d+$/.test(value)) {
		throw new Error(INVALID_PORT_MESSAGE);
	}

	const port = Number(value);
	if (!Number.isSafeInteger(port) || port < 1 || port > 65535) {
		throw new Error(INVALID_PORT_MESSAGE);
	}

	return port;
}
