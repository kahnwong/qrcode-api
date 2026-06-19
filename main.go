package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/kahnwong/qrcode-api/qrcode"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

const (
	rateLimitWindow   = time.Minute
	rateLimitRequests = 60
)

var (
	apiKey        = os.Getenv("QRCODE_API_KEY")
	protectedURLs = []*regexp.Regexp{
		regexp.MustCompile("^/add$"),
		regexp.MustCompile("^/title/"),
	}
)

func validateAPIKey(key string) bool {
	hashedAPIKey := sha256.Sum256([]byte(apiKey))
	hashedKey := sha256.Sum256([]byte(key))

	return subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1
}

func isProtectedURL(path string) bool {
	path = strings.ToLower(path)
	for _, pattern := range protectedURLs {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isProtectedURL(c.Request.URL.Path) {
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" || !validateAPIKey(apiKey) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid API key"})
				return
			}
		}
		c.Next()
	}
}

func rateLimitMiddleware() gin.HandlerFunc {
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Rate:  rateLimitWindow,
		Limit: rateLimitRequests,
	})

	return ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: rateLimitErrorHandler,
		KeyFunc:      func(c *gin.Context) string { return c.ClientIP() },
	})
}

func rateLimitErrorHandler(c *gin.Context, info ratelimit.Info) {
	retryAfter := time.Until(info.ResetTime).Seconds()
	if retryAfter < 0 {
		retryAfter = 0
	}

	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":       "Too many requests",
		"retry_after": int(retryAfter),
	})
}

func main() {
	// init
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	logLevel := zerolog.InfoLevel
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		parsedLogLevel, err := zerolog.ParseLevel(envLogLevel)
		if err != nil {
			fmt.Println("Invalid LOG_LEVEL, defaulting to info", err)
		} else {
			logLevel = parsedLogLevel
		}
	}
	zerolog.SetGlobalLevel(logLevel)
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	zerologger := zerolog.New(output).With().Timestamp().Logger()
	slog.SetDefault(slog.New(slogzerolog.Option{Logger: &zerologger}.NewZerologHandler()))
	router.Use(logger.SetLogger(logger.WithLogger(func(_ *gin.Context, l zerolog.Logger) zerolog.Logger {
		return zerologger
	})))

	// rate limiting
	router.Use(rateLimitMiddleware())

	// auth
	router.Use(authMiddleware())

	// routes
	router.GET("/title/:id", qrcode.TitleGetController)
	router.GET("/image/:id", qrcode.ImageGetController)
	router.POST("/add", qrcode.AddPostController)

	// start server
	err := router.Run(os.Getenv("LISTEN_ADDR"))
	if err != nil {
		fmt.Println("Error starting server", err)
	}
}
