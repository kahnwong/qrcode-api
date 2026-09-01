package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/kahnwong/qrcode-api/qrcode"
	"github.com/rs/zerolog"
	slogfiber "github.com/samber/slog-fiber"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

const (
	rateLimitWindow      = time.Minute
	rateLimitRequests    = 60
	defaultListenAddress = "127.0.0.1:3000"
)

func authMiddleware(expectedKey string) fiber.Handler {
	expectedHash := sha256.Sum256([]byte(expectedKey))

	return func(c fiber.Ctx) error {
		key := c.Get("X-API-Key")
		keyHash := sha256.Sum256([]byte(key))
		if key == "" || subtle.ConstantTimeCompare(expectedHash[:], keyHash[:]) != 1 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid API key"})
		}

		return c.Next()
	}
}

func rateLimitMiddleware() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        rateLimitRequests,
		Expiration: rateLimitWindow,
		LimitReached: func(c fiber.Ctx) error {
			retryAfter, _ := strconv.Atoi(c.GetRespHeader(fiber.HeaderRetryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many requests",
				"retry_after": retryAfter,
			})
		},
	})
}

func configureLogger() {
	level := slog.LevelDebug
	var parseErr error
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		parseErr = level.UnmarshalText([]byte(logLevel))
		if parseErr != nil {
			level = slog.LevelDebug
		}
	}

	zeroLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	slog.SetDefault(slog.New(slogzerolog.Option{
		Level:  level,
		Logger: &zeroLogger,
	}.NewZerologHandler()))
	if parseErr != nil {
		slog.Warn("invalid LOG_LEVEL, defaulting to debug", "value", logLevel, "error", parseErr)
	}
}

func main() {
	configureLogger()

	if err := qrcode.Initialize(); err != nil {
		slog.Error("error initializing application", "error", err)
		os.Exit(1)
	}

	app := fiber.New()
	app.Use(slogfiber.New(slog.Default()))
	app.Use(recover.New())
	app.Use(rateLimitMiddleware())

	auth := authMiddleware(os.Getenv("QRCODE_API_KEY"))
	app.Get("/title/:id", auth, qrcode.TitleGetController)
	app.Get("/image/:id", qrcode.ImageGetController)
	app.Post("/add", auth, qrcode.AddPostController)

	listenAddress := os.Getenv("LISTEN_ADDR")
	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}
	if err := app.Listen(listenAddress); err != nil {
		slog.Error("error starting server", "address", listenAddress, "error", err)
		os.Exit(1)
	}
}
