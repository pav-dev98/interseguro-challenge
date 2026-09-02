import express from "express";
import { calculateStatistics } from "./statistics/statistics";

export const app = express();

app.use(express.json());

app.get("/health", (_request, response) => {
  response.json({ status: "active" });
});

app.get("/internal/health", (_request, response) => {
  response.json({ status: "active", service: "node-api" });
});

app.post("/statistics", (request, response) => {
  try {
    response.json(calculateStatistics(request.body?.q, request.body?.r));
  } catch (error) {
    const message = error instanceof Error ? error.message : "invalid statistics request";
    response.status(400).json({ error: message });
  }
});

app.use((error: Error & { status?: number }, _request: express.Request, response: express.Response, next: express.NextFunction) => {
  if (error instanceof SyntaxError && error.status === 400) {
    response.status(400).json({ error: "request body must be valid JSON" });
    return;
  }

  next(error);
});
