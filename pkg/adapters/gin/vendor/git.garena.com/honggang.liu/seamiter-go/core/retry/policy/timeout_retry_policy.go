package policy

import (
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/context"
	"git.garena.com/honggang.liu/seamiter-go/util"
)

var (
	// DefaultTimeout 默认超时时间
	DefaultTimeout int64 = 1000
)

// TimeoutRtyPolicy 只有在没有超时的情况下才进行重试
type TimeoutRtyPolicy struct {
	Timeout int64
}

func (t *TimeoutRtyPolicy) CanRetry(ctx retry.RtyContext) bool {
	return (ctx.(*TimeoutRtyContext)).IsAlive()
}

func (t *TimeoutRtyPolicy) Open(parent retry.RtyContext) retry.RtyContext {
	return NewTimeoutRtyContext(parent, t.Timeout)
}

func (t *TimeoutRtyPolicy) Close(ctx retry.RtyContext) {
	// no-op
}

func (t *TimeoutRtyPolicy) RegisterError(ctx retry.RtyContext, err error) {
	(ctx.(*TimeoutRtyContext)).RegisterError(err)
	// otherwise no-op - we only time out, otherwise retry everything...
}

type TimeoutRtyContext struct {
	Timeout int64
	Start   int64
	context.RtyContextSupport
	retry.SimpleAttributeAccessorSupport
}

func NewTimeoutRtyContext(parent retry.RtyContext, timeout int64) *TimeoutRtyContext {
	var inst = new(TimeoutRtyContext)
	inst.Parent = parent
	inst.Timeout = timeout
	inst.Start = int64(util.CurrentTimeMillis())
	return inst
}

func (t *TimeoutRtyContext) IsAlive() bool {
	return int64(util.CurrentTimeMillis())-t.Start <= t.Timeout
}
