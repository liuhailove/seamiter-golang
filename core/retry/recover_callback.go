package retry

// RecoverCallback 在重试耗尽后的回调处理
type RecoverCallback interface {

	// Recover 重试恢复
	Recover(ctx RtyContext) error
}
