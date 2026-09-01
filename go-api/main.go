package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const dependencyRequestTimeout = 3 * time.Second

func main() {
	app := fiber.New()
	httpClient := &http.Client{Timeout: dependencyRequestTimeout}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "active"})
	})
	app.Get("/health/dependencies", dependenciesHealthHandler(httpClient))

	app.Listen(":3001")
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
