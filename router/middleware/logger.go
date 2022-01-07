package middleware

import (
	"time"

	"owl-engine/pkg/xlogs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerApi() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start time
		startTime := time.Now()
		//	deal req
		c.Next()

		// end time
		endTime := time.Now()

		// latency duration
		latencyDuration := endTime.Sub(startTime)

		// request method
		reqMethod := c.Request.Method

		// request uri
		reqUri := c.Request.RequestURI

		// status code
		statusCode := c.Writer.Status()

		// client IP
		clientIP := c.ClientIP()

		fields := []zap.Field{
			xlogs.Int("status_code", statusCode),
			xlogs.Duration("latencyDuration", latencyDuration),
			xlogs.String("client_ip", clientIP),
			xlogs.String("request_method", reqMethod),
			xlogs.String("request_uri", reqUri),
		}
		xlogs.With(fields...).Info("a new request input")
	}
}
