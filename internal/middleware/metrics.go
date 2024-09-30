package middleware

import (
	"net/http"
	"time"

	"github.com/berikulyBeket/todo-plus/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMetricsMiddleware handles infrastructure metrics collection
func TrackRequestMetrics(m metrics.Interface) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/metrics" {
			c.Next()
			return
		}

		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime).Seconds()

		m.TrackRequestDuration(c.FullPath(), duration)
		m.TrackRequestCount(c.FullPath(), http.StatusText(c.Writer.Status()))
	}
}
