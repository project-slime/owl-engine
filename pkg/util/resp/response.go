package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"owl-engine/pkg/model/constParam"
)

type ResultResp struct {
	Status     bool   `json:"status"`
	StatusCode string `json:"status_code"`
	Message    string `json:"message"`
}

type JsonResultResp struct {
	Status     bool        `json:"status"`
	StatusCode string      `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

func SuccessResp(c *gin.Context, statusCode, respMsg string) {
	resp(c, http.StatusOK, constParam.StatusSuccess, statusCode, respMsg)
}

func ErrorResp(c *gin.Context, statusCode, respMsg string) {
	resp(c, http.StatusBadRequest, constParam.StatusFail, statusCode, respMsg)
}

func SuccessJsonResp(c *gin.Context, statusCode, respMsg string, data interface{}) {
	jsonResp(c, http.StatusOK, constParam.StatusSuccess, statusCode, respMsg, data)
}

func resp(c *gin.Context, code int, status bool, statusCode, respMsg string) {
	response := ResultResp{Status: status, StatusCode: statusCode, Message: respMsg}
	c.JSON(code, response)
}

func jsonResp(c *gin.Context, code int, status bool, statusCode, respMsg string, data interface{}) {
	response := JsonResultResp{Status: status, StatusCode: statusCode, Message: respMsg, Data: data}
	c.JSON(code, response)
}
