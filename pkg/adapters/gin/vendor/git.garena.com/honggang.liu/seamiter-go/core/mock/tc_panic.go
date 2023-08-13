package mock

import "git.garena.com/honggang.liu/seamiter-go/core/base"

type panicTrafficShapingController struct {
	baseTrafficShapingController
}

func (p *panicTrafficShapingController) PerformCheckingFunc(ctx *base.EntryContext) *base.TokenResult {
	//if !p.HeadersCheck(ctx) {
	//	return nil
	//}
	//if !p.ContextCheck(ctx) {
	//	return nil
	//}
	return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", p.BoundRule(), p.BoundRule().ThenThrowMsg)
}

func (p *panicTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	item := p.ArgsCheck(ctx)
	if item != nil {
		return base.NewTokenResultBlockedWithCause(base.BlockTypeMockError, "", p.BoundRule(), item.ThenThrowMsg)
	}
	return nil
}
