package middleware

import (
	"net/http"
	"strings"

	"owl-engine/pkg/config"

	"github.com/gin-gonic/gin"
)

// Auth 对于每次的请求均作认证校验
func Auth() gin.HandlerFunc {
	conf := config.Get()

	return func(ctx *gin.Context) {
		token := ctx.GetHeader("auth-secret")
		if strings.Compare(token, conf.ServerOptions.Secret) == 0 {
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  false,
				"errCode": 403,
				"errMsg":  "permission deny",
			})
		}
	}
}
