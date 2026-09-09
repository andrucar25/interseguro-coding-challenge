import { app } from "./app.js";

const port = Number.parseInt(process.env.PORT ?? "8080", 10);

app.listen(port, "0.0.0.0", () => {
	console.info(`Node API listening on port ${port}.`);
});
