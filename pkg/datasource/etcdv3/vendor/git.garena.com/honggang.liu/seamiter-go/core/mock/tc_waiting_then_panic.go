package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"time"
)

type waitingThenPanicTrafficShapingController struct {
	baseTrafficShapingController
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingFunc(_ *base.EntryContext) *base.TokenResult {
	if nanosToWait := w.r.ThenReturnWaitingTimeMs * 1000; nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", w.BoundRule(), "panic")
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := w.ArgsCheck(ctx)
	if item != nil {
		util.Sleep(time.Duration(item.ThenReturnWaitingTimeMs * 1000))
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", w.BoundRule(), item.ThenThrowMsg)
	}
	return nil
}
