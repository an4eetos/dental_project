package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestID attaches a request id to context and response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = randomID()
		}
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}

// SlogLogger logs requests with structured fields.
func SlogLogger(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		status := c.Writer.Status()
		dur := time.Since(start)
		rid, _ := c.Get("request_id")

		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", dur.Milliseconds(),
			"request_id", rid,
		}
		if origin := c.GetHeader("Origin"); origin != "" {
			attrs = append(attrs, "origin", origin)
		}
		if status >= 400 {
			log.Warn("http", attrs...)
			return
		}
		log.Info("http", attrs...)
	}
}

// Recovery catches panics and returns 500 without leaking details.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		rid, _ := c.Get("request_id")
		log.Error("panic recovered", "request_id", rid, "recovered", recovered)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// WebhookSecret optionally validates Telegram webhook secret token header.
func WebhookSecret(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.Next()
			return
		}
		got := c.GetHeader("X-Telegram-Bot-Api-Secret-Token")
		if got != secret {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func randomID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	return hex.EncodeToString(b)
}
