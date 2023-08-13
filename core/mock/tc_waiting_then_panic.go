package mock

import (
	"github.com/liuhailove/seamiter-golang/core/base"
	"github.com/liuhailove/seamiter-golang/util"
	"time"
)

type waitingThenPanicTrafficShapingController struct {
	baseTrafficShapingController
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	if nanosToWait := w.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", w.BoundRule(), "panic")
}

func (w *waitingThenPanicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return w.DoInnerCheck(ctx)
}
