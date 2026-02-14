package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/kahnwong/qrcode-api/qrcode"
	"github.com/rs/zerolog"
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

func main() {
	// init
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerologger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	router.Use(logger.SetLogger(logger.WithLogger(func(_ *gin.Context, l zerolog.Logger) zerolog.Logger {
		return zerologger
	})))

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
