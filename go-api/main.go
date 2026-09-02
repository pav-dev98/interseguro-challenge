package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/interseguro/challenge/go-api/nodeclient"
	"github.com/interseguro/challenge/go-api/qr"
)

const dependencyRequestTimeout = nodeclient.RequestTimeout

func main() {
	app := fiber.New()
	httpClient := &http.Client{Timeout: dependencyRequestTimeout}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: frontendURL,
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "active"})
	})
	app.Get("/health/dependencies", dependenciesHealthHandler(httpClient))
	app.Post("/qr", qrHandler(nodeclient.New(os.Getenv("NODE_API_URL"), nil)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	app.Listen(":" + port)
}

func qrHandler(client nodeclient.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var request struct {
			Matrix *[][]float64 `json:"matrix"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request body must be valid JSON"})
		}
		if request.Matrix == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "matrix is required"})
		}

		q, r, err := qr.Factorize(*request.Matrix)
		if err != nil {
			status := fiber.StatusBadRequest
			if errors.Is(err, qr.ErrDependentColumns) {
				status = fiber.StatusUnprocessableEntity
			}
			return c.Status(status).JSON(fiber.Map{"error": err.Error()})
		}

		ctx, cancel := context.WithTimeout(context.Background(), nodeclient.RequestTimeout)
		defer cancel()

		statistics, err := client.Statistics(ctx, q, r)
		if err != nil {
			status := fiber.StatusServiceUnavailable
			if errors.Is(err, nodeclient.ErrUnexpectedStatus) || errors.Is(err, nodeclient.ErrInvalidResponse) {
				status = fiber.StatusBadGateway
			}
			return c.Status(status).JSON(fiber.Map{"error": "statistics service is unavailable"})
		}

		return c.JSON(fiber.Map{"q": q, "r": r, "statistics": statistics})
	}
}

func dependenciesHealthHandler(client *http.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		nodeAPIURL := strings.TrimRight(os.Getenv("NODE_API_URL"), "/")
		if nodeAPIURL == "" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "NODE_API_URL is not configured"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), dependencyRequestTimeout)
		defer cancel()

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, nodeAPIURL+"/internal/health", nil)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "node-api is unavailable"})
		}

		response, err := client.Do(request)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "node-api is unavailable"})
		}
		defer response.Body.Close()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "node-api is unavailable"})
		}

		return c.JSON(fiber.Map{
			"status": "active",
			"dependencies": fiber.Map{
				"node-api": "active",
			},
		})
	}
}
