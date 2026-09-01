import express from "express";

const app = express();
const port = 3002;

app.get("/health", (_request, response) => {
  response.json({ status: "active" });
});

app.listen(port, () => {
  console.log(`Node API listening on port ${port}`);
});
