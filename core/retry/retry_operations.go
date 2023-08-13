package retry

// RtyOperations 定义一组重试操作，这组操作根据配置的重试行为进行重试
type RtyOperations interface {

	//Execute 根据提供的重试语意执行callback操作
	Execute(callback RtyCallback) (interface{}, error)

	// ExecuteWithRecover 重试回调，在重试耗尽后，使用一个回退方法去执行RecoverCallback
	ExecuteWithRecover(callback RtyCallback, recoverCallback RecoverCallback) (interface{}, error)

	// ExecuteWithState 有状态的重试
	ExecuteWithState(callback RtyCallback, state RtyState) (interface{}, error)

	// ExecuteWithRecoverAndState 有状态的重试回调，在重试耗尽后，使用一个回退方法去执行RecoverCallback
	ExecuteWithRecoverAndState(callback RtyCallback, recoverCallback RecoverCallback, state RtyState) (interface{}, error)
}
