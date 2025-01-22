package wcode

import (
	"fmt"
	"testing"
)

func Test01(t *testing.T) {
	param := "age"
	condition := "大于30"
	requestId := "1881976623637467136"
	err := Failed.AddInternalMsgf("参数%s未满足%s条件", param, condition).NewErrorWithRequestId(requestId)
	fmt.Println(err.Error())
}
