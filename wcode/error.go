package wcode

import "fmt"

type errorCode struct {
	Code        int
	Message     string
	InternalMsg string
}

type RetCode struct {
	Code      *errorCode
	Cause     error
	RequestId string
}

type ErrorCode interface {
	AddInternalMsg(internalMsg string) ErrorCode
	AddInternalMsgf(internalMsgf string, args ...interface{}) ErrorCode
	NewError() error
	NewErrorWithCause(cause error) error
	NewErrorWithRequestId(requestId string) error
	NewErrorWithCauseAndRequestId(cause error, requestId string) error
	GetCode() int
	Equals(e1 ErrorCode) bool
}

func (e *RetCode) Error() string {
	if e.Code != nil {
		return fmt.Sprintf("Code: %d, Message: %s, RequestId: %s", e.Code.Code, e.Code.Message, e.RequestId)
	} else if e.Cause != nil {
		return fmt.Sprintf("Error: %s, RequestId: %s", e.Cause.Error(), e.RequestId)
	}
	return "Unknown error"
}

func NewErrorCode(Code int, Message string) ErrorCode {
	return &errorCode{
		Code:    Code,
		Message: Message,
	}
}

func (e *errorCode) AddInternalMsg(internalMsg string) ErrorCode {
	e.AddInternalMsgf(internalMsg)
	return e
}

func (e *errorCode) AddInternalMsgf(internalMsgf string, args ...interface{}) ErrorCode {
	e.InternalMsg = fmt.Sprintf(internalMsgf, args...)
	return e
}

func (e *errorCode) NewError() error {
	return e.NewErrorWithCauseAndRequestId(nil, "")
}

func (e *errorCode) NewErrorWithCause(cause error) error {
	return e.NewErrorWithCauseAndRequestId(cause, "")
}

func (e *errorCode) NewErrorWithRequestId(requestId string) error {
	return e.NewErrorWithCauseAndRequestId(nil, requestId)
}

func (e *errorCode) NewErrorWithCauseAndRequestId(cause error, requestId string) error {
	return &RetCode{
		Code:      e,
		Cause:     cause,
		RequestId: requestId,
	}
}

func (e *errorCode) GetCode() int {
	return e.Code
}

func (e *errorCode) Equals(e1 ErrorCode) bool {
	return e.Code == e1.GetCode()
}
