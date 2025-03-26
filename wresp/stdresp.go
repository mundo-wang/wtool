package wresp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HandlerWrapper func(c *gin.Context) (interface{}, error)

type MiddlewareWrapper func(c *gin.Context) error

type Server struct {
	Router *gin.Engine
}

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	PrintInfo string      `json:"printInfo,omitempty"` // 返回到前端，只要有该字段，弹窗展示给用户
	Data      interface{} `json:"data"`
}

type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ErrorCode) Error() string {
	return fmt.Sprintf("错误码: %d，错误原因: %s", e.Code, e.Message)
}

func NewErrorCode(code int, message string) *ErrorCode {
	return &ErrorCode{
		Code:    code,
		Message: message,
	}
}

// IsErrorCode 用于判断给定的错误是否为定义的错误码（即是否为ErrorCode类型）
func IsErrorCode(err error) bool {
	_, ok := err.(*ErrorCode)
	return ok
}

func handleErrorResponse(c *gin.Context, err error, abort bool) {
	response := &Response{}
	if e, ok := err.(*ErrorCode); ok {
		response.Code = e.Code
		response.Message = e.Message
		response.PrintInfo = e.Error()
	} else {
		response.Code = -1
		response.Message = e.Error()
		response.PrintInfo = "内部错误，请联系平台工作人员"
	}
	if abort {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response)
	} else {
		c.JSON(http.StatusInternalServerError, response)
	}
}

func (s *Server) WrapHandler(wrapper HandlerWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, false)
			return
		}
		response := &Response{
			Code:    0,
			Message: "成功",
			Data:    data,
		}
		c.JSON(http.StatusOK, response)
	}
}

func (s *Server) WrapMiddleware(wrapper MiddlewareWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, true)
			return
		}
	}
}
