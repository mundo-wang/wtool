package wcode

var (
	Success = NewErrorCode(10000, "成功")
	Failed  = NewErrorCode(10001, "失败")
	Unknown = NewErrorCode(10002, "未知异常")
)
