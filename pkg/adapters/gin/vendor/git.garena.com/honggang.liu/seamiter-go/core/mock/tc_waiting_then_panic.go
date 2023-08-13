package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"time"
)

type waitingThenPanicTrafficShapingController struct {
	baseTrafficShapingController
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	//if !w.HeadersCheck(ctx) {
	//	return nil
	//}
	//if !w.ContextCheck(ctx) {
	//	return nil
	//}
	if nanosToWait := w.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", w.BoundRule(), "panic")
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := w.ArgsCheck(ctx)
	if item != nil {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", w.BoundRule(), item.ThenThrowMsg)
	}
	return nil
}
