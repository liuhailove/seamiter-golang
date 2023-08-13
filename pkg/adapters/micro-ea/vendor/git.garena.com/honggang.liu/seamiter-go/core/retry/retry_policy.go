package retry

// RtyPolicy 用于分配和管理重试操作
type RtyPolicy interface {

	// CanRetry 判断当前状态是否可以重试
	CanRetry(ctx RtyContext) bool

	// Open 获取用于重试操作的资源
	Open(parent RtyContext) RtyContext

	// Close 关闭资源
	Close(ctx RtyContext)

	// RegisterError 每次重试尝试时，如果回调失败，就会把错误注册到ctx中
	RegisterError(ctx RtyContext, err error)
}
