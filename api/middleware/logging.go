package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger returns a logging middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Path
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get method
		method := c.Request.Method

		// Get error if any
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Build log fields
		fields := logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"query":      query,
			"ip":         clientIP,
			"latency":    latency.String(),
			"user-agent": c.Request.UserAgent(),
		}

		// Add error message if present
		if errorMessage != "" {
			fields["error"] = errorMessage
		}

		// Log based on status code
		if statusCode >= 500 {
			logrus.WithFields(fields).Error("Server error")
		} else if statusCode >= 400 {
			logrus.WithFields(fields).Warn("Client error")
		} else {
			logrus.WithFields(fields).Info("Request")
		}
	}
}

// loggingMiddleware is used by the server
func loggingMiddleware() gin.HandlerFunc {
	return Logger()
}

// RequestLogger logs detailed request information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request details
		logrus.WithFields(logrus.Fields{
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"query":   c.Request.URL.RawQuery,
			"ip":      c.ClientIP(),
			"headers": c.Request.Header,
		}).Debug("Incoming request")

		c.Next()
	}
}

// ResponseLogger logs response details
func ResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create response recorder
		startTime := time.Now()

		c.Next()

		logrus.WithFields(logrus.Fields{
			"status":   c.Writer.Status(),
			"size":     c.Writer.Size(),
			"latency":  time.Since(startTime).String(),
		}).Debug("Response sent")
	}
}
