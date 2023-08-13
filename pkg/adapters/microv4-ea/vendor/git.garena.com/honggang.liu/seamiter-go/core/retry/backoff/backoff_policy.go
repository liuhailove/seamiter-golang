package backoff

import "git.garena.com/honggang.liu/seamiter-go/core/retry"

// BackOffPolicy 回退策略，控制两次重试之间的回退策略.
type BackOffPolicy interface {

	// Start 开启一个阻塞回退操作。当调用这个方法时，可以选择暂停，也可以立刻返回。
	// @param content 包含用于怎样处理的上下文信息
	// @return BackoffContext 一个明确的实现或者为空
	Start(content retry.RtyContext) BackoffContext

	// BackOff 更具具体的实现进行回退或者暂停
	BackOff(ctx BackoffContext)
}
