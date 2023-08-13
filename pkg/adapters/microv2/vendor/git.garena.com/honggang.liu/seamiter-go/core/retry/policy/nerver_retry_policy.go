package policy

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/context"
)

// NeverRetryPolicy 允许第一次尝试，不允许之后的重试
type NeverRetryPolicy struct {
}

func (n *NeverRetryPolicy) CanRetry(ctx retry.RtyContext) bool {
	return (ctx.(*NeverRetryContext)).IsFinished()
}

func (n *NeverRetryPolicy) Open(parent retry.RtyContext) retry.RtyContext {
	var ntc = &NeverRetryContext{}
	ntc.Parent = parent
	return ntc
}

func (n *NeverRetryPolicy) Close(ctx retry.RtyContext) {
	// no-op
}

func (n NeverRetryPolicy) RegisterError(ctx retry.RtyContext, err error) {
	(ctx.(*NeverRetryContext)).setFinished()
	(ctx.(*NeverRetryContext)).RegisterError(err)
}

type NeverRetryContext struct {
	context.RtyContextSupport
	retry.SimpleAttributeAccessorSupport
	Finished bool
}

func (n *NeverRetryContext) IsFinished() bool {
	return n.Finished
}

func (n *NeverRetryContext) setFinished() {
	n.Finished = true
}

func NewNeverRetryPolicy() *NeverRetryPolicy {
	return &NeverRetryPolicy{}
}
