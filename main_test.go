package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestAuthMiddleware(t *testing.T) {
	app := fiber.New()
	app.Get("/title/:id", authMiddleware("secret"), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/image/:id", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	tests := []struct {
		name       string
		path       string
		key        string
		wantStatus int
	}{
		{name: "missing key", path: "/title/1", wantStatus: fiber.StatusUnauthorized},
		{name: "invalid key", path: "/title/1", key: "invalid", wantStatus: fiber.StatusUnauthorized},
		{name: "valid key", path: "/title/1", key: "secret", wantStatus: fiber.StatusOK},
		{name: "public route", path: "/image/1", wantStatus: fiber.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.key != "" {
				req.Header.Set("X-API-Key", tt.key)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(rateLimitMiddleware())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	for range rateLimitRequests {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("allowed request status = %d", resp.StatusCode)
		}
	}

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("limited request status = %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	var payload struct {
		Error      string `json:"error"`
		RetryAfter int    `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	if payload.Error != "Too many requests" || payload.RetryAfter <= 0 {
		t.Fatalf("unexpected response: %+v", payload)
	}
}
