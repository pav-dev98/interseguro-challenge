import express from "express";

export const app = express();

app.get("/health", (_request, response) => {
  response.json({ status: "active" });
});

app.get("/internal/health", (_request, response) => {
  response.json({ status: "active", service: "node-api" });
});
