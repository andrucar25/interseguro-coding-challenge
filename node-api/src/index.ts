import { app } from "./app.js";
import { resolvePort } from "./port.js";

const port = resolvePort(process.env.PORT);

app.listen(port, "0.0.0.0", () => {
	console.info(`Node API listening on port ${port}.`);
});
