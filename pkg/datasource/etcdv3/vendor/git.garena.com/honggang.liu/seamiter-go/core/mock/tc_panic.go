package mock

import "git.garena.com/honggang.liu/seamiter-go/core/base"

type panicTrafficShapingController struct {
	baseTrafficShapingController
}

func (p *panicTrafficShapingController) PerformCheckingFunc(_ *base.EntryContext) *base.TokenResult {
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", p.BoundRule(), p.BoundRule().ThenThrowMsg)
}

func (p *panicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := p.ArgsCheck(ctx)
	if item != nil {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMock, "", p.BoundRule(), item.ThenThrowMsg)
	}
	return nil
}
