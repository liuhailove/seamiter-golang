package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
	"git.garena.com/honggang.liu/seamiter-go/logging"
)

const (
	RuleCheckSlotOrder = 5000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	result := ctx.RuleCheckResult
	if len(res) == 0 {
		return result
	}
	if logging.DebugEnabled() {
		// mock命中，打印请求和mock数据
		logging.Debug("in mock check", "resourceName", res, "request", ctx.Input.Args)
	}
	tcs := getTrafficControllersFor(res)
	var r *base.TokenResult
	var cache = false
	for _, tc := range tcs {
		if !cache {
			cacheRequest(tc.BoundRule(), ctx)
			cache = true
		}
		if logging.DebugEnabled() {
			logging.Debug("in mock check", "metadata", ctx.Input.MetaData, "header", ctx.Input.Headers)
		}
		if !tc.MockCheck(ctx) {
			continue
		}
		if tc.BoundRule().Strategy == Func {
			r = tc.PerformCheckingFunc(ctx)
		} else {
			r = tc.PerformCheckingArgs(ctx)
		}
		if r == nil {
			continue
		}
		// mock命中，打印请求和mock数据
		logging.Info("mock hits", "request", ctx.Input.Args, "mockData", r.String())
		return r
	}
	return result
}
