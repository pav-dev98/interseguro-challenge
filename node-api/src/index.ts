import { app } from "./app";

const configuredPort = Number(process.env.PORT);
const port = Number.isInteger(configuredPort) && configuredPort > 0 ? configuredPort : 3002;

app.listen(port, () => {
  console.log(`Node API listening on port ${port}`);
});
