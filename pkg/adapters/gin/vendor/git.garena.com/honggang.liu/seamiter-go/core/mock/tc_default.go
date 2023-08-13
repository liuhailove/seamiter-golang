package mock

import "git.garena.com/honggang.liu/seamiter-go/core/base"

type defaultTrafficShapingController struct {
	baseTrafficShapingController
}

func (d *defaultTrafficShapingController) PerformCheckingFunc(_ *base.EntryContext) *base.TokenResult {
	return nil
}

func (d *defaultTrafficShapingController) PerformCheckingArgs(_ *base.EntryContext) *base.TokenResult {
	return nil
}
