package mock

import (
	"git.garena.com/honggang.liu/seamiter-go/core/base"
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
	tcs := getTrafficControllersFor(res)
	var r *base.TokenResult
	for _, tc := range tcs {
		if tc.BoundRule().Strategy == Func {
			r = tc.PerformCheckingFunc(ctx)
		} else {
			r = tc.PerformCheckingArgs(ctx)
		}
		if r == nil {
			continue
		}
		return r
	}
	return result
}
