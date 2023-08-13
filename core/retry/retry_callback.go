package retry

// RtyCallback 重试操作的回调接口
type RtyCallback interface {
	// DoWithRetry 执行具有重试语意的操作。重试操作一般是需要幂等的，但是业务自己可以选择重试语意的操作
	DoWithRetry(content RtyContext) interface{}
}
