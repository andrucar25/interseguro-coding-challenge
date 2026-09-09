import { createApp } from "./app.js";
import { loadServiceTokenConfig } from "./auth/service-token.js";
import { resolvePort } from "./port.js";

const port = resolvePort(process.env.PORT);
const app = createApp(loadServiceTokenConfig(process.env));

app.listen(port, "0.0.0.0", () => {
	console.info(`Node API listening on port ${port}.`);
});
