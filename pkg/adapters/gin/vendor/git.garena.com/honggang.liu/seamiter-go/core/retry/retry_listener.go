package retry

// RtyListener 重试监听，可以用于在重试时增加其他的行为
type RtyListener interface {

	// Open 在第一次重试前调用。例如，实现这个方法可以在RtyOperations前设置状态
	Open(ctx RtyContext, callback RtyCallback) bool

	// Close 在最后一次尝试后调用(成功或者失败).允许在这个过程中清理占用的资源
	Close(ctx RtyContext, callback RtyCallback, err error)

	// OnError 每次失败重试时被调用
	OnError(ctx RtyContext, callback RtyCallback, err error)

	// OnSuccess 成功时回调
	OnSuccess(ctx RtyContext, callback RtyCallback, result interface{})
}
