package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"backend-challenge/internal/core/ports"

	"github.com/gin-gonic/gin"
)

const userIDKey = "user_id"

func AuthMiddleware(tokens ports.TokenProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(c, http.StatusUnauthorized, "missing token")
			c.Abort()
			return
		}
		userID, err := tokens.Parse(parts[1])
		if err != nil {
			writeError(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		msg := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		slog.Info(msg,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
		)
	}
}
