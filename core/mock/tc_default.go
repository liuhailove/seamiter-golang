package mock

import "github.com/liuhailove/seamiter-golang/core/base"

type defaultTrafficShapingController struct {
	baseTrafficShapingController
}

func (d *defaultTrafficShapingController) PerformCheckingFunc(_ *base.EntryContext) *base.TokenResult {
	return nil
}

func (d *defaultTrafficShapingController) PerformCheckingArgs(ctx *base.EntryContext) *base.TokenResult {
	return d.DoInnerCheck(ctx)
}
