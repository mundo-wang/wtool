package wresp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

type handlerWrapper func(c *gin.Context) (interface{}, error)

type middlewareWrapper func(c *gin.Context) error

type fileDownloadWrapper func(c *gin.Context) (string, error)

type Server struct {
	Router *gin.Engine
}

type response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	PrintInfo string      `json:"printInfo,omitempty"` // 返回到前端，只要有该字段，弹窗展示给用户
	Data      interface{} `json:"data"`
}

type errorCode struct {
	code       int
	message    string
	httpStatus int
}

func (e *errorCode) Error() string {
	return fmt.Sprintf("错误码: %d，错误原因: %s", e.code, e.message)
}

func NewErrorCode(code int, message string) error {
	return NewErrorCodeWithStatus(code, message, http.StatusInternalServerError) // 默认设置状态码500
}

func NewErrorCodeWithStatus(code int, message string, httpStatus int) error {
	return &errorCode{
		code:       code,
		message:    message,
		httpStatus: httpStatus,
	}
}

// IsErrorCode 用于判断给定的错误是否为定义的错误码（即是否为*errorCode类型）
func IsErrorCode(target error) bool {
	_, ok := target.(*errorCode)
	return ok
}

func handleErrorResponse(c *gin.Context, err error, abort bool) {
	resp := &response{}
	httpStatus := http.StatusInternalServerError
	if e, ok := err.(*errorCode); ok {
		resp.Code = e.code
		resp.Message = e.message
		resp.PrintInfo = e.Error()
		httpStatus = e.httpStatus
	} else {
		resp.Code = -1
		resp.Message = err.Error() // 如果断言失败，e将会是nil，如果使用e.Error()会造成空指针
		resp.PrintInfo = "内部错误，请联系平台工作人员"
	}
	if abort {
		c.AbortWithStatusJSON(httpStatus, resp)
	} else {
		c.JSON(httpStatus, resp)
	}
}

func (s *Server) WrapHandler(wrapper handlerWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		data, err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, false)
			return
		}
		resp := &response{
			Code:    0,
			Message: "成功",
			Data:    data,
		}
		c.JSON(http.StatusOK, resp)
	}
}

func (s *Server) WrapMiddleware(wrapper middlewareWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, true)
			return
		}
	}
}

func (s *Server) WrapFileDownload(wrapper fileDownloadWrapper, download bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath, err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, false)
			return
		}
		if download {
			fileName := filepath.Base(filePath)
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; fileName=%s", fileName))
		}
		c.File(filePath)
	}
}
