package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"time"
)

type waitingTrafficShapingController struct {
	baseTrafficShapingController
}

func (w *waitingTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	if nanosToWait := w.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return nil
}

func (w *waitingTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := w.ArgsCheck(ctx)
	if item != nil {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
	}
	return nil
}
