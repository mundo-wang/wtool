package wresp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HandlerWrapper func(c *gin.Context) (interface{}, error)

type Server struct {
	Router *gin.Engine
}

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	PrintInfo string      `json:"PrintInfo"` // 返回到前端，只要不为空字符串，弹窗展示给用户
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

func handleErrorResponse(c *gin.Context, err error) {
	response := &Response{}
	if e, ok := err.(*ErrorCode); ok {
		response.Code = e.Code
		response.Message = e.Message
		response.PrintInfo = e.Error()
	} else {
		response.Code = -1
		response.Message = "未知错误"
		response.PrintInfo = "内部错误，请联系平台工作人员"
	}
	c.JSON(http.StatusInternalServerError, response)
}

func (s *Server) WrapHandler(handler HandlerWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := handler(c)
		if err != nil {
			handleErrorResponse(c, err)
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
