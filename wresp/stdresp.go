package wresp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

type handlerWrapper[T interface{}] func(c *gin.Context) (T, error)

type middlewareWrapper func(c *gin.Context) error

type fileDownloadWrapper func(c *gin.Context) (string, error)

type streamHandlerWrapper func(c *gin.Context) error

type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"` // 如果code值不为0，前端展示message内容给用户
	Data    interface{} `json:"data"`    // 接口调用成功时返回的数据
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
	return NewErrorCodeWithStatus(code, message, http.StatusInternalServerError) // 默认设置HTTP状态码500
}

func NewErrorCodeWithStatus(code int, message string, httpStatus int) error {
	return &errorCode{
		code:       code,
		message:    message,
		httpStatus: httpStatus,
	}
}

func IsErrorCode(target error) bool {
	_, ok := target.(*errorCode)
	return ok
}

// 接口返回错误时，data字段始终返回nil，避免语义不清
func handleErrorResponse(c *gin.Context, err error, abort bool) {
	resp := &response{}
	httpStatus := http.StatusInternalServerError
	if e, ok := err.(*errorCode); ok {
		resp.Code = e.code
		resp.Message = e.Error() // 将Error()方法返回的格式字符串写入到message
		httpStatus = e.httpStatus
	} else {
		resp.Code = -1
		resp.Message = "内部错误，请联系平台工作人员"
	}
	if abort {
		c.AbortWithStatusJSON(httpStatus, resp)
	} else {
		c.JSON(httpStatus, resp)
	}
}

func writeStreamError(c *gin.Context, err error) {
	if !c.Writer.Written() {
		handleErrorResponse(c, err, false)
		return
	}
	resp := &response{}
	if e, ok := err.(*errorCode); ok {
		resp.Code = e.code
		resp.Message = e.Error()
	} else {
		resp.Code = -1
		resp.Message = "内部错误，请联系平台工作人员"
	}
	c.SSEvent("error", resp)
	c.Writer.Flush()
}

func WrapHandler[T interface{}](wrapper handlerWrapper[T]) gin.HandlerFunc {
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

func WrapMiddleware(wrapper middlewareWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, true)
			return
		}
	}
}

func WrapFileDownload(wrapper fileDownloadWrapper, download bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath, err := wrapper(c)
		if err != nil {
			handleErrorResponse(c, err, false)
			return
		}
		if download {
			fileName := filepath.Base(filePath)
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		}
		c.File(filePath)
	}
}

func WrapStreamHandler(wrapper streamHandlerWrapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		err := wrapper(c)
		if err != nil {
			writeStreamError(c, err)
		}
	}
}
