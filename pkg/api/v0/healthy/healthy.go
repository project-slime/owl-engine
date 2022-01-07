package healthy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Pong(ctx *gin.Context) {
	var data = make(map[string]interface{})
	data["PING"] = "PONG"

	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"errCode": 0,
		"errMsg":  "ok",
		"data":    data,
	})
}

// 系统资源消耗的统计
func Status(ctx *gin.Context) {
	// TODO: 统计系统运行的资源消耗
	ctx.JSON(http.StatusOK, gin.H{
		"status":  true,
		"errCode": 0,
		"errMsg":  "ok",
		"data":    make(map[string]interface{}, 0),
	})
}
