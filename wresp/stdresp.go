package wresp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	RespCodeSuccess = 0
	RespCodeFailed  = -1
	RespCodeAbort   = -2
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func OK(c *gin.Context, data interface{}) {
	resp := NewResponse(RespCodeSuccess, "Success", data)
	c.JSON(http.StatusOK, resp)
}

func OKWithMsg(c *gin.Context, message string, data interface{}) {
	resp := NewResponse(RespCodeSuccess, message, data)
	c.JSON(http.StatusOK, resp)
}

func Fail(c *gin.Context, statusCode int, message string) {
	resp := NewResponse(RespCodeFailed, message, nil)
	c.JSON(statusCode, resp)
}

func FailWithData(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := NewResponse(RespCodeFailed, message, data)
	c.JSON(statusCode, resp)
}

// 使用在 Gin 的中间件里
func Abort(c *gin.Context, statusCode int, message string, data interface{}) {
	resp := NewResponse(RespCodeAbort, message, data)
	c.AbortWithStatusJSON(statusCode, resp)
}
