package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/util"
	"time"
)

type waitingThenMockTrafficShapingController struct {
	baseTrafficShapingController
}

func (m *waitingThenMockTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	if nanosToWait := m.r.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
		// Handle waiting action.
		util.Sleep(time.Duration(nanosToWait))
	}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), m.BoundRule().ThenReturnMockData)
}

func (m *waitingThenMockTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := m.ArgsCheck(ctx)
	if item != nil {
		if nanosToWait := item.ThenReturnWaitingTimeMs * time.Millisecond.Nanoseconds(); nanosToWait > 0 {
			// Handle waiting action.
			util.Sleep(time.Duration(nanosToWait))
		}
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", m.BoundRule(), item.ThenReturnMockData)
	}
	return nil
}
