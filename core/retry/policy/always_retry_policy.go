package policy

import "github.com/liuhailove/seamiter-golang/core/retry"

// AlwaysRetryPolicy 一种无穷重试策略
type AlwaysRetryPolicy struct {
	NeverRetryPolicy
}

func (a *AlwaysRetryPolicy) CanRetry(ctx retry.RtyContext) bool {
	return true
}
