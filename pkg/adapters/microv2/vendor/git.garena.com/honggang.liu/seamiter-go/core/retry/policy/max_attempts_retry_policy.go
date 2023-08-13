package policy

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/context"
)

const (
	DefaultMaxRetryAttempts int32 = 3
)

type MaxAttemptsRetryPolicy struct {
	// 最大重试次数
	MaxAttempts int32
}

func (m *MaxAttemptsRetryPolicy) CanRetry(ctx retry.RtyContext) bool {
	return ctx.GetRetryCount() < m.MaxAttempts
}

func (m *MaxAttemptsRetryPolicy) Open(parent retry.RtyContext) retry.RtyContext {
	return &context.RtyContextSupport{Parent: parent}
}

func (m *MaxAttemptsRetryPolicy) Close(ctx retry.RtyContext) {
	// no-op
}

func (m *MaxAttemptsRetryPolicy) RegisterError(ctx retry.RtyContext, err error) {
	(ctx.(*context.RtyContextSupport)).RegisterError(err)
}

func NewMaxAttemptsRetryPolicy() *MaxAttemptsRetryPolicy {
	return &MaxAttemptsRetryPolicy{MaxAttempts: DefaultMaxRetryAttempts}
}

func NewMaxAttemptsRetryPolicyWithAttempts(maxAttempts int32) *MaxAttemptsRetryPolicy {
	return &MaxAttemptsRetryPolicy{MaxAttempts: maxAttempts}
}
