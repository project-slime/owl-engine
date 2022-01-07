package middleware

import "github.com/gin-gonic/gin"

func InitMiddleware(r *gin.Engine) {
	// 全局中间件
	r.Use(LoggerApi())

	// oanic 捕获
	r.Use(gin.Recovery())

	r.Use(NoCache)
	r.Use(Options)
	r.Use(Secure)
}
