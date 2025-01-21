package wcode

import (
	"fmt"
	"time"
)

type ErrorCode struct {
	Code      int
	Message   string
	Cause     error
	Timestamp time.Time
}

func (e *ErrorCode) Error() string {
	return fmt.Sprintf("Code: %s, Message: %s, Time: %s", e.Code, e.Message, e.Timestamp)
}
